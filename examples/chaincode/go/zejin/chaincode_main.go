package main

import (
	// "bytes"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"reflect"
	// "math"
	// "strconv"
)

// ExampleChaincode implements a simple chaincode
type ExampleChaincode struct {
}

// Data .
type Data struct {
	ID           string `json:"id"`
	EbillCode    string `json:"ebill_code"`
	Amount       string `json:"amount"`
	AvailableAmt string `json:"available_amt"`
	WriterID     string `json:"writer_id"`
	HolderID     string `json:"holder_id"`
	OpenDate     string `json:"open_date"`
	DueDate      string `json:"due_date"`
	State        string `json:"state"`
}

var logger = shim.NewLogger("Example_Chaincode")

// Init .
func (t *ExampleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Infof("=========ExampleChaincode Init===========")

	return shim.Success(nil)
}

// Invoke to Update or Query DispatchWeight.
func (t *ExampleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Infof("=========ExampleChaincode Invoke===========")
	fn, args := stub.GetFunctionAndParameters()

	switch fn {
	case "CreateData":
		if len(args) != 1 {
			errstr := `{"Error":" args num should be 1"}`
			return ProcessErr(stub, errstr, logger)
		}
		return t.CreateData(stub, args[0])
	case "QueryDataByID":
		if len(args) != 1 {
			errstr := `{"Error":" args num should be 1"}`
			return ProcessErr(stub, errstr, logger)
		}
		return t.QueryDataByID(stub, args[0])
	case "QueryBlockByNum":
		if len(args) != 2 {
			errstr := `{"Error":" args num should be 2"}`
			return ProcessErr(stub, errstr, logger)
		}
		channelName := args[0]
		blockNum := args[1]
		return t.QueryBlockByNum(stub, channelName, blockNum)
	case "DeleteDataByID":
		if len(args) != 1 {
			errstr := `{"Error":" args num should be 1"}`
			return ProcessErr(stub, errstr, logger)
		}
		return t.DeleteDataByID(stub, args[0])
	case "Delete":
		return t.Delete(stub)
	case "StartContainer":
		return t.StartContainer(stub)
	}

	return shim.Error("Invalid invoke function name. Expecting ")
}

// CreateData 行方修改派案权重.
//args[0] Data
func (t *ExampleChaincode) CreateData(stub shim.ChaincodeStubInterface, DataJSON string) pb.Response {
	logger.Infof("===========CreateData===========")
	data := Data{}
	err := json.Unmarshal([]byte(DataJSON), &data)
	if err != nil {
		errstr := fmt.Sprintf("Error Unmarshal DataJSON: %s, err: %s", DataJSON, err.Error())
		return ProcessErr(stub, errstr, logger)
	}
	Value := reflect.ValueOf(data)
	Type := reflect.TypeOf(data)
	for i := 0; i < Type.NumField(); i++ {
		if _, ok := Value.Field(i).Interface().(string); !ok {
			errstr := fmt.Sprintf("Error %s type is not string", Type.Field(i).Name)
			return ProcessErr(stub, errstr, logger)
		}
		if len(Value.Field(i).String()) == 0 {
			errstr := fmt.Sprintf("Error %s should not be empty", Type.Field(i).Name)
			return ProcessErr(stub, errstr, logger)
		}
	}
	logger.Infof("Input Data: %v", data)
	err = stub.PutState(data.ID, []byte(DataJSON))
	if err != nil {
		errstr := fmt.Sprintf("Error PutState key: %s, DataJSON: %s, err: %s", data.ID, DataJSON, err.Error())
		return ProcessErr(stub, errstr, logger)
	}

	return shim.Success(nil)
}

// QueryDataByID .
// args[0] ID
func (t *ExampleChaincode) QueryDataByID(stub shim.ChaincodeStubInterface, ID string) pb.Response {
	logger.Infof("===========QueryDataByID===========")
	logger.Infof("Input: %s", ID)
	value, err := stub.GetState(ID)
	if err != nil {
		errstr := fmt.Sprintf("Error GetState key: %s, err: %s", ID, err.Error())
		return ProcessErr(stub, errstr, logger)
	}
	logger.Infof("QueryResult: %s", value)
	return shim.Success(value)
}

// DeleteDataByID 行方修改派案权重.
//args[0] Data
func (t *ExampleChaincode) DeleteDataByID(stub shim.ChaincodeStubInterface, ID string) pb.Response {
	logger.Infof("===========DeleteDataByID===========")
	err := stub.DelState(ID)
	if err != nil {
		errstr := fmt.Sprintf("Error DelState key: %s, err: %s", ID, err.Error())
		return ProcessErr(stub, errstr, logger)
	}
	return shim.Success(nil)
}

// QueryBlockByNum .
// args[0] channelName
// args[1] blockNum
func (t *ExampleChaincode) QueryBlockByNum(stub shim.ChaincodeStubInterface, channelName string, blockNum string) pb.Response {
	logger.Infof("===========QueryBlockByNum===========")
	logger.Infof("Input: channelName: %s, blockNum: %s", channelName, blockNum)
	invokeArgs := ToChaincodeArgs("GetBlockByNumber", channelName, blockNum)
	response := stub.InvokeChaincode("qscc", invokeArgs, "")
	if response.Status != shim.OK {
		errstr := fmt.Sprintf("err QueryBlockByNum response: %s", string(response.Payload))
		return ProcessErr(stub, errstr, logger)
	}
	value := response.Payload
	// JSON := Decode("common.Block", value)
	logger.Infof("QueryResult: %s", value)
	return shim.Success(value)
}

//ToChaincodeArgs converts string args to []byte args.
func ToChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

//Delete .
func (t ExampleChaincode) Delete(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Infof("===========Delete===========")
	resultsIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return ProcessErr(stub, err.Error(), logger)
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return ProcessErr(stub, err.Error(), logger)
		}
		logger.Info("Delete Key : %s", queryResponse.Key)
		err = stub.DelState(queryResponse.Key)
		if err != nil {
			return ProcessErr(stub, err.Error(), logger)
		}
	}
	// resultsIterator, err = stub.GetStateByPartialCompositeKey("packorg", []string{})
	// if err != nil {
	// 	return ProcessErr(stub, err.Error(), logger)
	// }
	// defer resultsIterator.Close()

	// // buffer is a JSON array containing QueryResults

	// for resultsIterator.HasNext() {
	// 	queryResponse, err := resultsIterator.Next()
	// 	fmt.Println(queryResponse)

	// 	if err != nil {
	// 		return ProcessErr(stub, err.Error(), logger)
	// 	}
	// 	logger.Infof("Delete Key : %s", queryResponse.Key)
	// 	err = stub.DelState(queryResponse.Key)
	// 	if err != nil {
	// 		return ProcessErr(stub, err.Error(), logger)
	// 	}
	// }
	logger.Infof("Delete all data")
	return shim.Success(nil)
}

// StartContainer .
func (t ExampleChaincode) StartContainer(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("===========startContainer starting===========")
	logger.Info("===========startContainer end===========")
	return shim.Success(nil)
}

// ProcessErr .
func ProcessErr(stub shim.ChaincodeStubInterface, errstr string, logger *shim.ChaincodeLogger) pb.Response {
	// errJsonStr := BuildDbJson(errCode, errstr,"")
	logger.Debug("failure, return :", errstr)
	fmt.Println("failure, return :", errstr)
	response := shim.Error(errstr)
	return response
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(ExampleChaincode)); err != nil {
		logger.Infof("Error starting ExampleChaincode chaincode: %s", err)
	}
}
