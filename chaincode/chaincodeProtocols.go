package chaincode

import (
	"time"
)

const (
	InstallChaincodeTimeout     = 5 * time.Second
	InstantiateChaincodeTimeout = 5 * time.Second
	WaitTxTimeout               = 20 * time.Second
)
const (
	acceptAllPolicy = "OutOf(0, 'None.member')"
)

type InstallChaincodeRequest struct {
	Org       string
	CcTarPath string
	CcPath    string
	CcName    string
	CcVersion string
	PeerNodes []*ServiceNode
}
type InstantiateChaincodeRequest struct {
	Org          string
	ChannelName  string
	CcName       string
	CcVersion    string
	Policy       string
	Args         [][]byte
	PeerNodes    []*ServiceNode
	OrdererNodes []*ServiceNode
}

type InvokeRequest struct {
	Org          string
	ChannelName  string
	CcName       string
	Args         [][]byte
	PeerNodes    []*ServiceNode
	OrdererNodes []*ServiceNode
}

type ServiceNode struct {
	ID               string
	Endpoint         string
	ExternalEndpoint string
	Public           bool
}
