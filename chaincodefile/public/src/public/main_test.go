package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	// "strconv"
	"testing"
)

func TestCCInvoke(t *testing.T) {
	publicCC := new(PublicChaincode)
	stub := shim.NewMockStub("public", publicCC)

	// Init
	checkInit(t, stub, [][]byte{})
}

func TestGetInvitationSignStatus(t *testing.T) {
	publicCC := new(PublicChaincode)
	stub := shim.NewMockStub("public", publicCC)

	checkInvoke(t, stub, [][]byte{[]byte("AddOrgInfo"), []byte("publicchain"), []byte("orgA"), []byte("infoA")})
	checkInvoke(t, stub, [][]byte{[]byte("AddOrgInfo"), []byte("publicchain"), []byte("orgD"), []byte("infoD")})
	res := checkInvoke(t, stub, [][]byte{[]byte("GetOrgInfo"), []byte("publicchain"), []byte("orgA")})
	checkInvoke(t, stub, [][]byte{[]byte("GetAllOrgInfo"), []byte("publicchain")})
	res = checkInvoke(t, stub, [][]byte{[]byte("GetAllOrgname"), []byte("publicchain")})
	if res.Status != shim.OK {
		t.Fatal(res)
	}
	orgnames := []string{}
	err := json.Unmarshal(res.Payload, &orgnames)
	if err != nil {
		fmt.Printf("%+v\n", res)
		t.Fatal(err)
	}
	fmt.Println("orgnames: \n", orgnames)
	checkInvoke(t, stub, [][]byte{[]byte("StartInvitation"), []byte("publicchain"), []byte("orgA"), []byte("orgB"), []byte("RawData")})
	res = checkInvoke(t, stub, [][]byte{[]byte("GetAllInvitation"), []byte("publicchain")})
	if res.Status != shim.OK {
		t.Fatal(res)
	}
	invitations := []Invitation{}
	err = json.Unmarshal(res.Payload, &invitations)
	if err != nil {
		fmt.Printf("%+v", res)
		t.Fatal(err)
	}
	fmt.Println(invitations)

	checkInvoke(t, stub, [][]byte{[]byte("SignInvitation"), []byte("publicchain"), []byte("orgA"), []byte("orgB"), []byte("orgA"), []byte("signatureA"), []byte("Accept")})
	// checkInvoke(t, stub, [][]byte{[]byte("SignInvitation"), []byte("publicchain"), []byte("orgA"), []byte("orgB"), []byte("orgA"), []byte("signatureA"), []byte("Reject")})

	res = checkInvoke(t, stub, [][]byte{[]byte("GetInvitationSignStatus"), []byte("publicchain"), []byte("orgA"), []byte("orgB")})
	if res.Status != shim.OK {
		t.Fatal(res)
	}
	signStatusList := []InvitationSignStatus{}
	err = json.Unmarshal(res.Payload, &signStatusList)
	if err != nil {
		fmt.Printf("%+v", res)
		t.Fatal(err)
	}
	fmt.Println(signStatusList)
	checkInvoke(t, stub, [][]byte{[]byte("ConfirmInvitation"), []byte("publicchain"), []byte("orgA"), []byte("orgB")})
	res = checkInvoke(t, stub, [][]byte{[]byte("GetAllInvitation"), []byte("publicchain")})
	if res.Status != shim.OK {
		t.Fatal(res)
	}
	invitations = []Invitation{}
	err = json.Unmarshal(res.Payload, &invitations)
	if err != nil {
		fmt.Printf("%+v", res)
		t.Fatal(err)
	}
	fmt.Println(invitations)
}

func checkInit(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInit("1", args)
	if res.Status != shim.OK {
		fmt.Println("Init failed", string(res.Message))
		t.FailNow()
	}
}

func checkInvoke(t *testing.T, stub *shim.MockStub, args [][]byte) pb.Response {
	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.FailNow()
	}
	return res
}

func checkQuery(t *testing.T, stub *shim.MockStub, name string, value string) {
	res := stub.MockInvoke("1", [][]byte{[]byte("query"), []byte(name)})
	if res.Status != shim.OK {
		fmt.Println("Query", name, "failed", string(res.Message))
		t.FailNow()
	}
	if res.Payload == nil {
		fmt.Println("Query", name, "failed to get value")
		t.FailNow()
	}
	if string(res.Payload) != value {
		fmt.Println("Query value", name, "was not", value, "as expected")
		t.FailNow()
	}
}
