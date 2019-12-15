package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"testing"
)

func checkInit(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInit("1", args)
	if res.Status != shim.OK {
		fmt.Println("Init failed", string(res.Message))
		t.FailNow()
	}
}

func checkState(t *testing.T, stub *shim.MockStub, name string, value string) {
	bytes := stub.State[name]
	if bytes == nil {
		fmt.Println("State", name, "failed to get value")
		t.FailNow()
	}
	if string(bytes) != value {
		fmt.Println("State value", name, "was not", value, "as expected")
		t.FailNow()
	}
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

func checkInvoke(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInvoke("1", args)
	// fmt.Println(res)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.FailNow()
	}
}

func TestExample02_Invoke(t *testing.T) {
	scc := new(ExampleChaincode)
	stub := shim.NewMockStub("Chaincode", scc)

	// Init
	checkInit(t, stub, [][]byte{})

	// Invoke update
	// checkInvoke(t, stub, [][]byte{[]byte("CreateData"), []byte(`{"EbillCode":"0001","Amount":"00000001","AvailableAmt":"0.3"}`)})
	checkInvoke(t, stub, [][]byte{[]byte("CreateData"), []byte(`{"id":"3","ebill_code":"0003","amount":"00000001","available_amt":"0.3","writer_id":"writer_id","holder_id":"holder_id","open_date":"open_date","due_date":"due_date","state":"state"}`)})
	// checkInvoke(t, stub, [][]byte{[]byte("CreateData"), []byte(`{"id":"2","ebill_code":"0002","amount":"00000001","available_amt":"0.3","writer_id":"writer_id","holder_id":"holder_id","open_date":"open_date","due_date":"due_date","state":"state"}`)})
	// checkInvoke(t, stub, [][]byte{[]byte("CreateData"), []byte(`{"id":"1","ebill_code":"0001","amount":"00000001","available_amt":"0.3","writer_id":"writer_id","holder_id":"holder_id","open_date":"open_date","due_date":"due_date","state":"state"}`)})

	// Invoke query
	checkInvoke(t, stub, [][]byte{[]byte("QueryDataByID"), []byte("3")})
	// checkInvoke(t, stub, [][]byte{[]byte("QueryDataByID"), []byte("2")})
	// checkInvoke(t, stub, [][]byte{[]byte("QueryDataByID"), []byte("1")})
	// Invoke delete

	checkInvoke(t, stub, [][]byte{[]byte("DeleteDataByID"), []byte("0")})

	checkInvoke(t, stub, [][]byte{[]byte("QueryDataByID"), []byte("0")})
	// checkInvoke(t, stub, [][]byte{[]byte("Delete")})
	t.Error("a")
}
