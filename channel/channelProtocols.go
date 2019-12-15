package channel

import (
	"time"

	"github.com/hyperledger/fabric/common/tools/configtxgen/encoder"
	"github.com/hyperledger/fabric/sdk"
)

const (
	CreateChannelTimeout = 5 * time.Second
	EndorseTimeout       = 5 * time.Second
	waitTxTimeout        = 20 * time.Second
)
const (
	addOrgInfo    = "AddOrgInfo"
	getOrgInfo    = "GetOrgInfo"
	getAllOrgInfo = "GetAllOrgInfo"
	getAllOrgname = "GetAllOrgname"
	updateOrgInfo = "UpdateOrgInfo"

	startInvitation   = "StartInvitation"
	getInvitation     = "GetInvitation"
	getAllInvitation  = "GetAllInvitation"
	confirmInvitation = "ConfirmInvitation"

	signInvitation          = "SignInvitation"
	getInvitationSignStatus = "GetInvitationSignStatus"

	updateChainOrgInfo   = "UpdateChainOrgInfo"
	getChainOrgInfo      = "GetChainOrgInfo"
	getAllOrgnameOfChain = "GetAllOrgnameOfChain"
)
const (
	PublicChainID = "publicchain"

	// PublicCCName ...
	PublicCCName      = "publicchaincode"
	DefaultConsortium = sdk.DefaultConsortium
)

const (
	initState    = "init"
	confirmState = "confirm"

	acceptState = "Accept"
	rejectState = "Reject"
)

type ServiceNode struct {
	ID               string
	Endpoint         string
	ExternalEndpoint string
	Public           bool
}

type OrgInfo struct {
	OrgName      string
	MspID        string
	OrgMSP       string
	OrgCA        *sdk.CA
	Client       *sdk.Client
	PeerNodes    []*ServiceNode
	OrdererNodes []*ServiceNode
}
type NewCreateChannelRequest struct {
	Orgs        []*OrgInfo
	ChannelName string
}

type JoinChannelRequest struct {
	Orgs        []*OrgInfo
	ChannelName string
}

type IdentityRequest struct {
	Orgs []*OrgInfo
}
type AddOrgRequest struct {
	Orgs        []*OrgInfo
	Identity    []byte
	ChannelName string
}

type DeleteOrgRequest struct {
	Orgs        []*OrgInfo
	DelOrg      string
	DelOrderers []string
	ChannelName string
}

type GenCryptoRequest struct {
	Orgs []*OrgInfo
}

type GenGenesisBlockRequest struct {
	Orgs   []*OrgInfo
	Kafkas []string
}

type InviteCodeRequest struct {
	Orgs        []*OrgInfo
	ChannelName string
}

type OrgJoinChannelRequest struct {
	Orgs       []*OrgInfo
	InviteCode []byte
}
type IdentityCode struct {
	Org          string
	OrgMSP       []byte
	Orderers     []string
	Anchors      []string
	ChainOrgInfo *ChainOrgInfo
}

type ChainOrgInfo struct {
	Peers       []*sdk.Endpoint
	Orderers    []*sdk.Endpoint
	OrgName     string
	ChannelName string
}

type bytesList [][]byte
type ccInvitation struct {
	Inviter    string `json:"inviter"`
	Invitee    string `json:"invitee"`
	Status     string `json:"status"`
	InviteTime int64  `json:"inviteTime"`
	RawData    string `json:"rawdata"`
}

type ccInvitationSignStatus struct {
	Inviter   string `json:"inviter"`
	Invitee   string `json:"invitee"`
	Signer    string `json:"signer"`
	Signature string `json:"signature"`
	Accepted  string `json:"accepted"`
	SignTime  int64  `json:"signTime"`
}
type InviteCode struct {
	ChannelGenesisBlock []byte
}

const (
	DefaultMSPType               = "bccsp"
	DefaultBatchTimeout          = 5 * time.Second
	DefaultMaxMessageCount       = 10
	DefaultAbsoluteMaxBytes      = 100 * 1024 * 1024
	DefaultPreferredMaxBytes     = 512 * 1024
	DefaultChannelCapability     = "V1_1"
	DefaultOrdererCapability     = "V1_1"
	DefaultApplicationCapability = "V1_2"
	DefaultPolicyType            = encoder.ImplicitMetaPolicyType
)
