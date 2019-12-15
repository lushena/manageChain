package sdk

import (
	"context"
	"time"

	pp "github.com/hyperledger/fabric/protos/peer"
	putils "github.com/hyperledger/fabric/protos/utils"
	"google.golang.org/grpc"
)

const (
	defaultTimeout = time.Second * 3
)

// EndorserClient ...
type EndorserClient struct {
	pp.EndorserClient
	conn *grpc.ClientConn
}

// Close ...
func (ec *EndorserClient) Close() error {
	return ec.conn.Close()
}

// GetEndorsers ...
func (client *Client) GetEndorsers(endpoints []*Endpoint) ([]*EndorserClient, error) {
	var clients []*EndorserClient
	for _, ep := range endpoints {
		ec, err := newEndorserClient(ep)
		if err != nil {
			logger.Error("Error creating endorser client", err)
			for _, c := range clients {
				c.Close()
			}
			return nil, err
		}
		clients = append(clients, ec)
	}
	return clients, nil
}

// Endorse ...
// Only support chaincode written in golang
func (client *Client) Endorse(chainID string, chaincode string, args [][]byte, transient map[string][]byte, endorsers []*Endpoint) (string, *pp.Proposal, []*pp.ProposalResponse, error) {

	creator, err := client.signer.Serialize()
	if err != nil {
		logger.Error("Error serializing identity for", client.signer.GetIdentifier())
		return "", nil, nil, err
	}

	txID, prop, err := CreateChaincodeProposal(chainID, chaincode, args, transient, creator)
	if err != nil {
		logger.Error("Error creating CreateChaincodeProposalBytes", err)
		return "", nil, nil, err
	}

	propBytes, err := putils.GetBytesProposal(prop)
	if err != nil {
		logger.Error("Error getting bytes of proposal", err)
		return "", nil, nil, err
	}

	signature, err := client.signer.Sign(propBytes)
	if err != nil {
		logger.Error("Error signning proposal", err)
		return "", nil, nil, err
	}

	ecList, err := client.GetEndorsers(endorsers)
	if err != nil {
		logger.Error("Error getting endorsers", err)
		return "", nil, nil, err
	}
	defer func() {
		for _, ec := range ecList {
			ec.Close()
		}
	}()

	responses, err := endorse(propBytes, signature, ecList)
	return txID, prop, responses, err
}

func endorse(proposalBytes []byte, signature []byte, endorsers []*EndorserClient) ([]*pp.ProposalResponse, error) {
	signedProposal := &pp.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}
	var responses []*pp.ProposalResponse
	for _, client := range endorsers {
		resp, err := client.ProcessProposal(context.Background(), signedProposal)
		if err != nil {
			logger.Error("Error processing proposal", err)
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil

}

// Endorse ...
func Endorse(proposalBytes []byte, signature []byte, endorsers []*Endpoint) ([]*pp.ProposalResponse, error) {
	var clients []*EndorserClient
	for _, endorser := range endorsers {
		client, err := newEndorserClient(endorser)
		if err != nil {
			logger.Error("Error creating endorserClient", err)
			return nil, err
		}

		defer client.Close()
		clients = append(clients, client)
	}
	return endorse(proposalBytes, signature, clients)

}

// EndorseToBytes ...
// Return: proposalResponseBytes, responsePayloadBytes, error
func EndorseToBytes(resps []*pp.ProposalResponse) ([][]byte, [][]byte, error) {
	var responses [][]byte
	var payloads [][]byte
	for _, resp := range resps {
		bytes, err := putils.Marshal(resp)
		if err != nil {
			logger.Error("Error marshaling ProposalResponse", err)
			return nil, nil, err
		}
		responses = append(responses, bytes)
		payloads = append(payloads, getResponsePayloadFromProposalResponse(resp))
	}
	return responses, payloads, nil
}

func newEndorserClient(endorser *Endpoint) (*EndorserClient, error) {
	conn, err := createConnection(endorser)
	if err != nil {
		logger.Error("Error creating connection", err)
		return nil, err
	}
	return &EndorserClient{
		EndorserClient: pp.NewEndorserClient(conn),
		conn:           conn,
	}, nil

}

func getResponsePayloadFromProposalResponse(pr *pp.ProposalResponse) []byte {
	return pr.Response.Payload
}
