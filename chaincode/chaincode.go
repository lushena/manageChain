package chaincode

import (
	"errors"
	logs "gglogs"
	"io/ioutil"

	pp "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/sdk"
)

var logger *logs.BeeLogger

func init() {
	logger = logs.GetBeeLogger()
}

type Chaincode struct {
	org       string
	orgMSP    string
	ccTarPath string
	ccPath    string
	ccName    string
	ccVersion string
	orgCA     *sdk.CA
	client    *sdk.Client
}

func NewChaincode(orgMSP string, ccTarPath string, ccPath string, ccName string, ccVersion string, orgCA *sdk.CA, gm bool) (*Chaincode, error) {
	client, err := sdk.NewClient(orgCA.AdminCommonName(), orgMSP, orgCA.AdminMSPDir(), gm)
	if err != nil {
		logger.Error("Error creating client for org", err)
		return nil, err
	}

	return &Chaincode{
		orgMSP:    orgMSP,
		orgCA:     orgCA,
		ccTarPath: ccTarPath,
		ccPath:    ccPath,
		ccName:    ccName,
		ccVersion: ccVersion,
		client:    client,
	}, nil
}

func (cc *Chaincode) InstallChaincode(endorsers []*sdk.Endpoint) error {
	ccTarPath := cc.ccTarPath
	ccPath := cc.ccPath
	ccName := cc.ccName
	ccVersion := cc.ccVersion
	if ccTarPath == "" {
		return errors.New("chaincode package path should not be empty")
	}

	err := installChaincode(cc.client, endorsers, ccTarPath, ccPath, ccName, ccVersion)
	if err != nil {
		logger.Error("Error installing  chaincode", err)
		return err
	}
	logger.Info("Successfully installing  chaincode")
	return nil
}

func installChaincode(client *sdk.Client, endorsers []*sdk.Endpoint, tarPath string, ccPath string, name string, version string) error {
	data, err := ioutil.ReadFile(tarPath)
	if err != nil {
		logger.Error("Error reading file", err)
		return err
	}
	return client.InstallChaincode(name, version, ccPath, data, endorsers)
}

func (cc *Chaincode) GetOrgCA() *sdk.CA {
	return cc.orgCA
}

func (cc *Chaincode) InstantiateChaincode(endorsers []*sdk.Endpoint, casters []*sdk.Endpoint, channelName string, policy string, args [][]byte) error {
	ccName := cc.ccName
	ccVersion := cc.ccVersion

	err := instantiateChaincode(cc.client, channelName, ccName, ccVersion, endorsers, casters, args, policy)
	if err != nil {
		logger.Error("Error Instantiate chaincode", err)
		return err
	}
	logger.Info("Successfully Instantiate  chaincode")
	return nil
}

func instantiateChaincode(client *sdk.Client, chainID string, ccName string, version string, endorsers []*sdk.Endpoint, casters []*sdk.Endpoint, args [][]byte, policy string) error {
	logger.Info("policy:%s\n\n", policy)
	for _, endorser := range endorsers {
		if err := client.InstantiateChaincode(chainID, ccName, version, args, policy, nil, endorser, casters); err != nil {
			logger.Error("Error Instantiate chaincode", err)
			continue
		}
		return nil
	}
	return errors.New("failed Instantiate chaincode")
}

func (cc *Chaincode) Invoke(channelName string, peers []*sdk.Endpoint, orderers []*sdk.Endpoint, args [][]byte) error {
	client := cc.client
	ccName := cc.ccName

	err := invoke(client, channelName, ccName, args, peers, orderers)
	if err != nil {
		logger.Error("Error invoke chaincode", err)
		return err
	}
	logger.Info("Successfully invoke  chaincode")

	return nil
}

func invoke(client *sdk.Client, chainID string, chaincode string, args [][]byte, peers []*sdk.Endpoint, orderers []*sdk.Endpoint) error {
	txID, prop, resps, endorder, err := endorseOneOfList(client, chainID, chaincode, args, nil, peers)
	if err != nil {
		logger.Error("Error endorsing", err)
		return err
	}

	err = broadcastOneOfList(client, prop, resps, orderers)
	if err != nil {
		logger.Error("Error broadcasing", err)
		return err
	}

	valid, err := client.WaitTx(chainID, txID, endorder, WaitTxTimeout)
	if err != nil {
		logger.Error("Error waiting transaction", err)
		return err
	}

	if !valid {
		return errors.New("invoke is not valid, please try again")
	}

	return nil
}

func endorseOneOfList(client *sdk.Client, chainID string, chaincode string, args [][]byte, transient map[string][]byte, peerEndpoints []*sdk.Endpoint) (txID string, prop *pp.Proposal, resps []*pp.ProposalResponse, endorser *sdk.Endpoint, err error) {
	for _, peer := range peerEndpoints {
		txID, prop, resps, err = client.Endorse(chainID, chaincode, args, transient, []*sdk.Endpoint{peer})
		if err == nil {
			endorser = peer
			break
		}
		logger.Error("Error endorsing", err)
	}
	if err != nil {
		return "", nil, nil, nil, errors.New("failed proposing through all peers")
	}
	return
}

func broadcastOneOfList(client *sdk.Client, prop *pp.Proposal, resps []*pp.ProposalResponse, ordererEndpoints []*sdk.Endpoint) (err error) {
	for _, orderer := range ordererEndpoints {
		err = client.Broadcast(prop, resps, orderer)
		if err == nil {
			break
		}
		logger.Error("Error broadcasting", err)
	}
	if err != nil {
		return errors.New("failed broadcasting through all orderers")
	}
	return
}
