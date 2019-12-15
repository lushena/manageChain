package channel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func TestGenCrypto(t *testing.T) {
	org1Peers := []*ServiceNode{
		&ServiceNode{
			ID:               "peer0",
			Endpoint:         "172.16.93.215:56051",
			ExternalEndpoint: "172.16.93.215:56051",
			Public:           true,
		},
		&ServiceNode{
			ID:               "peer1",
			Endpoint:         "172.16.93.215:56151",
			ExternalEndpoint: "172.16.93.215:56151",
			Public:           true,
		},
	}

	org1Orderers := []*ServiceNode{
		&ServiceNode{
			ID:               "orderer0",
			Endpoint:         "172.16.93.215:56050",
			ExternalEndpoint: "172.16.93.215:56050",
			Public:           true,
		},
	}

	org2Peers := []*ServiceNode{
		&ServiceNode{
			ID:               "peer0",
			Endpoint:         "172.16.93.215:56251",
			ExternalEndpoint: "172.16.93.215:56251",
			Public:           true,
		},
		&ServiceNode{
			ID:               "peer1",
			Endpoint:         "172.16.93.215:56351",
			ExternalEndpoint: "172.16.93.215:56351",
			Public:           true,
		},
	}

	org2Orderers := []*ServiceNode{
		&ServiceNode{
			ID:               "orderer0",
			Endpoint:         "172.16.93.215:57050",
			ExternalEndpoint: "172.16.93.215:57050",
			Public:           true,
		},
	}
	org3Peers := []*ServiceNode{
		&ServiceNode{
			ID:               "peer0",
			Endpoint:         "172.16.93.215:56451",
			ExternalEndpoint: "172.16.93.215:56451",
			Public:           true,
		},
		&ServiceNode{
			ID:               "peer1",
			Endpoint:         "172.16.93.215:56551",
			ExternalEndpoint: "172.16.93.215:56551",
			Public:           true,
		},
	}

	org3Orderers := []*ServiceNode{
		&ServiceNode{
			ID:               "orderer0",
			Endpoint:         "172.16.93.215:58050",
			ExternalEndpoint: "172.16.93.215:58050",
			Public:           true,
		},
	}
	orgs := []*OrgInfo{
		&OrgInfo{
			OrgName:      "testorg1",
			OrgMSP:       "testorg1",
			MspID:        "testorg1",
			PeerNodes:    org1Peers,
			OrdererNodes: org1Orderers,
		},
		&OrgInfo{
			OrgName:      "testorg2",
			OrgMSP:       "testorg2",
			MspID:        "testorg2",
			PeerNodes:    org2Peers,
			OrdererNodes: org2Orderers,
		},
		&OrgInfo{
			OrgName:      "testorg3",
			OrgMSP:       "testorg3",
			MspID:        "testorg3",
			PeerNodes:    org3Peers,
			OrdererNodes: org3Orderers,
		},
	}

	genCryptoreq := &GenCryptoRequest{
		Orgs: orgs,
	}

	data, err := json.Marshal(genCryptoreq)
	if err != nil {
		t.Fatal(err)
	}

	wrt := bytes.NewBuffer(data)

	resp, err := http.Post("http://127.0.0.1:8080/gencrypto", "application/json", wrt)

	if err != nil {
		t.Fatal(err)
	}
	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(ret))
}

func TestGenGenesisBlock(t *testing.T) {
	org1Peers := []*ServiceNode{
		&ServiceNode{
			ID:               "peer0",
			Endpoint:         "172.16.93.215:56051",
			ExternalEndpoint: "172.16.93.215:56051",
			Public:           true,
		},
		&ServiceNode{
			ID:               "peer1",
			Endpoint:         "172.16.93.215:56151",
			ExternalEndpoint: "172.16.93.215:56151",
			Public:           true,
		},
	}

	org1Orderers := []*ServiceNode{
		&ServiceNode{
			ID:               "orderer0",
			Endpoint:         "172.16.93.215:56050",
			ExternalEndpoint: "172.16.93.215:56050",
			Public:           true,
		},
	}
	org2Peers := []*ServiceNode{
		&ServiceNode{
			ID:               "peer0",
			Endpoint:         "172.16.93.215:56251",
			ExternalEndpoint: "172.16.93.215:56251",
			Public:           true,
		},
		&ServiceNode{
			ID:               "peer1",
			Endpoint:         "172.16.93.215:56351",
			ExternalEndpoint: "172.16.93.215:56351",
			Public:           true,
		},
	}

	org2Orderers := []*ServiceNode{
		&ServiceNode{
			ID:               "orderer0",
			Endpoint:         "172.16.93.215:57050",
			ExternalEndpoint: "172.16.93.215:57050",
			Public:           true,
		},
	}

	orgs := []*OrgInfo{
		&OrgInfo{
			OrgName:      "testorg1",
			OrgMSP:       "testorg1",
			MspID:        "testorg1",
			PeerNodes:    org1Peers,
			OrdererNodes: org1Orderers,
		},
		&OrgInfo{
			OrgName:      "testorg2",
			OrgMSP:       "testorg2",
			MspID:        "testorg2",
			PeerNodes:    org2Peers,
			OrdererNodes: org2Orderers,
		},
	}
	kafkas := []string{
		"172.16.93.215:9092",
		"172.16.93.215:9093",
		"172.16.93.215:9094",
		"172.16.93.215:9095",
	}
	genGenesisBlockreq := &GenGenesisBlockRequest{
		Orgs:   orgs,
		Kafkas: kafkas,
	}

	data, err := json.Marshal(genGenesisBlockreq)
	if err != nil {
		t.Fatal(err)
	}

	wrt := bytes.NewBuffer(data)

	resp, err := http.Post("http://127.0.0.1:8080/gengenesisblock", "application/json", wrt)

	if err != nil {
		t.Fatal(err)
	}
	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(ret))
}

func TestIdentity(t *testing.T) {
	peers := []*ServiceNode{
		&ServiceNode{
			ID:               "peer0",
			Endpoint:         "172.16.93.215:56451",
			ExternalEndpoint: "172.16.93.215:56451",
			Public:           true,
		},
		&ServiceNode{
			ID:               "peer1",
			Endpoint:         "172.16.93.215:56551",
			ExternalEndpoint: "172.16.93.215:56551",
			Public:           true,
		},
	}

	orderers := []*ServiceNode{
		&ServiceNode{
			ID:               "orderer0",
			Endpoint:         "172.16.93.215:58050",
			ExternalEndpoint: "172.16.93.215:58050",
			Public:           true,
		},
	}
	orgs := []*OrgInfo{
		&OrgInfo{
			OrgName:      "testorg3",
			OrgMSP:       "testorg3",
			MspID:        "testorg3",
			PeerNodes:    peers,
			OrdererNodes: orderers,
		},
	}

	idreq := &IdentityRequest{
		Orgs: orgs,
	}

	data, err := json.Marshal(idreq)
	if err != nil {
		t.Fatal(err)
	}

	wrt := bytes.NewBuffer(data)
	fmt.Printf("wrt:%s", wrt)
	fmt.Printf("data:%s", data)
	resp, err := http.Post("http://127.0.0.1:8080/channel/identity", "application/json", wrt)

	if err != nil {
		t.Fatal(err)
	}
	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile("newOrgIdentity", ret, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAddOrg(t *testing.T) {
	id, err := ioutil.ReadFile("newOrgIdentity")
	if err != nil {
		t.Fatal(err)
	}

	channelName := "channel1"
	//operate org
	org1Peers := []*ServiceNode{
		&ServiceNode{
			ID:               "peer0",
			Endpoint:         "172.16.93.215:56051",
			ExternalEndpoint: "172.16.93.215:56051",
			Public:           true,
		},
		&ServiceNode{
			ID:               "peer1",
			Endpoint:         "172.16.93.215:56151",
			ExternalEndpoint: "172.16.93.215:56151",
			Public:           true,
		},
	}
	org1Orderers := []*ServiceNode{
		&ServiceNode{
			ID:               "orderer0",
			Endpoint:         "172.16.93.215:56050",
			ExternalEndpoint: "172.16.93.215:56050",
			Public:           true,
		},
	}

	org2Peers := []*ServiceNode{
		&ServiceNode{
			ID:               "peer0",
			Endpoint:         "172.16.93.215:56251",
			ExternalEndpoint: "172.16.93.215:56251",
			Public:           true,
		},
		&ServiceNode{
			ID:               "peer1",
			Endpoint:         "172.16.93.215:56351",
			ExternalEndpoint: "172.16.93.215:56351",
			Public:           true,
		},
	}
	org2Orderers := []*ServiceNode{
		&ServiceNode{
			ID:               "orderer0",
			Endpoint:         "172.16.93.215:57050",
			ExternalEndpoint: "172.16.93.215:57050",
			Public:           true,
		},
	}

	orgs := []*OrgInfo{
		&OrgInfo{
			OrgName:      "testorg1",
			OrgMSP:       "testorg1",
			MspID:        "testorg1",
			PeerNodes:    org1Peers,
			OrdererNodes: org1Orderers,
		},
		&OrgInfo{
			OrgName:      "testorg2",
			OrgMSP:       "testorg2",
			MspID:        "testorg2",
			PeerNodes:    org2Peers,
			OrdererNodes: org2Orderers,
		},
	}

	addOrgReq := &AddOrgRequest{}
	addOrgReq.Identity = id
	addOrgReq.ChannelName = channelName
	addOrgReq.Orgs = orgs
	data, err := json.Marshal(addOrgReq)
	if err != nil {
		t.Fatal(err)
	}

	wrt := bytes.NewBuffer(data)

	resp, err := http.Post("http://127.0.0.1:8080/channel/addorg", "application/json", wrt)
	if err != nil {
		t.Fatal(err)
	}

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(ret))

}

func TestDeleteOrg(t *testing.T) {
	channelName := "channel1"
	//operate org
	org1Peers := []*ServiceNode{
		&ServiceNode{
			ID:               "peer0",
			Endpoint:         "172.16.93.215:56051",
			ExternalEndpoint: "172.16.93.215:56051",
			Public:           true,
		},
		&ServiceNode{
			ID:               "peer1",
			Endpoint:         "172.16.93.215:56151",
			ExternalEndpoint: "172.16.93.215:56151",
			Public:           true,
		},
	}
	org1Orderers := []*ServiceNode{
		&ServiceNode{
			ID:               "orderer0",
			Endpoint:         "172.16.93.215:56050",
			ExternalEndpoint: "172.16.93.215:56050",
			Public:           true,
		},
	}

	org2Peers := []*ServiceNode{
		&ServiceNode{
			ID:               "peer0",
			Endpoint:         "172.16.93.215:56251",
			ExternalEndpoint: "172.16.93.215:56251",
			Public:           true,
		},
		&ServiceNode{
			ID:               "peer1",
			Endpoint:         "172.16.93.215:56351",
			ExternalEndpoint: "172.16.93.215:56351",
			Public:           true,
		},
	}
	org2Orderers := []*ServiceNode{
		&ServiceNode{
			ID:               "orderer0",
			Endpoint:         "172.16.93.215:57050",
			ExternalEndpoint: "172.16.93.215:57050",
			Public:           true,
		},
	}

	orgs := []*OrgInfo{
		&OrgInfo{
			OrgName:      "testorg1",
			OrgMSP:       "testorg1",
			MspID:        "testorg1",
			PeerNodes:    org1Peers,
			OrdererNodes: org1Orderers,
		},
		&OrgInfo{
			OrgName:      "testorg2",
			OrgMSP:       "testorg2",
			MspID:        "testorg2",
			PeerNodes:    org2Peers,
			OrdererNodes: org2Orderers,
		},
	}

	delOrgReq := DeleteOrgRequest{}
	delOrgReq.ChannelName = channelName
	delOrgReq.Orgs = orgs
	delOrgReq.DelOrg = "testorg3"
	delOrgReq.DelOrderers = []string{"172.16.93.215:58050"}
	data, err := json.Marshal(delOrgReq)
	if err != nil {
		t.Fatal(err)
	}

	wrt := bytes.NewBuffer(data)

	resp, err := http.Post("http://127.0.0.1:8080/channel/deleteorg", "application/json", wrt)
	if err != nil {
		t.Fatal(err)
	}

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(ret))
}

func TestCreateChannel(t *testing.T) {
	org1Peers := []*ServiceNode{
		&ServiceNode{
			ID:               "peer0",
			Endpoint:         "172.16.93.215:56051",
			ExternalEndpoint: "172.16.93.215:56051",
			Public:           true,
		},
		&ServiceNode{
			ID:               "peer1",
			Endpoint:         "172.16.93.215:56151",
			ExternalEndpoint: "172.16.93.215:56151",
			Public:           true,
		},
	}
	org1Orderers := []*ServiceNode{
		&ServiceNode{
			ID:               "orderer0",
			Endpoint:         "172.16.93.215:56050",
			ExternalEndpoint: "172.16.93.215:56050",
			Public:           true,
		},
	}

	org2Peers := []*ServiceNode{
		&ServiceNode{
			ID:               "peer0",
			Endpoint:         "172.16.93.215:56251",
			ExternalEndpoint: "172.16.93.215:56251",
			Public:           true,
		},
		&ServiceNode{
			ID:               "peer1",
			Endpoint:         "172.16.93.215:56351",
			ExternalEndpoint: "172.16.93.215:56351",
			Public:           true,
		},
	}
	org2Orderers := []*ServiceNode{
		&ServiceNode{
			ID:               "orderer0",
			Endpoint:         "172.16.93.215:57050",
			ExternalEndpoint: "172.16.93.215:57050",
			Public:           true,
		},
	}

	orgs := []*OrgInfo{
		&OrgInfo{
			OrgName:      "testorg1",
			OrgMSP:       "testorg1",
			MspID:        "testorg1",
			PeerNodes:    org1Peers,
			OrdererNodes: org1Orderers,
		},
		&OrgInfo{
			OrgName:      "testorg2",
			OrgMSP:       "testorg2",
			MspID:        "testorg2",
			PeerNodes:    org2Peers,
			OrdererNodes: org2Orderers,
		},
	}
	channelName := "channel1"
	ccr := &NewCreateChannelRequest{
		Orgs:        orgs,
		ChannelName: channelName,
	}

	data, err := json.Marshal(ccr)
	if err != nil {
		t.Fatal(err)
	}
	wrt := bytes.NewBuffer(data)

	resp, err := http.Post("http://127.0.0.1:8080/channel/create", "application/json", wrt)
	if err != nil {
		t.Fatal(err)
	}
	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(ret))
}

func TestJoinChannel(t *testing.T) {
	channelname := "channel1"
	org1Peers := []*ServiceNode{
		&ServiceNode{
			ID:               "peer0",
			Endpoint:         "172.16.93.215:56051",
			ExternalEndpoint: "172.16.93.215:56051",
			Public:           true,
		},
		&ServiceNode{
			ID:               "peer1",
			Endpoint:         "172.16.93.215:56151",
			ExternalEndpoint: "172.16.93.215:56151",
			Public:           true,
		},
	}
	org1Orderers := []*ServiceNode{
		&ServiceNode{
			ID:               "orderer0",
			Endpoint:         "172.16.93.215:56050",
			ExternalEndpoint: "172.16.93.215:56050",
			Public:           true,
		},
	}
	// org2Peers := []*ServiceNode{
	// 	&ServiceNode{
	// 		ID:               "peer0",
	// 		Endpoint:         "172.16.93.215:56251",
	// 		ExternalEndpoint: "172.16.93.215:56251",
	// 		Public:           true,
	// 	},
	// 	&ServiceNode{
	// 		ID:               "peer1",
	// 		Endpoint:         "172.16.93.215:56351",
	// 		ExternalEndpoint: "172.16.93.215:56351",
	// 		Public:           true,
	// 	},
	// }
	// org2Orderers := []*ServiceNode{
	// 	&ServiceNode{
	// 		ID:               "orderer0",
	// 		Endpoint:         "172.16.93.215:57050",
	// 		ExternalEndpoint: "172.16.93.215:57050",
	// 		Public:           true,
	// 	},
	// }

	// org3Peers := []*ServiceNode{
	// 	&ServiceNode{
	// 		ID:               "peer0",
	// 		Endpoint:         "172.16.93.215:56451",
	// 		ExternalEndpoint: "172.16.93.215:56451",
	// 		Public:           true,
	// 	},
	// 	&ServiceNode{
	// 		ID:               "peer1",
	// 		Endpoint:         "172.16.93.215:56551",
	// 		ExternalEndpoint: "172.16.93.215:56551",
	// 		Public:           true,
	// 	},
	// }
	// org3Orderers := []*ServiceNode{
	// 	&ServiceNode{
	// 		ID:               "orderer0",
	// 		Endpoint:         "172.16.93.215:58050",
	// 		ExternalEndpoint: "172.16.93.215:58050",
	// 		Public:           true,
	// 	},
	// }

	orgs := []*OrgInfo{
		&OrgInfo{
			OrgName:      "testorg1",
			OrgMSP:       "testorg1",
			MspID:        "testorg1",
			PeerNodes:    org1Peers,
			OrdererNodes: org1Orderers,
		},
		// &OrgInfo{
		// 	OrgName:      "testorg2",
		// 	OrgMSP:       "testorg2",
		// 	MspID:        "testorg2",
		// 	PeerNodes:    org2Peers,
		// 	OrdererNodes: org2Orderers,
		// },
		// &OrgInfo{
		// 	OrgName:      "testorg3",
		// 	OrgMSP:       "testorg3",
		// 	MspID:        "testorg3",
		// 	PeerNodes:    org3Peers,
		// 	OrdererNodes: org3Orderers,
		// },
	}

	jcr := &JoinChannelRequest{
		Orgs:        orgs,
		ChannelName: channelname,
	}

	data, err := json.Marshal(jcr)
	if err != nil {
		t.Fatal(err)
	}
	wrt := bytes.NewBuffer(data)

	resp, err := http.Post("http://127.0.0.1:8080/channel/join", "application/json", wrt)
	if err != nil {
		t.Fatal(err)
	}
	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(ret))
}
