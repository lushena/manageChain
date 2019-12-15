package sdk

import (
	"context"
	"fmt"

	dis "github.com/hyperledger/fabric/discovery/client"
	pd "github.com/hyperledger/fabric/protos/discovery"
	"google.golang.org/grpc"
)

// MSPConfig ...
type MSPConfig struct {
	RootCert  []byte
	AdminCert []byte
	TLSCert   []byte
	Nodes     map[string]Node
}

// Node channel relative
type Node struct {
	Endpoint     string
	LedgerHeight uint64
	Chaincodes   []Chaincode
}

// Chaincode ...
type Chaincode struct {
	Name    string
	Version string
}

// DiscoveryChannel ...
func (client *Client) DiscoveryChannel(chainID string, peer *Endpoint) (map[string]*MSPConfig, error) {
	identity, err := client.signer.Serialize()
	if err != nil {
		logger.Error("Error getting client identity", err)
		return nil, err
	}
	dialer := func() (*grpc.ClientConn, error) {
		return createConnection(peer)
	}
	dc := dis.NewClient(dialer, client.signer.Sign)
	ctx := context.TODO()
	req := dis.NewRequest().OfChannel(chainID).AddConfigQuery().AddPeersQuery()
	auth := &pd.AuthInfo{
		ClientIdentity: identity,
	}
	resp, err := dc.Send(ctx, req, auth)
	if err != nil {
		logger.Error("Error sending discovery request", err)
		return nil, err
	}

	ret := make(map[string]*MSPConfig)

	peers, err := resp.ForChannel(chainID).Peers()
	if err != nil {
		logger.Error("Error getting peers info", err)
		return nil, err
	}
	for _, peer := range peers {
		if ret[peer.MSPID] == nil {
			ret[peer.MSPID] = &MSPConfig{
				Nodes: make(map[string]Node),
			}
		}
		mspConf := ret[peer.MSPID]

		if aliveMsg := peer.AliveMessage; aliveMsg != nil {
			if gossipMsp := aliveMsg.GossipMessage; gossipMsp != nil {
				if alive := gossipMsp.GetAliveMsg(); alive != nil {
					if mem := alive.Membership; mem != nil {
						node := Node{Endpoint: mem.Endpoint}
						mspConf.Nodes[mem.Endpoint] = node
						// set ledger height and chaincodes
						if stateMsg := peer.StateInfoMessage; stateMsg != nil {
							if gossipMsg := stateMsg.GossipMessage; gossipMsg != nil {
								if info := gossipMsg.GetStateInfo(); info != nil {
									node.LedgerHeight, _ = info.LedgerHeight()
									if prop := info.Properties; prop != nil {
										for _, cc := range prop.Chaincodes {
											node.Chaincodes = append(node.Chaincodes, Chaincode{
												Name:    cc.Name,
												Version: cc.Version,
											})
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	conf, err := resp.ForChannel(chainID).Config()
	if err != nil {
		logger.Error("Error getting config", err)
		return nil, err
	}
	for mspid, value := range conf.Msps {
		if ret[mspid] == nil {
			ret[mspid] = &MSPConfig{
				Nodes: make(map[string]Node),
			}
		}
		mspConf := ret[mspid]
		if len(value.Admins) > 0 {
			mspConf.AdminCert = value.Admins[0]
		}
		if len(value.RootCerts) > 0 {
			mspConf.RootCert = value.RootCerts[0]
		}
		if len(value.TlsRootCerts) > 0 {
			mspConf.TLSCert = value.TlsRootCerts[0]
		}
	}

	for mspid, value := range conf.Orderers {
		if ret[mspid] == nil {
			ret[mspid] = &MSPConfig{
				Nodes: make(map[string]Node),
			}
		}
		mspConf := ret[mspid]
		for _, ep := range value.Endpoint {
			addr := fmt.Sprintf("%s:%d", ep.Host, ep.Port)
			mspConf.Nodes[addr] = Node{Endpoint: addr}
		}
	}

	return ret, nil
}
