package chaincode

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestInstallChaincode(t *testing.T) {
	org := "testorg1"
	ccTarPath := "chaincodefile/example.tar.gz"
	ccPath := "example_cc"
	ccName := "mycc"
	ccVersion := "1.0"
	peernodes := []*ServiceNode{
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

	icr := &InstallChaincodeRequest{
		Org:       org,
		CcTarPath: ccTarPath,
		CcPath:    ccPath,
		CcName:    ccName,
		CcVersion: ccVersion,
		PeerNodes: peernodes,
	}

	data, err := json.Marshal(icr)
	if err != nil {
		t.Fatal(err)
	}
	wrt := bytes.NewBuffer(data)

	resp, err := http.Post("http://127.0.0.1:8080/chaincode/install", "application/json", wrt)
	if err != nil {
		t.Fatal(err)
	}

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(ret))
}

func TestInstantiateChaincode(t *testing.T) {
	org := "testorg1"
	channelName := "channel1"
	ccName := "mycc"
	ccVersion := "1.0"
	policy := acceptAllPolicy
	args := [][]byte{
		[]byte("init"),
		[]byte("a"),
		[]byte("1000000"),
		[]byte("b"),
		[]byte("1000000"),
	}
	peernodes := []*ServiceNode{
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

	ordernodes := []*ServiceNode{
		&ServiceNode{
			ID:               "orderer0",
			Endpoint:         "172.16.93.215:56050",
			ExternalEndpoint: "172.16.93.215:56050",
			Public:           true,
		},
	}

	icr := &InstantiateChaincodeRequest{
		Org:          org,
		CcName:       ccName,
		CcVersion:    ccVersion,
		Policy:       policy,
		Args:         args,
		ChannelName:  channelName,
		PeerNodes:    peernodes,
		OrdererNodes: ordernodes,
	}

	data, err := json.Marshal(icr)
	if err != nil {
		t.Fatal(err)
	}
	wrt := bytes.NewBuffer(data)

	resp, err := http.Post("http://127.0.0.1:8080/chaincode/instantiate", "application/json", wrt)
	if err != nil {
		t.Fatal(err)
	}

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(ret))
}

func TestMoveInvoke(t *testing.T) {
	org := "testorg1"
	channelName := "channel1"
	ccName := "mycc"
	args := [][]byte{
		[]byte("invoke"),
		[]byte("move"),
		[]byte("a"),
		[]byte("b"),
		[]byte("10"),
	}
	peernodes := []*ServiceNode{
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

	ordernodes := []*ServiceNode{
		&ServiceNode{
			ID:               "orderer0",
			Endpoint:         "172.16.93.215:56050",
			ExternalEndpoint: "172.16.93.215:56050",
			Public:           true,
		},
	}

	icr := &InvokeRequest{
		Org:         org,
		CcName:      ccName,
		ChannelName: channelName,
		Args:        args,

		PeerNodes:    peernodes,
		OrdererNodes: ordernodes,
	}

	data, err := json.Marshal(icr)
	if err != nil {
		t.Fatal(err)
	}
	wrt := bytes.NewBuffer(data)

	resp, err := http.Post("http://127.0.0.1:8080/chaincode/invoke", "application/json", wrt)
	if err != nil {
		t.Fatal(err)
	}

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(ret))
}

func TestQueryInvoke(t *testing.T) {
	org := "testorg3"
	channelName := "channel1"
	ccName := "mycc"
	args := [][]byte{
		[]byte("invoke"),
		[]byte("query"),
		[]byte("a"),
	}
	peernodes := []*ServiceNode{
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

	ordernodes := []*ServiceNode{
		&ServiceNode{
			ID:               "orderer0",
			Endpoint:         "172.16.93.215:58050",
			ExternalEndpoint: "172.16.93.215:58050",
			Public:           true,
		},
	}

	icr := &InvokeRequest{
		Org:         org,
		CcName:      ccName,
		ChannelName: channelName,
		Args:        args,

		PeerNodes:    peernodes,
		OrdererNodes: ordernodes,
	}

	data, err := json.Marshal(icr)
	if err != nil {
		t.Fatal(err)
	}
	wrt := bytes.NewBuffer(data)

	resp, err := http.Post("http://127.0.0.1:8080/chaincode/invoke", "application/json", wrt)
	if err != nil {
		t.Fatal(err)
	}

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(ret))
}
