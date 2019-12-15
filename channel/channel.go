package channel

import (
	"encoding/json"
	"errors"
	"fmt"
	logs "gglogs"

	cb "github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/sdk"
)

var logger *logs.BeeLogger

func init() {
	logger = logs.GetBeeLogger()
}

type Channel struct {
	orgs []*OrgInfo
}

func NewChannel(orgs []*OrgInfo, gm bool) (*Channel, error) {
	if 0 == len(orgs) {
		logger.Error("args err")
		return nil, errors.New("args err")
	}
	channel := &Channel{}
	for _, org := range orgs {
		orgMSP := org.OrgMSP
		orgCA := org.OrgCA

		client, err := sdk.NewClient(orgCA.AdminCommonName(), orgMSP, orgCA.AdminMSPDir(), gm)
		if err != nil {
			logger.Error("Error creating client for org", err)
			return nil, err
		}
		org.Client = client
		channel.orgs = append(channel.orgs, org)
	}

	return channel, nil
}

func (c *Channel) CreateChannel(ChainID string) error {

	var organizations []*sdk.Organization
	for _, org := range c.orgs {
		anchors := AnchorPeers(org.PeerNodes)
		if len(anchors) == 0 {
			logger.Warning("no anchors can be found, cross-org gossip can't work")
		}
		organizations = append(organizations, &sdk.Organization{
			Name:        org.OrgMSP,
			ID:          org.OrgMSP,
			MSPDir:      org.OrgCA.MSPDir(),
			AnchorPeers: anchors,
		})
	}

	conf := &sdk.ChannelConfig{
		ChainID:       ChainID,
		Consortium:    sdk.DefaultConsortium,
		AdminsPolicy:  sdk.PolicyMajorityAdmins,
		ReadersPolicy: sdk.PolicyAnyReaders,
		WritersPolicy: sdk.PolicyAnyWriters,
		Organizations: organizations,
	}

	//use org1
	orgCA := c.GetOrgCA()
	casters := serviceNodesToEndpointList(c.orgs[0].OrdererNodes, CreateChannelTimeout, orgCA.TLSCACert())

	for _, caster := range casters {
		if err := c.orgs[0].Client.CreateChannel(conf, caster); err != nil {
			logger.Error("Error creating channel", err)
			continue
		} else {
			logger.Info("Successfully creating channel")
			return nil
		}
	}

	return errors.New(fmt.Sprintf("failed creating %s chain after try all orderers", ChainID))

}

//use org1
func (c *Channel) GetOrgCA() *sdk.CA {
	return c.orgs[0].OrgCA
}

func (c *Channel) JoinChannel(channelName string) error {
	var err error
	var block *cb.Block
	orgCA := c.GetOrgCA()
	endorsers := serviceNodesToEndpointList(c.orgs[0].PeerNodes, EndorseTimeout, orgCA.TLSCACert())
	casters := serviceNodesToEndpointList(c.orgs[0].OrdererNodes, CreateChannelTimeout, orgCA.TLSCACert())
	for _, caster := range casters {
		if block, err = c.orgs[0].Client.GetBlockByChannel(channelName, 0, caster); err == nil {
			break
		}
		logger.Error("Error getting block", err)
	}
	if err != nil {
		return errors.New("failed getting block after try all orderers")
	}
	return c.orgs[0].Client.JoinChannel(channelName, block, endorsers)
}

func (c *Channel) peers(channelName string) (peers []*sdk.Endpoint, err error) {
	args := [][]byte{
		[]byte(getChainOrgInfo),
		[]byte(channelName),
		[]byte(c.orgs[0].OrgName),
	}

	endorsers := serviceNodesToEndpointList(c.orgs[0].PeerNodes, CreateChannelTimeout, c.orgs[0].OrgCA.TLSCACert())
	data, err := query(c.orgs[0].Client, PublicChainID, PublicCCName, args, endorsers)
	if err != nil {
		logger.Error("Error querying", err)
		return nil, err
	}

	orgChainInfo := &ChainOrgInfo{}
	if err := json.Unmarshal(data, orgChainInfo); err != nil {
		logger.Error("Error unmarshaling invitation", err)
		return nil, err
	}

	peers = orgChainInfo.Peers

	if len(peers) == 0 {
		err = errors.New("no peers can be found")
	}
	logger.Info("channel:%s, orgName:%s, peers:%s", orgChainInfo.ChannelName, orgChainInfo.OrgName, peers)
	return
}

func (c *Channel) orderers(channelName string) (orderers []*sdk.Endpoint, err error) {
	args := [][]byte{
		[]byte(getChainOrgInfo),
		[]byte(channelName),
		[]byte(c.orgs[0].OrgName),
	}
	endorsers := serviceNodesToEndpointList(c.orgs[0].PeerNodes, CreateChannelTimeout, c.orgs[0].OrgCA.TLSCACert())
	data, err := query(c.orgs[0].Client, PublicChainID, PublicCCName, args, endorsers)
	if err != nil {
		logger.Error("Error querying", err)
		return nil, err
	}

	orgChainInfo := &ChainOrgInfo{}
	if err := json.Unmarshal(data, orgChainInfo); err != nil {
		logger.Error("Error unmarshaling invitation", err)
		return nil, err
	}
	orderers = orgChainInfo.Orderers
	if len(orderers) == 0 {
		err = errors.New("no orderers can be found")
	}
	logger.Info("channel:%s, orgName:%s, orderers:%s", orgChainInfo.ChannelName, orgChainInfo.OrgName, orderers)
	return
}

func (c *Channel) peersAndOrderers(channelName string) (peers, orderers []*sdk.Endpoint, err error) {
	args := [][]byte{
		[]byte(getChainOrgInfo),
		[]byte(channelName),
		[]byte(c.orgs[0].OrgName),
	}
	endorsers := serviceNodesToEndpointList(c.orgs[0].PeerNodes, CreateChannelTimeout, c.orgs[0].OrgCA.TLSCACert())
	data, err := query(c.orgs[0].Client, PublicChainID, PublicCCName, args, endorsers)
	if err != nil {
		logger.Error("Error querying", err)
		return nil, nil, err
	}
	orgChainInfo := &ChainOrgInfo{}
	if err := json.Unmarshal(data, orgChainInfo); err != nil {
		logger.Error("Error unmarshaling invitation", err)
		return nil, nil, err
	}
	peers = orgChainInfo.Peers
	orderers = orgChainInfo.Orderers
	return
}
