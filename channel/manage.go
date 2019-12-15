package channel

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/astaxie/beego"
	cb "github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/hyperledger/fabric/sdk"
)

const (
	tmpMSPDir = "tmpMspDir"
)

const (
	defaultConsensusType = "kafka"
)

func (c *Channel) IdentityCode() (*IdentityCode, error) {
	mspData, err := c.orgs[0].OrgCA.MSPBytes(c.orgs[0].OrgMSP)
	if err != nil {
		logger.Error("Error generating msp bytes", err)
		return nil, err
	}
	orderers := Orderers(c.orgs[0].OrdererNodes)
	anchors := AnchorPeers(c.orgs[0].PeerNodes)
	chainOrgInfo := &ChainOrgInfo{}
	chainOrgInfo.Peers = serviceNodesToEndpointList(c.orgs[0].PeerNodes, CreateChannelTimeout, c.orgs[0].OrgCA.TLSCACert())
	chainOrgInfo.Orderers = serviceNodesToEndpointList(c.orgs[0].OrdererNodes, CreateChannelTimeout, c.orgs[0].OrgCA.TLSCACert())
	chainOrgInfo.OrgName = c.orgs[0].OrgName
	// chainOrgInfo.ChannelName =
	return &IdentityCode{
		Org:          c.orgs[0].OrgName,
		OrgMSP:       mspData,
		Orderers:     orderers,
		Anchors:      anchors,
		ChainOrgInfo: chainOrgInfo,
	}, nil
}

func (c *Channel) AddOrg(identity []byte, operateOrg []*OrgInfo, channelName string) error {
	logger.Info("start add org")
	ic := &IdentityCode{}
	if err := json.Unmarshal(identity, ic); err != nil {
		logger.Error("error unmarshal", err)
		return err
	}
	logger.Info("mspdata:%s", ic.OrgMSP)
	mspDir, mspID, err := sdk.WriteMSPDir(tmpMSPDir, ic.OrgMSP)
	if err != nil {
		logger.Error("error writing certs to msp dir", err)
		return err
	}

	broadcasters := serviceNodesToEndpointList(operateOrg[0].OrdererNodes, CreateChannelTimeout, operateOrg[0].OrgCA.TLSCACert())

	peerOrgs := []*sdk.Organization{&sdk.Organization{
		Name:        mspID,
		ID:          mspID,
		MSPDir:      mspDir,
		AnchorPeers: ic.Anchors,
	}}

	ordererOrgs := []*sdk.Organization{&sdk.Organization{
		Name:   mspID,
		ID:     mspID,
		MSPDir: mspDir,
	}}
	consortiumOrgs := make(map[string][]*sdk.Organization)
	consortiumOrgs[DefaultConsortium] = peerOrgs

	systemUpdate, err := c.createAddOrgChannelConfigUpdate(sdk.DefaultSystemChainID, nil, ordererOrgs, consortiumOrgs, ic.Orderers, broadcasters)
	if err != nil {
		logger.Error("Error create system channel config update", err)
		return err
	}
	channelUpdate, err := c.createAddOrgChannelConfigUpdate(channelName, peerOrgs, ordererOrgs, nil, ic.Orderers, broadcasters)
	if err != nil {
		logger.Error("Error create channel config update", err)
		return err
	}

	//sign
	systemSigs := []*cb.ConfigSignature{}
	channelSigs := []*cb.ConfigSignature{}
	for _, org := range operateOrg {
		sSigHeader, sSignedSigHeader, err := org.Client.SignChannelConfigUpdate(systemUpdate)
		if err != nil {
			logger.Error("Error signing system config update", err)
			return err
		}
		cSigHeader, cSignedSigHeader, err := org.Client.SignChannelConfigUpdate(channelUpdate)
		if err != nil {
			logger.Error("Error signing channel config update", err)
			return err
		}

		systemSigs = append(systemSigs, &cb.ConfigSignature{
			SignatureHeader: sSigHeader,
			Signature:       sSignedSigHeader,
		})

		channelSigs = append(channelSigs, &cb.ConfigSignature{
			SignatureHeader: cSigHeader,
			Signature:       cSignedSigHeader,
		})

	}

	for _, broadcaster := range broadcasters {
		err = operateOrg[0].Client.UpdateChannelByConfigUpdate(sdk.DefaultSystemChainID, systemUpdate, systemSigs, broadcaster)
		if err != nil {
			logger.Error("Error update system channel", err)
			continue
		}
		err = operateOrg[0].Client.UpdateChannelByConfigUpdate(channelName, channelUpdate, channelSigs, broadcaster)
		if err != nil {
			logger.Error("Error update channel", err)
			return err
		}
		logger.Info("Suceesfully add new org")
		return nil
	}

	logger.Info("end add new org.")
	return nil
}

func (c *Channel) DeleteOrg(delOrg string, delOrderers []string, channelName string, operateOrg []*OrgInfo) error {
	logger.Info("start delete org.")
	broadcasters := serviceNodesToEndpointList(operateOrg[0].OrdererNodes, CreateChannelTimeout, operateOrg[0].OrgCA.TLSCACert())

	systemUpdate, err := c.createDelOrgChannelConfigUpdate(sdk.DefaultSystemChainID, delOrg, delOrderers, broadcasters)
	if err != nil {
		logger.Error("Error create system channel config update", err)
		return err
	}
	logger.Info("systemUpdate:%s", systemUpdate)
	channelUpdate, err := c.createDelOrgChannelConfigUpdate(channelName, delOrg, delOrderers, broadcasters)
	if err != nil {
		logger.Error("Error create channle config update", err)
		return err
	}
	logger.Info("channelUpdate:%s", channelUpdate)

	// sign
	systemSigs := []*cb.ConfigSignature{}
	channelSigs := []*cb.ConfigSignature{}
	for _, org := range operateOrg {
		sSigHeader, sSignedSigHeader, err := org.Client.SignChannelConfigUpdate(systemUpdate)
		if err != nil {
			logger.Error("Error signing system config update", err)
			return err
		}
		cSigHeader, cSignedSigHeader, err := org.Client.SignChannelConfigUpdate(channelUpdate)
		if err != nil {
			logger.Error("Error signing channel config update", err)
		}

		systemSigs = append(systemSigs, &cb.ConfigSignature{
			SignatureHeader: sSigHeader,
			Signature:       sSignedSigHeader,
		})

		channelSigs = append(channelSigs, &cb.ConfigSignature{
			SignatureHeader: cSigHeader,
			Signature:       cSignedSigHeader,
		})
	}

	for _, broadcaster := range broadcasters {
		err := operateOrg[0].Client.UpdateChannelByConfigUpdate(sdk.DefaultSystemChainID, systemUpdate, systemSigs, broadcaster)
		if err != nil {
			logger.Error("Error update system channel", err)
			continue
		}

		err = operateOrg[0].Client.UpdateChannelByConfigUpdate(channelName, channelUpdate, channelSigs, broadcaster)
		if err != nil {
			logger.Error("Error update channel ", err)
			return err
		}
		logger.Info("Succeesfully delete org.")
	}
	logger.Info("end delete org.")
	return nil
}

func (c *Channel) createAddOrgChannelConfigUpdate(chainID string, peerOrgs, ordererOrgs []*sdk.Organization, consortiumOrgs map[string][]*sdk.Organization, orderers []string, casters []*sdk.Endpoint) ([]byte, error) {
	for _, caster := range casters {
		configBlock, err := c.orgs[0].Client.GetConfigBlockByChannel(chainID, caster)
		if err != nil {
			logger.Error("Error getting config block from chain %s: %s", chainID, err)
			continue
		}
		logger.Info("Successfully getting config block from chain %s", chainID)

		return c.orgs[0].Client.GetAddOrgChannelConfigUpdate(chainID, configBlock, ordererOrgs, peerOrgs, consortiumOrgs, orderers)
	}
	return nil, errors.New("failed getAddOrgChannelConfigUpdate after try all orderers")
}

func (c Channel) createDelOrgChannelConfigUpdate(chainID string, delOrg string, delOrderers []string, casters []*sdk.Endpoint) ([]byte, error) {
	logger.Info("start createDelOrgChannelConfigUpdate.")
	for _, caster := range casters {
		configBlock, err := c.orgs[0].Client.GetConfigBlockByChannel(chainID, caster)
		if err != nil {
			logger.Error("Error getting config block from chain %s: %s", chainID, err)
			continue
		}
		logger.Info("Successfully getting config block from chain %s", chainID)
		return c.orgs[0].Client.GetDelOrgChannelConfigUpdate(chainID, configBlock, delOrg, delOrderers)
	}
	logger.Info("end createDelOrgChannelConfigUpdate.")
	return nil, errors.New("failed getDelOrgChannelConfigUpdate after try all orderers")
}

func GenerateCrypto(orgs []*OrgInfo) error {
	mspDir := beego.AppConfig.String("MSPDir")
	// gm, _ := beego.AppConfig.Bool("GM")
	var orginfos []*OrgInfo
	for _, org := range orgs {
		orgCA, err := GetCA(path.Join(mspDir, org.OrgName), org.OrgName)
		if err != nil {
			logger.Error("Error getting peer ca", err)
			return err
		}
		org.OrgCA = orgCA
		orginfos = append(orginfos, org)
	}

	for _, org := range orginfos {
		//orderers
		for _, orderer := range org.OrdererNodes {
			var san []string
			san = append(san, splitIP(orderer.ExternalEndpoint))
			certs := []*sdk.CertConfig{&sdk.CertConfig{
				CN:       orderer.ID,
				SAN:      filterSAN(san),
				NodeType: sdk.OrdererNode,
			}}
			if err := org.OrgCA.GenerateMSP(certs, nil); err != nil {
				logger.Error("Error generating msp", err)
				return err
			}
		}
		//peers
		for _, peer := range org.PeerNodes {
			var san []string
			san = append(san, splitIP(peer.ExternalEndpoint))
			certs := []*sdk.CertConfig{&sdk.CertConfig{
				CN:       peer.ID,
				SAN:      filterSAN(san),
				NodeType: sdk.PeerNode,
			}}
			if err := org.OrgCA.GenerateMSP(certs, nil); err != nil {
				logger.Error("Error generating msp", err)
				return err
			}
		}
	}
	return nil
}

func GetCA(dir string, msp string) (*sdk.CA, error) {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return sdk.NewCA(dir, msp)
	}
	if err == nil {
		if !info.IsDir() {
			return nil, errors.New("msp path is not a directory, but a file")
		}
		return sdk.ConstructCAFromDir(dir)
	}
	return nil, err
}

func splitIP(addr string) string {
	return strings.Split(addr, ":")[0]
}

func filterSAN(san []string) (ret []string) {
	m := make(map[string]bool)
	for _, ip := range san {
		if ip != "" {
			m[ip] = true
		}
	}
	for k := range m {
		ret = append(ret, k)
	}
	return
}

func GenGenesisBlock(orgs []*OrgInfo, kafkas []string) (*cb.Block, error) {

	var orderers []string
	var peerOrgs []*sdk.Organization
	var ordererOrgs []*sdk.Organization

	for _, org := range orgs {
		for _, orderer := range org.OrdererNodes {
			orderers = append(orderers, orderer.ExternalEndpoint)
		}
	}
	logger.Info("orderers:", orderers)

	for _, org := range orgs {
		peerOrg := &sdk.Organization{
			Name:   org.OrgMSP,
			ID:     org.OrgMSP,
			MSPDir: org.OrgCA.MSPDir(),
		}
		ordererOrg := &sdk.Organization{
			Name:   org.OrgMSP,
			ID:     org.OrgMSP,
			MSPDir: org.OrgCA.MSPDir(),
		}
		peerOrgs = append(peerOrgs, peerOrg)
		ordererOrgs = append(ordererOrgs, ordererOrg)
	}
	logger.Info("peerorgs,ordererorgs:", peerOrgs, ordererOrgs)

	conf := &sdk.GenesisConfig{
		ChainID:                 sdk.DefaultSystemChainID,
		OrdererType:             defaultConsensusType,
		Addresses:               orderers,
		AdminsPolicy:            sdk.PolicyMajorityAdmins,
		WritersPolicy:           sdk.PolicyAnyWriters,
		ReadersPolicy:           sdk.PolicyAnyReaders,
		OrdererOrganizations:    ordererOrgs,
		ConsortiumOrganizations: peerOrgs,
		ConsortiumName:          sdk.DefaultConsortium,
		KafkaBrokers:            kafkas,
	}
	logger.Info("genesis block conf:", conf)
	block := sdk.CreateGenesisBlock(conf)

	err := ioutil.WriteFile("orderer.block", utils.MarshalOrPanic(block), 0644)
	if err != nil {
		logger.Info("write file err:", err)
	}
	return block, nil
}
