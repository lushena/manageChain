package sdk

import (
	"context"
	"errors"
	"io/ioutil"
	"net/url"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/common/channelconfig"
	"github.com/hyperledger/fabric/common/tools/configtxgen/encoder"
	"github.com/hyperledger/fabric/common/tools/configtxgen/localconfig"
	"github.com/hyperledger/fabric/common/tools/configtxlator/update"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/scc/cscc"
	"github.com/hyperledger/fabric/msp"
	cb "github.com/hyperledger/fabric/protos/common"
	ab "github.com/hyperledger/fabric/protos/orderer"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
)

const (
	defaultMSPType               = "bccsp"
	defaultBatchTimeout          = 5 * time.Second
	defaultMaxMessageCount       = 10
	defaultAbsoluteMaxBytes      = 100 * 1024 * 1024
	defaultPreferredMaxBytes     = 512 * 1024
	defaultChannelCapability     = "V1_1"
	defaultOrdererCapability     = "V1_1"
	defaultApplicationCapability = "V1_2"
	defaultPolicyType            = encoder.ImplicitMetaPolicyType
)

const (
	// DefaultSystemChainID is the default name of system chain
	DefaultSystemChainID = "systemchain"
)

const (
	// DefaultConsortium is the name of the default consortim
	DefaultConsortium = "defaultConsortium"
)

var acceptAllPolicy = &localconfig.Policy{
	Type: encoder.SignaturePolicyType,
	Rule: "OutOf(0, 'None.member')",
}

// ImplicitMetaPolicy ...
type ImplicitMetaPolicy string

// policy
const (
	PolicyAnyAdmins      ImplicitMetaPolicy = "ANY Admins"
	PolicyMajorityAdmins ImplicitMetaPolicy = "MAJORITY Admins"
	PolicyAllAdmins      ImplicitMetaPolicy = "ALL Admins"

	PolicyAnyWriters      ImplicitMetaPolicy = "ANY Writers"
	PolicyAllWriters      ImplicitMetaPolicy = "ALL Writers"
	PolicyMajorityWriters ImplicitMetaPolicy = "MAJORITY Writers"

	PolicyAnyReaders      ImplicitMetaPolicy = "ANY Readers"
	PolicyAllReaders      ImplicitMetaPolicy = "ALL Readers"
	PolicyMajorityReaders ImplicitMetaPolicy = "MAJORITY Readers"
)

// GenesisConfig ...
type GenesisConfig struct {
	ChainID                 string
	OrdererType             string
	Addresses               []string
	BatchTimeout            time.Duration
	KafkaBrokers            []string
	MaxMessageCount         uint32
	AbsoluteMaxBytes        uint32
	PreferredMaxBytes       uint32
	MaxChannels             uint64
	OrdererOrganizations    []*Organization
	ConsortiumOrganizations []*Organization
	ConsortiumName          string
	AdminsPolicy            ImplicitMetaPolicy
	WritersPolicy           ImplicitMetaPolicy
	ReadersPolicy           ImplicitMetaPolicy
}

// ChannelConfig ...
type ChannelConfig struct {
	ChainID       string
	Consortium    string
	Organizations []*Organization
	AdminsPolicy  ImplicitMetaPolicy
	WritersPolicy ImplicitMetaPolicy
	ReadersPolicy ImplicitMetaPolicy
}

// Organization ...
// Only support fabric msp type
// AnchorPeers: ["grpcs://192.168.9.11:3432", ...]
type Organization struct {
	Name        string
	ID          string
	MSPDir      string
	AnchorPeers []string
}

// SignChannelConfigUpdate ...
func (client *Client) SignChannelConfigUpdate(update []byte) ([]byte, []byte, error) {
	creator, err := client.signer.Serialize()
	if err != nil {
		logger.Error("Error serializing signer", err)
		return nil, nil, err
	}
	sigHeader, err := newSignatureHeaderWithCreator(creator)
	if err != nil {
		logger.Error("Error creating signature header", err)
		return nil, nil, err
	}
	signatureHeader := utils.MarshalOrPanic(sigHeader)

	toSignBytes := util.ConcatenateBytes(signatureHeader, update)

	signedSigHeader, err := client.signer.Sign(toSignBytes)
	if err != nil {
		logger.Error("Error signning sigHeader", err)
		return nil, nil, err
	}
	return signatureHeader, signedSigHeader, nil
}

// GetAddOrgChannelConfigUpdate ...
func (client *Client) GetAddOrgChannelConfigUpdate(chainID string, block *cb.Block, newOrdererOrgs []*Organization, newApplicationOrgs []*Organization, newConsortiumOrgs map[string][]*Organization, orderers []string) ([]byte, error) {
	tx, err := configUpdate(chainID, block, newOrdererOrgs, newApplicationOrgs, newConsortiumOrgs, orderers)
	if err != nil {
		logger.Error("Error computing update", err)
		return nil, err
	}
	return utils.Marshal(tx)
}

func (client *Client) GetDelOrgChannelConfigUpdate(chainID string, block *cb.Block, delOrg string, delOrderers []string) ([]byte, error) {
	tx, err := delOrgConfigUpdate(chainID, block, delOrg, delOrderers)
	if err != nil {
		logger.Error("Error", err)
		return nil, err
	}
	return utils.Marshal(tx)
}

// UpdateChannel ...
func (client *Client) UpdateChannel(chainID string, block *cb.Block, newOrdererOrgs []*Organization, newApplicationOrgs []*Organization, newConsortiumOrgs map[string][]*Organization, orderers []string, caster *Endpoint) error {
	return updateChannel(chainID, block, newOrdererOrgs, newApplicationOrgs, newConsortiumOrgs, orderers, caster, client.signer)
}

// UpdateChannelByConfigUpdate ...
func (client *Client) UpdateChannelByConfigUpdate(chainID string, configUpdate []byte, sigs []*cb.ConfigSignature, caster *Endpoint) error {
	creator, err := client.signer.Serialize()
	if err != nil {
		logger.Error("Error serializing signer", err)
		return err
	}

	envelopeBytes, err := CreateChannelEnvelopeBytes(chainID, creator, configUpdate, sigs)
	if err != nil {
		logger.Error("Error creating newChannelEnvelope payload", err)
		return err
	}

	signature, err := client.signer.Sign(envelopeBytes)
	if err != nil {
		logger.Error("Error signning payload", err)
		return err
	}
	return Broadcast(envelopeBytes, signature, caster)
}

func configUpdate(chainID string, block *cb.Block, newOrdererOrgs []*Organization, newApplicationOrgs []*Organization, newConsortiumOrgs map[string][]*Organization, orderers []string) (*cb.ConfigUpdate, error) {
	env := utils.ExtractEnvelopeOrPanic(block, 0)
	payload, err := utils.GetPayload(env)
	if err != nil {
		logger.Error("Error getting payload from block", err)
		return nil, err
	}
	configEnv := &cb.ConfigEnvelope{}
	err = proto.Unmarshal(payload.Data, configEnv)
	if err != nil {
		logger.Error("Error unmarshaling ConfigEnvelope", err)
		return nil, err
	}

	oldConf := configEnv.Config
	newConf := proto.Clone(oldConf).(*cb.Config)
	// orderers
	if orderers != nil {
		val := newConf.ChannelGroup.Values[channelconfig.OrdererAddressesKey].Value
		oa := &cb.OrdererAddresses{}
		if err = proto.Unmarshal(val, oa); err != nil {
			logger.Error("Error unmarshaling OrdererAddresses", err)
			return nil, err
		}
		oldAddrMap := make(map[string]bool)
		for _, addr := range oa.Addresses {
			oldAddrMap[addr] = true
		}
		for _, addr := range orderers {
			if !oldAddrMap[addr] {
				oa.Addresses = append(oa.Addresses, addr)
			}
		}
		newConf.ChannelGroup.Values[channelconfig.OrdererAddressesKey].Value, err = proto.Marshal(oa)
		if err != nil {
			logger.Error("Error marshaling OrdererAddresses", err)
			return nil, err
		}
	}

	// add orgs
	if newOrdererOrgs != nil {
		// system chain
		for _, org := range newOrdererOrgs {
			newConf.ChannelGroup.Groups[channelconfig.OrdererGroupKey].Groups[org.Name], err = encoder.NewOrdererOrgGroup(&localconfig.Organization{
				Name:    org.Name,
				ID:      org.ID,
				MSPDir:  org.MSPDir,
				MSPType: defaultMSPType,
			})
			if err != nil {
				logger.Error("Error creating ordererOrgGroup", err)
				return nil, err
			}
		}
	}

	if newConsortiumOrgs != nil {
		// update consortium orgs
		for name, orgs := range newConsortiumOrgs {
			var localOrgs []*localconfig.Organization
			for _, org := range orgs {
				localOrgs = append(localOrgs, &localconfig.Organization{
					Name:    org.Name,
					ID:      org.ID,
					MSPDir:  org.MSPDir,
					MSPType: defaultMSPType,
				})
			}

			if _, ok := newConf.ChannelGroup.Groups[channelconfig.ConsortiumsGroupKey].Groups[name]; !ok {
				newConf.ChannelGroup.Groups[channelconfig.ConsortiumsGroupKey].Groups[name], err = encoder.NewConsortiumGroup(&localconfig.Consortium{
					Organizations: localOrgs,
				})
			} else {
				for _, org := range localOrgs {
					newConf.ChannelGroup.Groups[channelconfig.ConsortiumsGroupKey].Groups[name].Groups[org.Name], err = encoder.NewOrdererOrgGroup(org)
					if err != nil {
						logger.Error("Error creating ordererOrgGroup", err)
						return nil, err
					}
				}
			}
		}
	}

	if newApplicationOrgs != nil {
		// application chain
		for _, org := range newApplicationOrgs {

			peers := []*localconfig.AnchorPeer{}
			for _, ap := range org.AnchorPeers {
				peers = append(peers, parseURL(ap))
			}

			newConf.ChannelGroup.Groups[channelconfig.ApplicationGroupKey].Groups[org.Name], err = encoder.NewApplicationOrgGroup(&localconfig.Organization{
				Name:        org.Name,
				ID:          org.ID,
				MSPDir:      org.MSPDir,
				MSPType:     defaultMSPType,
				AnchorPeers: peers,
			})
			if err != nil {
				logger.Error("Error creating applicationOrgGroup", err)
				return nil, err
			}
		}
	}

	updateTx, err := update.Compute(oldConf, newConf)
	if err != nil {
		return nil, err
	}
	updateTx.ChannelId = chainID
	return updateTx, nil

}

func delOrgConfigUpdate(chainID string, block *cb.Block, delOrg string, delOrderers []string) (*cb.ConfigUpdate, error) {
	logger.Info("start del org.\n")
	env := utils.ExtractEnvelopeOrPanic(block, 0)
	payload, err := utils.GetPayload(env)
	if err != nil {
		logger.Error("Error getting payload from block", err)
		return nil, err
	}
	configEnv := &cb.ConfigEnvelope{}
	err = proto.Unmarshal(payload.Data, configEnv)
	if err != nil {
		logger.Error("Error unmarshaling ConfigEnvelope", err)
		return nil, err
	}
	oldConf := configEnv.Config
	newConf := proto.Clone(oldConf).(*cb.Config)
	logger.Info("old conf:\n", oldConf)

	if delOrderers != nil {
		val := newConf.ChannelGroup.Values[channelconfig.OrdererAddressesKey].Value
		oa := &cb.OrdererAddresses{}
		if err = proto.Unmarshal(val, oa); err != nil {
			logger.Error("Error unmarshaling OrdererAddress", err)
			return nil, err
		}
		delAddrMap := make(map[string]bool)
		for _, addr := range delOrderers {
			delAddrMap[addr] = true
		}

		newOa := &cb.OrdererAddresses{}
		for _, addr := range oa.Addresses {
			if !delAddrMap[addr] {
				newOa.Addresses = append(newOa.Addresses, addr)
			}
		}
		logger.Info("new orderer address:", newOa)
		newConf.ChannelGroup.Values[channelconfig.OrdererAddressesKey].Value, err = proto.Marshal(newOa)
		if err != nil {
			logger.Error("Error marshaling OrdererAddresses", err)
			return nil, err
		}
	}

	//delete orderer orgs in system chain
	if _, ok := newConf.ChannelGroup.Groups[channelconfig.OrdererGroupKey].Groups[delOrg]; ok {
		logger.Info("delete orderer orgs in system chain,delorg:", delOrg)

		delete(newConf.ChannelGroup.Groups[channelconfig.OrdererGroupKey].Groups, delOrg)
		logger.Info("OrdererGroup:", newConf.ChannelGroup.Groups[channelconfig.OrdererGroupKey].Groups)
	}

	//delete consortium orgs
	if chainID == DefaultSystemChainID {
		logger.Info("consortiumgroup:", newConf.ChannelGroup.Groups[channelconfig.ConsortiumsGroupKey].Groups[DefaultConsortium].Groups)
		if _, ok := newConf.ChannelGroup.Groups[channelconfig.ConsortiumsGroupKey].Groups[DefaultConsortium].Groups[delOrg]; ok {
			delete(newConf.ChannelGroup.Groups[channelconfig.ConsortiumsGroupKey].Groups[DefaultConsortium].Groups, delOrg)
		}
		logger.Info("after delete consortiumgroup:", newConf.ChannelGroup.Groups[channelconfig.ConsortiumsGroupKey].Groups[DefaultConsortium].Groups)
	}

	//delete application orgs
	if chainID != DefaultSystemChainID {
		if _, ok := newConf.ChannelGroup.Groups[channelconfig.ApplicationGroupKey].Groups[delOrg]; ok {
			logger.Info("start delete application orgs:", newConf.ChannelGroup.Groups[channelconfig.ApplicationGroupKey].Groups)
			delete(newConf.ChannelGroup.Groups[channelconfig.ApplicationGroupKey].Groups, delOrg)
			logger.Info("end delete application orgs:%s", newConf.ChannelGroup.Groups[channelconfig.ApplicationGroupKey].Groups)
		}
	}

	logger.Info("end delete org.")
	logger.Info("new conf:\n", newConf)

	updateTx, err := update.Compute(oldConf, newConf)
	if err != nil {
		return nil, err
	}
	updateTx.ChannelId = chainID
	return updateTx, nil
}

func updateChannel(chainID string, block *cb.Block, newOrdererOrgs []*Organization, newApplicationOrgs []*Organization, newConsortiumOrgs map[string][]*Organization, orderers []string, caster *Endpoint, signer msp.SigningIdentity) error {
	updateTx, err := configUpdate(chainID, block, newOrdererOrgs, newApplicationOrgs, newConsortiumOrgs, orderers)
	if err != nil {
		if isNoDiffError(err) {
			logger.Warning("No differences detected between original and updated config")
			return nil
		}
		logger.Error("Error computing config update", err)
		return err
	}

	creator, err := signer.Serialize()
	if err != nil {
		logger.Error("Error serializing signer", err)
		return err
	}

	configUpdate, signatureHeader, toSignBytes, err := CreateConfigUpdateEnvelopeBytes(creator, updateTx)
	if err != nil {
		logger.Error("Error creating configUpdateEnvelopeBytes", err)
		return err
	}
	signedSigHeader, err := signer.Sign(toSignBytes)
	if err != nil {
		logger.Error("Error signning sigHeader", err)
		return err
	}

	envelopeBytes, err := CreateChannelEnvelopeBytes(chainID, creator, configUpdate, []*cb.ConfigSignature{&cb.ConfigSignature{SignatureHeader: signatureHeader, Signature: signedSigHeader}})
	if err != nil {
		logger.Error("Error creating newChannelEnvelope payload", err)
		return err
	}

	signature, err := signer.Sign(envelopeBytes)
	if err != nil {
		logger.Error("Error signning payload", err)
		return err
	}

	return Broadcast(envelopeBytes, signature, caster)

}

// GetConfigBlockByChannel ...
func (client *Client) GetConfigBlockByChannel(chainID string, deliver *Endpoint) (*cb.Block, error) {
	return getConfigBlockByChannel(chainID, deliver, client.signer)
}

func getConfigBlockByChannel(chainID string, deliver *Endpoint, signer msp.SigningIdentity) (*cb.Block, error) {
	seekI := seekInfo(seekNewest, seekNewest)
	block, err := seekBlockByChannel(chainID, seekI, deliver, signer)
	if err != nil {
		logger.Error("Error getting block by channel", err)
		return nil, err
	}
	lc, err := utils.GetLastConfigIndexFromBlock(block)
	if err != nil {
		logger.Error("Error getting last config index from block", err)
		return nil, err
	}
	return getBlockByChannel(chainID, lc, deliver, signer)
}

// GetBlockByChannel ...
func (client *Client) GetBlockByChannel(chainID string, index uint64, deliver *Endpoint) (*cb.Block, error) {
	return getBlockByChannel(chainID, index, deliver, client.signer)
}

func createBlockRequest(chainID string, seekI *ab.SeekInfo, signer msp.SigningIdentity) (*cb.Envelope, error) {
	creator, err := signer.Serialize()
	if err != nil {
		logger.Error("Error serializing", err)
		return nil, err
	}
	paylBytes, err := CreateDeliverEnvelopeBytes(chainID, seekI, creator)
	if err != nil {
		logger.Error("Error creating deliverEnvelope", err)
		return nil, err
	}
	sig, err := signer.Sign(paylBytes)
	if err != nil {
		logger.Error("Error signning", err)
		return nil, err
	}
	env := &cb.Envelope{Payload: paylBytes, Signature: sig}
	return env, nil
}

func seekBlockByChannel(chainID string, seekI *ab.SeekInfo, deliver *Endpoint, signer msp.SigningIdentity) (*cb.Block, error) {
	env, err := createBlockRequest(chainID, seekI, signer)
	if err != nil {
		logger.Error("Error creating block request envelope", err)
		return nil, err
	}
	return NewDeliverClient(deliver).RequestBlock(env)
}

func getBlocksByChannel(chainID string, seekI *ab.SeekInfo, deliver *Endpoint, signer msp.SigningIdentity) (*BlockIterator, error) {
	env, err := createBlockRequest(chainID, seekI, signer)
	if err != nil {
		logger.Error("Error creating block request envelope", err)
		return nil, err
	}
	return NewDeliverClient(deliver).RequestBlocks(env)
}

// GetBlockByChannel ...
func getBlockByChannel(chainID string, index uint64, deliver *Endpoint, signer msp.SigningIdentity) (*cb.Block, error) {
	seekS := seekSpecified(index)
	seekI := seekInfo(seekS, seekS)
	return seekBlockByChannel(chainID, seekI, deliver, signer)
}

// GetNewBlocksByChannel ...
func (client *Client) GetNewBlocksByChannel(chainID string, deliver *Endpoint) (*BlockIterator, error) {
	return getNewBlocksByChannel(chainID, deliver, client.signer)
}

func getNewBlocksByChannel(chainID string, deliver *Endpoint, signer msp.SigningIdentity) (*BlockIterator, error) {
	seekI := seekInfo(seekNewest, seekMax)
	return getBlocksByChannel(chainID, seekI, deliver, signer)
}

func getNewCommittedFilteredBlocksByChannel(chainID string, committer *Endpoint, signer msp.SigningIdentity) (*BlockIterator, error) {
	seekI := seekInfo(seekNewest, seekMax)
	return getCommittedFilteredBlocksByChannel(chainID, seekI, committer, signer)

}

// GetNewCommittedFilteredBlocksByChannel ...
func (client *Client) GetNewCommittedFilteredBlocksByChannel(chainID string, committer *Endpoint) (*BlockIterator, error) {
	return getNewCommittedFilteredBlocksByChannel(chainID, committer, client.signer)
}

func getCommittedFilteredBlocksByChannel(chainID string, seekI *ab.SeekInfo, committer *Endpoint, signer msp.SigningIdentity) (*BlockIterator, error) {
	env, err := createBlockRequest(chainID, seekI, signer)
	if err != nil {
		logger.Error("Error creating block request envelope", err)
		return nil, err
	}

	return NewPeerDeliverClient(committer).RequestFilteredBlocks(env)
}

// JoinChannel ...
func (client *Client) JoinChannel(chainID string, gb *cb.Block, endorsers []*Endpoint) error {
	return joinChannel(chainID, gb, endorsers, client.signer)
}

func joinChannel(chainID string, block *cb.Block, endorsers []*Endpoint, signer msp.SigningIdentity) error {
	spec := &pb.ChaincodeSpec{
		Type:        pb.ChaincodeSpec_GOLANG,
		ChaincodeId: &pb.ChaincodeID{Name: "cscc"},
		Input:       &pb.ChaincodeInput{Args: [][]byte{[]byte(cscc.JoinChain), utils.MarshalOrPanic(block)}},
	}

	invocation := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}
	creator, err := signer.Serialize()
	if err != nil {
		logger.Error("Error serializing identity", err)
		return err
	}

	prop, _, err := utils.CreateProposalFromCIS(cb.HeaderType_CONFIG, "", invocation, creator)
	if err != nil {
		logger.Error("Error creating proposal for join", err)
		return err
	}
	signedProp, err := utils.GetSignedProposal(prop, signer)
	if err != nil {
		logger.Error("Error creating signed proposal", err)
		return err
	}

	for _, endorser := range endorsers {
		ec, err := newEndorserClient(endorser)
		if err != nil {
			logger.Error("Error creating endorserClient", err)
			return err
		}
		defer ec.Close()
		proposalResp, err := ec.ProcessProposal(context.Background(), signedProp)
		if err != nil {
			logger.Errorf("Error processing proposal for %s: %s", endorser.Address, err)
			return err
		}

		if proposalResp == nil {
			logger.Errorf("Get nil proposal response from %s", endorser.Address)
			return errors.New("nil proposal response")
		}

		if proposalResp.Response.Status != 0 && proposalResp.Response.Status != 200 {
			logger.Errorf("bad proposal response %d: %s", proposalResp.Response.Status, proposalResp.Response.Message)
			return errors.New("bad proposal response")
		}
		logger.Infof("Successfully submitted proposal to join channel for %s", endorser.Address)
	}

	return nil
}

// CreateChannel ...
func (client *Client) CreateChannel(conf *ChannelConfig, caster *Endpoint) error {
	creator, err := client.signer.Serialize()
	if err != nil {
		logger.Error("Error serializing identity", err)
		return err
	}
	config := newChannelProfile(conf)
	newChannelConfigUpdate, err := encoder.NewChannelCreateConfigUpdate(conf.ChainID, nil, config)
	if err != nil {
		logger.Error("Error creating configUpdate", err)
		return err
	}

	configUpdate, sigHeader, toSign, err := CreateConfigUpdateEnvelopeBytes(creator, newChannelConfigUpdate)
	if err != nil {
		logger.Error("Error creating configUpdateEnvelopeBytes", err)
		return err
	}

	signedSigHeader, err := client.signer.Sign(toSign)
	if err != nil {
		logger.Error("Error signning sigHeader", err)
		return err
	}

	payload, err := CreateChannelEnvelopeBytes(conf.ChainID, creator, configUpdate, []*cb.ConfigSignature{&cb.ConfigSignature{SignatureHeader: sigHeader, Signature: signedSigHeader}})
	if err != nil {
		logger.Error("Error creating newChannelEnvelope payload", err)
		return err
	}

	signature, err := client.signer.Sign(payload)
	if err != nil {
		logger.Error("Error signning payload", err)
		return err
	}

	return Broadcast(payload, signature, caster)
}

// CreateChannelTx ...
func CreateChannelTx(config *ChannelConfig) (*cb.Envelope, error) {
	conf := newChannelProfile(config)
	return encoder.MakeChannelCreationTransaction(config.ChainID, nil, nil, conf)
}

// WriteChannelTx ...
func WriteChannelTx(output string, env *cb.Envelope) error {
	logger.Info("Writing new channel tx")
	return ioutil.WriteFile(output, utils.MarshalOrPanic(env), 0644)
}

// CreateGenesisBlock ...
// No need to sign it
func CreateGenesisBlock(config *GenesisConfig) *cb.Block {
	conf := newGenesisProfile(config)
	pgen := encoder.New(conf)
	logger.Info("Generating genesis block")
	if conf.Consortiums == nil {
		logger.Warning("Genesis block does not contain a consortiums group definition.  This block cannot be used for orderer bootstrap.")
	}
	return pgen.GenesisBlockForChannel(config.ChainID)
}

// WriteGenesisBlock ...
func WriteGenesisBlock(output string, block *cb.Block) error {
	logger.Info("Writing genesis block")
	return ioutil.WriteFile(output, utils.MarshalOrPanic(block), 0644)

}

// format: 'grpcs://xxxx:xx'
func parseURL(rawURL string) *localconfig.AnchorPeer {
	anchor := &localconfig.AnchorPeer{}
	addr, err := url.Parse(rawURL)
	if err != nil {
		logger.Panic(err)
	}
	anchor.Host = addr.Hostname()
	anchor.Port, _ = strconv.Atoi(addr.Port())
	return anchor
}

func newChannelProfile(conf *ChannelConfig) *localconfig.Profile {
	profile := &localconfig.Profile{}

	profile.Consortium = conf.Consortium

	profile.Application = &localconfig.Application{}

	// add policy
	profile.Application.Policies = make(map[string]*localconfig.Policy)

	defaultAdmins := PolicyAnyAdmins
	if conf.AdminsPolicy != "" {
		defaultAdmins = conf.AdminsPolicy
	}
	profile.Application.Policies["Admins"] = &localconfig.Policy{
		Type: defaultPolicyType,
		Rule: string(defaultAdmins),
	}

	defaultWriters := PolicyAnyWriters
	if conf.WritersPolicy != "" {
		defaultWriters = conf.WritersPolicy
	}
	profile.Application.Policies["Writers"] = &localconfig.Policy{
		Type: defaultPolicyType,
		Rule: string(defaultWriters),
	}

	defaultReaders := PolicyAnyReaders
	if conf.ReadersPolicy != "" {
		defaultReaders = conf.ReadersPolicy
	}
	profile.Application.Policies["Readers"] = &localconfig.Policy{
		Type: defaultPolicyType,
		Rule: string(defaultReaders),
	}

	orgs := []*localconfig.Organization{}
	for _, org := range conf.Organizations {

		peers := []*localconfig.AnchorPeer{}
		for _, ap := range org.AnchorPeers {
			peers = append(peers, parseURL(ap))
		}
		orgs = append(orgs, &localconfig.Organization{
			Name:        org.Name,
			ID:          org.ID,
			MSPDir:      org.MSPDir,
			MSPType:     defaultMSPType,
			AnchorPeers: peers,
		})
	}
	profile.Application.Organizations = orgs
	profile.Application.Capabilities = make(map[string]bool)
	profile.Application.Capabilities[defaultApplicationCapability] = true

	return profile
}

func newGenesisProfile(conf *GenesisConfig) *localconfig.Profile {
	profile := &localconfig.Profile{}
	orderer := &localconfig.Orderer{}
	orderer.Addresses = conf.Addresses
	orderer.OrdererType = conf.OrdererType

	orderer.BatchTimeout = conf.BatchTimeout
	if conf.BatchTimeout == time.Duration(0) {
		orderer.BatchTimeout = defaultBatchTimeout
	}

	orderer.BatchSize = localconfig.BatchSize{
		MaxMessageCount:   conf.MaxMessageCount,
		AbsoluteMaxBytes:  conf.AbsoluteMaxBytes,
		PreferredMaxBytes: conf.PreferredMaxBytes,
	}
	if conf.MaxMessageCount == 0 {
		orderer.BatchSize.MaxMessageCount = defaultMaxMessageCount
	}
	if conf.AbsoluteMaxBytes == 0 {
		orderer.BatchSize.AbsoluteMaxBytes = defaultAbsoluteMaxBytes
	}
	if conf.PreferredMaxBytes == 0 {
		orderer.BatchSize.PreferredMaxBytes = defaultPreferredMaxBytes
	}

	orderer.Kafka.Brokers = conf.KafkaBrokers

	orderer.MaxChannels = conf.MaxChannels

	for _, org := range conf.OrdererOrganizations {
		orderer.Organizations = append(orderer.Organizations, &localconfig.Organization{
			Name:    org.Name,
			ID:      org.ID,
			MSPDir:  org.MSPDir,
			MSPType: defaultMSPType,
		})
	}

	orderer.Capabilities = make(map[string]bool)
	orderer.Capabilities[defaultOrdererCapability] = true
	orderer.Policies = make(map[string]*localconfig.Policy)

	profile.Orderer = orderer
	profile.Policies = make(map[string]*localconfig.Policy)

	profile.Consortiums = make(map[string]*localconfig.Consortium)

	consortiumOrgs := []*localconfig.Organization{}
	for _, org := range conf.ConsortiumOrganizations {
		consortiumOrgs = append(consortiumOrgs, &localconfig.Organization{
			Name:    org.Name,
			ID:      org.ID,
			MSPDir:  org.MSPDir,
			MSPType: defaultMSPType,
		})
	}

	consortium := &localconfig.Consortium{
		Organizations: consortiumOrgs,
	}

	if conf.ConsortiumName == "" {
		profile.Consortiums[DefaultConsortium] = consortium
	} else {
		profile.Consortiums[conf.ConsortiumName] = consortium
	}

	profile.Capabilities = make(map[string]bool)
	profile.Capabilities[defaultChannelCapability] = true

	defaultAdmins := PolicyAnyAdmins
	if conf.AdminsPolicy != "" {
		defaultAdmins = conf.AdminsPolicy
	}

	profile.Orderer.Policies["Admins"] = &localconfig.Policy{
		Type: defaultPolicyType,
		Rule: string(defaultAdmins),
	}
	profile.Policies["Admins"] = profile.Orderer.Policies["Admins"]

	defaultWriters := PolicyAnyWriters
	if conf.WritersPolicy != "" {
		defaultWriters = conf.WritersPolicy
	}
	profile.Orderer.Policies["Writers"] = &localconfig.Policy{
		Type: defaultPolicyType,
		Rule: string(defaultWriters),
	}
	profile.Policies["Writers"] = profile.Orderer.Policies["Writers"]

	defaultReaders := PolicyAnyReaders
	if conf.ReadersPolicy != "" {
		defaultReaders = conf.ReadersPolicy
	}
	profile.Orderer.Policies["Readers"] = &localconfig.Policy{
		Type: defaultPolicyType,
		Rule: string(defaultReaders),
	}
	profile.Policies["Readers"] = profile.Orderer.Policies["Readers"]

	return profile

}

func isNoDiffError(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "no differences detected between original and updated config"
}
