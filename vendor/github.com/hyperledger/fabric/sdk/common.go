package sdk

import (
	"time"

	"github.com/hyperledger/fabric/core/comm"
	"google.golang.org/grpc"
)

// Endpoint ...
type Endpoint struct {
	Address  string
	Override string
	TLS      []byte
	Timeout  time.Duration
}

func createConnection(endpoint *Endpoint) (*grpc.ClientConn, error) {
	clientConfig := comm.ClientConfig{}
	timeout := endpoint.Timeout
	if timeout == time.Duration(0) {
		timeout = defaultTimeout
	}
	clientConfig.Timeout = timeout
	if endpoint.TLS != nil {
		secOpts := &comm.SecureOptions{
			UseTLS: true,
		}
		secOpts.ServerRootCAs = [][]byte{endpoint.TLS}
		clientConfig.SecOpts = secOpts
	}

	gClient, err := comm.NewGRPCClient(clientConfig)
	if err != nil {
		logger.Error("Failed to create PeerClient from config", err)
		return nil, err
	}
	return gClient.NewConnection(endpoint.Address, endpoint.Override)
}
