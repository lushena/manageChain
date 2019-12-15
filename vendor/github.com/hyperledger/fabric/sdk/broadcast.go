package sdk

import (
	"context"

	comm "github.com/hyperledger/fabric/protos/common"
	ab "github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// BroadcastClient ...
type BroadcastClient struct {
	ab.AtomicBroadcast_BroadcastClient
	conn *grpc.ClientConn
}

// Close ...
func (bc *BroadcastClient) Close() {
	bc.CloseSend()
	bc.conn.Close()
}

func (bc *BroadcastClient) ack() error {
	msg, err := bc.Recv()
	if err != nil {
		return err
	}
	if msg.Status != comm.Status_SUCCESS {
		return errors.Errorf("got unexpected status: %v -- %s", msg.Status, msg.Info)
	}
	return nil
}

// Roundtrip ...
// Send and wait for ack
func (bc *BroadcastClient) Roundtrip(env *comm.Envelope) error {
	if err := bc.Send(env); err != nil {
		logger.Error("Error sending envelope", err)
		return err
	}
	return bc.ack()
}

func newBroadcastClient(caster *Endpoint) (*BroadcastClient, error) {
	conn, err := createConnection(caster)
	if err != nil {
		logger.Error("Error creating connection", err)
		return nil, err
	}

	bc, err := ab.NewAtomicBroadcastClient(conn).Broadcast(context.TODO())
	if err != nil {
		logger.Error("Error creating AtomicBroadcastClient", err)
		conn.Close()
		return nil, err
	}

	return &BroadcastClient{
		AtomicBroadcast_BroadcastClient: bc,
		conn: conn,
	}, nil

}

func broadcast(payload []byte, signature []byte, bc *BroadcastClient) error {
	env := &comm.Envelope{
		Payload:   payload,
		Signature: signature,
	}
	err := bc.Roundtrip(env)
	return err
}

// Broadcast ...
func Broadcast(payload []byte, signature []byte, caster *Endpoint) error {
	bc, err := newBroadcastClient(caster)
	if err != nil {
		logger.Error("Error creating BroadcastClient", err)
		return err
	}
	defer bc.Close()
	return broadcast(payload, signature, bc)
}

// Broadcast ...
func (client *Client) Broadcast(proposal *peer.Proposal, responses []*peer.ProposalResponse, caster *Endpoint) error {
	payload, err := CreateChaincodeEnvelopeBytes(proposal, responses)
	if err != nil {
		logger.Error("Error creating ChaincodeEnvelopeBytes", err)
		return err

	}
	signature, err := client.signer.Sign(payload)
	if err != nil {
		logger.Error("Error signning payload", err)
		return err
	}
	bc, err := client.GetBroadcaster(caster)
	if err != nil {
		logger.Error("Error getting broadcaster", err)
		return err
	}
	defer bc.Close()

	return broadcast(payload, signature, bc)
}

// GetBroadcaster ...
func (client *Client) GetBroadcaster(caster *Endpoint) (*BroadcastClient, error) {
	bc, err := newBroadcastClient(caster)
	if err != nil {
		logger.Error("Error creating BroadcastClient", err)
		return nil, err
	}
	return bc, nil

}
