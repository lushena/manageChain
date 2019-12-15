package channel

import (
	"encoding/json"
	"errors"
	"time"

	pp "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/sdk"
)

func AnchorPeers(peerNodes []*ServiceNode) (peers []string) {
	for _, info := range peerNodes {
		if info.Public {
			peers = append(peers, "grpcs://"+info.ExternalEndpoint)
		}
	}
	return
}

func Orderers(ordererNodes []*ServiceNode) (orderers []string) {
	for _, info := range ordererNodes {
		orderers = append(orderers, info.ExternalEndpoint)
	}
	return
}

func serviceNodesToEndpointList(serviceNodes []*ServiceNode, timeout time.Duration, cert []byte) []*sdk.Endpoint {
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

	valid, err := client.WaitTx(chainID, txID, endorder, waitTxTimeout)
	if err != nil {
		logger.Error("Error waiting transaction", err)
		return err
	}

	if !valid {
		return errors.New("invoke is not valid, please try again")
	}

	return nil

}

func query(client *sdk.Client, chainID string, chaincode string, args [][]byte, peers []*sdk.Endpoint) ([]byte, error) {
	_, _, resps, _, err := endorseOneOfList(client, chainID, chaincode, args, nil, peers)
	if err != nil {
		logger.Error("Error querying", err)
		return nil, err
	}
	return resps[0].Response.Payload, nil
}

func (bl *bytesList) Serialize() ([]byte, error) {
	return json.Marshal(bl)
}

func (bl *bytesList) Deserialize(data []byte) error {
	b := bytesList{}
	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}
	*bl = b
	return nil
}

func (bl *bytesList) Data(index int) []byte {
	return [][]byte(*bl)[index]
}
