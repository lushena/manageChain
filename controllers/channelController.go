package controllers

import (
	"encoding/json"
	"manageChain/channel"
	"net/url"
	"path"
	"strconv"

	"github.com/astaxie/beego"
	logger "github.com/astaxie/beego/logs"
	"github.com/hyperledger/fabric/common/tools/configtxgen/localconfig"
)

type ChannelController struct {
	BaseController
}

func newChannel(orgs []*channel.OrgInfo) (*channel.Channel, error) {
	mspDir := beego.AppConfig.String("MSPDir")
	gm, _ := beego.AppConfig.Bool("GM")
	var orginfo []*channel.OrgInfo

	for _, org := range orgs {
		orgCA, err := channel.GetCA(path.Join(mspDir, org.OrgName), org.OrgName)
		if err != nil {
			logger.Error("Error getting peer ca", err)
			return nil, err
		}
		org.OrgCA = orgCA
		orginfo = append(orginfo, org)
	}

	return channel.NewChannel(orginfo, gm)
}

func (c *ChannelController) CreateChannel() error {
	logger.Info("start create channel")

	ccr := &channel.NewCreateChannelRequest{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, ccr)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}

	channelName := ccr.ChannelName
	channel, err := newChannel(ccr.Orgs)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}

	err = channel.CreateChannel(channelName)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}
	c.ReturnOKMsg("OK")
	logger.Info("successfully create channel")
	return nil
}

func (c *ChannelController) JoinChannel() error {
	logger.Info("start join channel")

	jcr := &channel.JoinChannelRequest{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, jcr)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}

	channelName := jcr.ChannelName
	channel, err := newChannel(jcr.Orgs)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}

	err = channel.JoinChannel(channelName)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}
	c.ReturnOKMsg("OK")
	logger.Info("successfully join channel")
	return nil
}

func (c *ChannelController) GenCrypto() error {
	logger.Info("start generate crypto config")
	genCryptoReq := &channel.GenCryptoRequest{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, genCryptoReq)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}
	orgs := genCryptoReq.Orgs
	err = channel.GenerateCrypto(orgs)
	if err != nil {
		logger.Error("Error generate crypto")
		c.ReturnErrorMsg(err)
		return nil
	}

	logger.Info("end generate crypto config")
	return nil
}

// Generate genesis block
func (c *ChannelController) GenGenesisBlock() error {
	logger.Info("start generate Genesis block")
	genGbReq := &channel.GenGenesisBlockRequest{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, genGbReq)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}
	orgs := genGbReq.Orgs
	kafkas := genGbReq.Kafkas

	var orginfos []*channel.OrgInfo
	mspDir := beego.AppConfig.String("MSPDir")
	for _, org := range orgs {
		orgCA, err := channel.GetCA(path.Join(mspDir, org.OrgName), org.OrgName)
		if err != nil {
			logger.Error("Error getting peer ca", err)
			c.ReturnErrorMsg(err)
			return nil
		}
		org.OrgCA = orgCA
		orginfos = append(orginfos, org)
	}
	block, err := channel.GenGenesisBlock(orginfos, kafkas)
	if err != nil {
		logger.Error("error generate genesis block:", err)
		c.ReturnErrorMsg(err)
		return nil
	}
	logger.Info("genesis block:", block)
	c.ReturnOKMsg("generate config block ok")
	logger.Info("end genetate Genesis block")
	return nil
}

// format: 'grpcs://xxxx:xx'
func parseURL(rawURL string) *localconfig.AnchorPeer {
	anchor := &localconfig.AnchorPeer{}
	addr, err := url.Parse(rawURL)
	if err != nil {
		logger.Error("err parse", err)
		return nil
	}
	anchor.Host = addr.Hostname()
	anchor.Port, _ = strconv.Atoi(addr.Port())
	return anchor
}

// Identity ...
func (c *ChannelController) Identity() error {
	logger.Info("start generate org identity")

	idr := &channel.IdentityRequest{}
	// logger.Info("reqbody:", c.Ctx.Input.RequestBody)
	err := json.Unmarshal(c.Ctx.Input.RequestBody, idr)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}

	orgs := idr.Orgs
	newChannel, err := newChannel(orgs)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}

	id, err := newChannel.IdentityCode()
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}
	c.ReturnOKMsg(id)
	logger.Info("successfully generate identity")
	return nil
}

func (c *ChannelController) AddOrg() error {
	logger.Info("start add org.")
	addOrgReq := &channel.AddOrgRequest{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &addOrgReq)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}
	orgs := addOrgReq.Orgs
	channelName := addOrgReq.ChannelName
	newChannel, err := newChannel(orgs)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}
	id := addOrgReq.Identity
	err = newChannel.AddOrg(id, orgs, channelName)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}

	c.ReturnOKMsg("OK")
	logger.Info("successfully add org.")
	return nil
}

func (c *ChannelController) DeleteOrg() error {
	logger.Info("start delete org")
	delOrgReq := &channel.DeleteOrgRequest{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, delOrgReq)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}
	delOrg := delOrgReq.DelOrg
	delOrderers := delOrgReq.DelOrderers
	channelName := delOrgReq.ChannelName
	operateOrg := delOrgReq.Orgs
	newChannel, err := newChannel(operateOrg)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}
	err = newChannel.DeleteOrg(delOrg, delOrderers, channelName, operateOrg)
	if err != nil {
		c.ReturnErrorMsg(err)
		return nil
	}

	c.ReturnOKMsg("OK")
	logger.Info("successfully delete org.")
	return nil
}
