package controllers

import (
	"encoding/json"
	"manageChain/chaincode"
	"manageChain/channel"
	"path"
	"time"

	"github.com/astaxie/beego"
	logger "github.com/astaxie/beego/logs"
	"github.com/hyperledger/fabric/sdk"
)

type ChaincodeController struct {
	BaseController
}

func (c *ChaincodeController) InstallChaincode() error {
	logger.Info("start Install Chaincode")

	icq := &chaincode.InstallChaincodeRequest{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, icq)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}

	org := icq.Org
	ccTarPath := icq.CcTarPath
	ccPath := icq.CcPath
	ccName := icq.CcName
	ccVersion := icq.CcVersion

	newchaincode, err := newChaincode(org, ccTarPath, ccPath, ccName, ccVersion)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}

	orgCA := newchaincode.GetOrgCA()
	endorsers := serviceNodesToEndpointList(icq.PeerNodes, chaincode.InstallChaincodeTimeout, orgCA.TLSCACert())

	err = newchaincode.InstallChaincode(endorsers)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}

	c.ReturnOKMsg("OK")
	logger.Info("successfully Install Chaincode")
	return nil
}

func newChaincode(org string, ccTarPath string, ccPath string, ccName string, ccVersion string) (*chaincode.Chaincode, error) {
	mspDir := beego.AppConfig.String("MSPDir")
	gm, _ := beego.AppConfig.Bool("GM")

	orgCA, err := channel.GetCA(path.Join(mspDir, org), org)
	if err != nil {
		logger.Error("Error getting peer ca", err)
		return nil, err
	}

	return chaincode.NewChaincode(org, ccTarPath, ccPath, ccName, ccVersion, orgCA, gm)
}

func (c *ChaincodeController) InstantiateChaincode() error {
	logger.Info("start Instantiate Chaincode")

	icq := &chaincode.InstantiateChaincodeRequest{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, icq)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}
	org := icq.Org
	ccTarPath := ""
	ccPath := ""
	ccName := icq.CcName
	ccVersion := icq.CcVersion

	newchaincode, err := newChaincode(org, ccTarPath, ccPath, ccName, ccVersion)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}

	channelName := icq.ChannelName
	policy := icq.Policy
	args := icq.Args
	orgCA := newchaincode.GetOrgCA()
	endorsers := serviceNodesToEndpointList(icq.PeerNodes, chaincode.InstantiateChaincodeTimeout, orgCA.TLSCACert())
	casters := serviceNodesToEndpointList(icq.OrdererNodes, chaincode.InstantiateChaincodeTimeout, orgCA.TLSCACert())
	err = newchaincode.InstantiateChaincode(endorsers, casters, channelName, policy, args)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}

	c.ReturnOKMsg("OK")
	logger.Info("successfully Instantiate Chaincode")
	return nil
}

func (c *ChaincodeController) Invoke() error {
	logger.Info("start Invoke Chaincode")

	iq := &chaincode.InvokeRequest{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, iq)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}
	org := iq.Org
	ccTarPath := ""
	ccPath := ""
	ccName := iq.CcName
	ccVersion := ""

	newchaincode, err := newChaincode(org, ccTarPath, ccPath, ccName, ccVersion)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}

	channelName := iq.ChannelName
	args := iq.Args

	orgCA := newchaincode.GetOrgCA()
	endorsers := serviceNodesToEndpointList(iq.PeerNodes, chaincode.InstantiateChaincodeTimeout, orgCA.TLSCACert())
	casters := serviceNodesToEndpointList(iq.OrdererNodes, chaincode.InstantiateChaincodeTimeout, orgCA.TLSCACert())
	err = newchaincode.Invoke(channelName, endorsers, casters, args)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}

	c.ReturnOKMsg("OK")
	logger.Info("successfully Invoke Chaincode")
	return nil
}

func serviceNodesToEndpointList(serviceNodes []*chaincode.ServiceNode, timeout time.Duration, cert []byte) []*sdk.Endpoint {
	var endpoints []*sdk.Endpoint
	for _, sn := range serviceNodes {
		endpoints = append(endpoints, &sdk.Endpoint{
			Address:  sn.Endpoint,
			Override: "", // pay attention
			TLS:      cert,
			Timeout:  timeout,
		})
	}
	return endpoints
}
