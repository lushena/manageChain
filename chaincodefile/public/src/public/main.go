/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const (
	orgPrefix        = "Org"
	invitePrefix     = "Invite"
	updateDataPrefix = "UpdateData"
	signedPrefix     = "Signed"
	channelOrgPrefix = "ChannelOrg"
)

const (
	addOrgInfo    = "AddOrgInfo"
	getOrgInfo    = "GetOrgInfo"
	getAllOrgInfo = "GetAllOrgInfo"
	updateOrgInfo = "UpdateOrgInfo"
	getAllOrgname = "GetAllOrgname"

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
	initState    = "init"
	confirmState = "confirm"

	acceptState = "Accept"
	rejectState = "Reject"
)

type OrgInfo struct {
	ChainId string `json:"chainId"`
	Orgname string `json:"orgname"`
	Info    string `json:"info"`
}

type Invitation struct {
	ChainId    string `json:"chainId"`
	Inviter    string `json:"inviter"`
	Invitee    string `json:"invitee"`
	Status     string `json:"status"`
	InviteTime int64  `json:"inviteTime"`
	RawData    string `json:"rawdata"`
}

type InvitationSignStatus struct {
	ChainId   string `json:"chainId"`
	Inviter   string `json:"inviter"`
	Invitee   string `json:"invitee"`
	Signer    string `json:"signer"`
	Signature string `json:"signature"`
	Accepted  string `json:"accepted"`
	SignTime  int64  `json:"signTime"`
}

type PublicChaincode struct {
}

var logger = shim.NewLogger("PublicChaincode")

func (t *PublicChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("====================Init===================")
	return shim.Success([]byte("Successfully init"))
}

func (t *PublicChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	switch function {
	case addOrgInfo:
		if len(args) != 3 {
			return shim.Error("AddOrgInfo must include three arguments: [chainId, orgname, info]")
		}
		chainId := args[0]
		orgname := args[1]
		info := args[2]
		return t.AddOrgInfo(stub, chainId, orgname, info)
	case updateOrgInfo:
		if len(args) != 3 {
			return shim.Error("UpdateOrgInfo must include three arguments: [chainId, orgname, info]")
		}
		chainId := args[0]
		orgname := args[1]
		info := args[2]
		return t.UpdateOrgInfo(stub, chainId, orgname, info)
	case getOrgInfo:
		if len(args) != 2 {
			return shim.Error("GetOrgInfo must include two argument: [chainId, orgname]")
		}
		chainId := args[0]
		orgname := args[1]
		return t.GetOrgInfo(stub, chainId, orgname)
	case getAllOrgInfo:
		if len(args) != 1 {
			return shim.Error("GetAllOrgInfo must include one argument: chainId")
		}
		chainId := args[0]
		return t.GetAllOrgInfo(stub, chainId)
	case getAllOrgname:
		if len(args) != 1 {
			return shim.Error("GetOrgInfo must include one argument: chainId")
		}
		chainId := args[0]
		return t.GetAllOrgname(stub, chainId)
	case startInvitation:
		if len(args) != 4 {
			return shim.Error("StartInvitation must include four arguments, [chainId, inviter, invitee,rawData]")
		}
		chainId := args[0]
		inviter := args[1]
		invitee := args[2]
		rawData := args[3]
		return t.StartInvitation(stub, chainId, inviter, invitee, rawData)
	case confirmInvitation:
		if len(args) != 3 {
			return shim.Error("ConfirmInvitation must include three arguments, [chainId, inviter, invitee]")
		}
		chainId := args[0]
		inviter := args[1]
		invitee := args[2]

		return t.ConfirmInvitation(stub, chainId, inviter, invitee)
	case getInvitation:
		if len(args) != 3 {
			return shim.Error("GetInvitation info must include three arguments, [chainId, inviter, invitee]")
		}
		chainId := args[0]
		inviter := args[1]
		invitee := args[2]

		return t.GetInvitation(stub, chainId, inviter, invitee)
	case getAllInvitation:
		if len(args) != 1 {
			return shim.Error("GetAllInvitation info must include one argument, chainId")
		}
		chainId := args[0]
		return t.GetAllInvitation(stub, chainId)
	case signInvitation:
		if len(args) != 6 {
			return shim.Error("SignInvitation must include six arguments, [chainId, inviter, invitee, signer, accepted(Accept or Reject)]")
		}
		chainId := args[0]
		inviter := args[1]
		invitee := args[2]
		signer := args[3]
		signature := args[4]
		accepted := args[5]
		return t.SignInvitation(stub, chainId, inviter, invitee, signer, signature, accepted)
	case getInvitationSignStatus:
		if len(args) != 3 {
			return shim.Error("GetInvitationSignStatus must include three arguments, [chainId, inviter, invitee]")
		}
		chainId := args[0]
		inviter := args[1]
		invitee := args[2]
		return t.GetInvitationSignStatus(stub, chainId, inviter, invitee)
	case updateChainOrgInfo:
		if len(args) != 3 {
			return shim.Error("updateChainOrgInfo must include 3 arguments, ")
		}
		chainId := args[0]
		orgName := args[1]
		data := args[2]
		return t.UpdateChainOrgInfo(stub, chainId, orgName, data)
	case getChainOrgInfo:
		if len(args) != 2 {
			return shim.Error("getChainOrgInfo must include 2 arguments, ")
		}
		chainId := args[0]
		orgName := args[1]
		return t.GetChainOrgInfo(stub, chainId, orgName)
	case getAllOrgnameOfChain:
		if len(args) != 1 {
			return shim.Error("GetOrgInfo must include one argument: chainId")
		}
		chainId := args[0]
		return t.GetAllOrgnameOfChain(stub, chainId)
	default:
		return shim.Error("Unsupported operation")
	}
}
func (t *PublicChaincode) AddOrgInfo(stub shim.ChaincodeStubInterface, chainId, orgname, info string) pb.Response {
	logger.Infof("===============Start AddOrgInfo============, chainId: %s, orgname: %s, info: %s", chainId, orgname, info)
	key, err := stub.CreateCompositeKey(orgPrefix, []string{chainId, orgname})
	if err != nil {
		return shim.Error(fmt.Sprintf("Error creating composit key, err: %s, prefix: %s, []string: %s", err, orgPrefix, []string{chainId, orgname}))
	}

	value, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting state: %s", err))
	}
	if value != nil {
		return shim.Error("Org already exists, use `UpdateOrgInfo` to update it")
	}

	data := []byte(info)

	if err := stub.PutState(key, data); err != nil {
		return shim.Error(fmt.Sprintf("Error adding org info: %s", err))
	}
	logger.Infof("===============End AddOrgInfo============")
	return shim.Success([]byte("Successfully adding org info"))
}

func (t *PublicChaincode) UpdateOrgInfo(stub shim.ChaincodeStubInterface, chainId, orgname, info string) pb.Response {
	logger.Infof("===============UpdateOrgInfo============, chainId: %s, orgname: %s, info: %s", chainId, orgname, info)
	key, err := stub.CreateCompositeKey(orgPrefix, []string{chainId, orgname})
	if err != nil {
		return shim.Error(fmt.Sprintf("Error creating composit key, err: %s, prefix: %s, []string: %s", err, orgPrefix, []string{chainId, orgname}))
	}

	value, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting state: %s", err))
	}
	if value == nil {
		return shim.Error("Org not exists, use `AddOrgInfo` to add info first")
	}

	data := []byte(info)

	if err := stub.PutState(key, data); err != nil {
		return shim.Error(fmt.Sprintf("Error update org info: %s", err))
	}
	logger.Infof("===============End UpdateOrgInfo============")
	return shim.Success([]byte("Successfully updating org info"))
}

func (t *PublicChaincode) GetOrgInfo(stub shim.ChaincodeStubInterface, chainId, orgname string) pb.Response {
	logger.Infof("===============Start GetOrgInfo============, chainId: %s, orgname: %s", chainId, orgname)
	key, err := stub.CreateCompositeKey(orgPrefix, []string{chainId, orgname})
	if err != nil {
		return shim.Error(fmt.Sprintf("Error creating composit key, err: %s, prefix: %s, []string: %s", err, orgPrefix, []string{chainId, orgname}))
	}

	data, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting org info: %s", err))
	}
	logger.Infof("===============End GetOrgInfo============")
	return shim.Success(data)
}

func (t *PublicChaincode) GetAllOrgInfo(stub shim.ChaincodeStubInterface, chainId string) pb.Response {
	logger.Infof("===============Start GetAllOrgInfo============, chainId: %s", chainId)
	iter, err := stub.GetStateByPartialCompositeKey(orgPrefix, []string{chainId})
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting state by partial composit key: %s", err))
	}

	orgInfos := []OrgInfo{}
	for iter.HasNext() {
		k, err := iter.Next()
		if err != nil {
			return shim.Error(fmt.Sprintf("Error getting next state: %s", err))
		}
		_, partials, err := stub.SplitCompositeKey(k.Key)
		if err != nil {
			return shim.Error(fmt.Sprintf("Error splitting composit key: %s", err))
		}
		orgInfo := OrgInfo{}
		orgInfo.ChainId = partials[0]
		orgInfo.Orgname = partials[1]
		orgInfo.Info = string(k.Value)
		orgInfos = append(orgInfos, orgInfo)
	}

	logger.Infof("orgInfos: %+v\n", orgInfos)

	data, err := json.Marshal(orgInfos)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error marshaling: %s", err))
	}
	logger.Infof("===============End GetAllOrgInfo============")
	return shim.Success(data)
}

func (t *PublicChaincode) GetAllOrgname(stub shim.ChaincodeStubInterface, chainId string) pb.Response {
	logger.Infof("===============Start GetAllOrgname============, chainId: %s", chainId)
	iter, err := stub.GetStateByPartialCompositeKey(orgPrefix, []string{chainId})
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting state by partial composit key: %s", err))
	}

	orgnames := []string{}
	for iter.HasNext() {
		k, err := iter.Next()
		if err != nil {
			return shim.Error(fmt.Sprintf("Error getting next state: %s", err))
		}
		_, partials, err := stub.SplitCompositeKey(k.Key)
		if err != nil {
			return shim.Error(fmt.Sprintf("Error splitting composit key: %s", err))
		}
		orgname := partials[1]
		orgnames = append(orgnames, orgname)
	}

	logger.Infof("orgnames: %+v\n", orgnames)

	data, err := json.Marshal(orgnames)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error marshaling: %s", err))
	}
	logger.Infof("===============End GetAllOrgname============")
	return shim.Success(data)
}

func (t *PublicChaincode) GetAllOrgnameOfChain(stub shim.ChaincodeStubInterface, chainId string) pb.Response {
	logger.Infof("===============Start GetAllOrgnameOfChain============, chainId: %s", chainId)
	iter, err := stub.GetStateByPartialCompositeKey(channelOrgPrefix, []string{chainId})
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting state by partial composit key: %s", err))
	}

	orgnames := []string{}
	for iter.HasNext() {
		k, err := iter.Next()
		if err != nil {
			return shim.Error(fmt.Sprintf("Error getting next state: %s", err))
		}
		_, partials, err := stub.SplitCompositeKey(k.Key)
		if err != nil {
			return shim.Error(fmt.Sprintf("Error splitting composit key: %s", err))
		}
		orgname := partials[1]
		orgnames = append(orgnames, orgname)
	}

	logger.Infof("orgnames: %+v\n", orgnames)

	data, err := json.Marshal(orgnames)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error marshaling: %s", err))
	}
	logger.Infof("===============End GetAllOrgnameOfChain============")
	return shim.Success(data)
}

func (t *PublicChaincode) StartInvitation(stub shim.ChaincodeStubInterface, chainId, inviter, invitee, rawData string) pb.Response {
	logger.Infof("===============Start StartInvitation============, chainId: %s, inviter: %s, invitee: %s, rawData: %s", chainId, inviter, invitee, rawData)

	timestamp := time.Now().Unix()

	key, err := stub.CreateCompositeKey(invitePrefix, []string{chainId, inviter, invitee})
	if err != nil {
		return shim.Error(fmt.Sprintf("Error creating composit key: %s", err))
	}
	if !InviterExist(stub, "publicchain", inviter) {
		return shim.Error(fmt.Sprintf("Error please make sure inviter org already in chain: %s", chainId))
	}
	if InviteeExist(stub, chainId, invitee) {
		return shim.Error(fmt.Sprintf("Error invitee org is already in chain: %s", chainId))
	}

	if HasBeenInvited(stub, chainId, invitee) {
		return shim.Error(fmt.Sprintf("Error invitee is already been invited in chain: %s", chainId))
	}

	val, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting state: %s", err))
	}
	if val != nil {
		return shim.Error("invitation already exists")
	}

	invitation := Invitation{}

	invitation.ChainId = chainId
	invitation.Inviter = inviter
	invitation.Invitee = invitee
	invitation.Status = initState
	invitation.InviteTime = timestamp
	invitation.RawData = rawData

	bytes, err := json.Marshal(invitation)
	if err != nil {
		return shim.Error(fmt.Sprintf("Marshal invitation: %+v, err: %s", invitation, err))
	}

	err = stub.PutState(key, bytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error putting data: %s", err))
	}
	logger.Infof("===============End StartInvitation============")

	return shim.Success([]byte("Successfully starting invitation"))
}

func (t *PublicChaincode) ConfirmInvitation(stub shim.ChaincodeStubInterface, chainId, inviter, invitee string) pb.Response {
	logger.Infof("===============Start ConfirmInvitation============, chainId: %s, inviter: %s, invitee: %s", chainId, inviter, invitee)
	key, err := stub.CreateCompositeKey(invitePrefix, []string{chainId, inviter, invitee})
	if err != nil {
		return shim.Error(fmt.Sprintf("Error creating composit key: %s", err))
	}

	data, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting state: %s", err))
	}

	invitation := Invitation{}
	err = json.Unmarshal(data, &invitation)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error Unmarshal data: %s, err: %s", data, err))
	}

	if invitation.Status != initState {
		return shim.Error(fmt.Sprintf("invitation should be %s, got %s", initState, invitation.Status))
	}

	invitation.Status = confirmState

	bytes, err := json.Marshal(invitation)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error Unmarshal invitation: %+v, err: %s", invitation, err))
	}

	err = stub.PutState(key, bytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error putting data: %s", err))
	}
	logger.Infof("===============End ConfirmInvitation============")
	return shim.Success([]byte("Successfully confirming invitation"))
}

func (t *PublicChaincode) GetInvitation(stub shim.ChaincodeStubInterface, chainId, inviter, invitee string) pb.Response {
	logger.Infof("===============Start GetInvitation, chainId: %s, inviter: %s, invitee: %s============", chainId, inviter, invitee)
	key, err := stub.CreateCompositeKey(invitePrefix, []string{chainId, inviter, invitee})
	if err != nil {
		return shim.Error(fmt.Sprintf("Error creating composit key: %s", err))
	}
	data, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting state: %s", err))
	}
	logger.Infof("===============End GetInvitation============")
	return shim.Success(data)
}

func (t *PublicChaincode) GetAllInvitation(stub shim.ChaincodeStubInterface, chainId string) pb.Response {
	logger.Infof("===============Start GetAllInvitation============, chainId: %s", chainId)
	iter, err := stub.GetStateByPartialCompositeKey(invitePrefix, []string{chainId})
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting state by partial composit key: %s", err))
	}
	defer iter.Close()
	invaitations := []Invitation{}
	for iter.HasNext() {
		k, err := iter.Next()
		if err != nil {
			return shim.Error(fmt.Sprintf("Error getting next state: %s", err))
		}
		// _, partials, err := stub.SplitCompositeKey(k.Key)
		// if err != nil {
		// 	return shim.Error(fmt.Sprintf("Error splitting composit key: %s", err))
		// }

		invaitation := Invitation{}
		err = json.Unmarshal(k.Value, &invaitation)
		if err != nil {
			return shim.Error(fmt.Sprintf("Error Unmarshal k.Value: %s, err: %s", k.Value, err))
		}

		invaitations = append(invaitations, invaitation)
	}
	logger.Infof("invaitations: %+v\n", invaitations)
	data, err := json.Marshal(invaitations)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error marshaling: %s", err))
	}
	logger.Infof("===============End GetAllInvitation============")
	return shim.Success(data)
}

func (t *PublicChaincode) SignInvitation(stub shim.ChaincodeStubInterface, chainId, inviter, invitee, signer, signature, accepted string) pb.Response {
	logger.Infof("===============Start SignInvitation, chainId: %s, inviter: %s, invitee: %s, signer: %s, signature :%s, accepted: %s============", chainId, inviter, invitee, signer, signature, accepted)
	timestamp := time.Now().Unix()
	key, err := stub.CreateCompositeKey(signedPrefix, []string{chainId, inviter, invitee, signer})
	if err != nil {
		return shim.Error(fmt.Sprintf("Error creating composit key: %s", err))
	}

	signStateBytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting state: %s", err))
	}
	signState := InvitationSignStatus{}
	if signStateBytes != nil {
		err = json.Unmarshal(signStateBytes, &signState)
		if err != nil {
			return shim.Error(fmt.Sprintf("Error Unmarshal signState, err: %s", err))
		}
	}

	if signState.Accepted == acceptState || signState.Accepted == rejectState {
		return shim.Error(fmt.Sprintf("Error signer: %s has already %s to invite %s into chain %s from inviter: %s, cannot change anymore", signer, signState.Accepted, invitee, chainId, inviter))
	}

	signState.ChainId = chainId
	signState.Inviter = inviter
	signState.Invitee = invitee
	signState.Signer = signer
	signState.Signature = signature
	signState.Accepted = accepted
	signState.SignTime = timestamp

	bytes, err := json.Marshal(signState)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error Marshal signState: %+v, err: %s", signState, err))
	}

	err = stub.PutState(key, bytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error putting data: %s", err))
	}

	logger.Infof("===============End SignInvitation============")
	return shim.Success([]byte("Successfully signing invitation"))
}

func (t *PublicChaincode) GetInvitationSignStatus(stub shim.ChaincodeStubInterface, chainId, inviter, invitee string) pb.Response {
	logger.Infof("===============Start GetInvitationSignStatus, chainId: %s, inviter: %s, invitee: %s============", chainId, inviter, invitee)
	iter, err := stub.GetStateByPartialCompositeKey(signedPrefix, []string{chainId, inviter, invitee})
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting state by partial composit key: %s", err))
	}
	defer iter.Close()

	signStatusList := []InvitationSignStatus{}
	for iter.HasNext() {
		k, err := iter.Next()
		if err != nil {
			return shim.Error(fmt.Sprintf("Error getting next state: %s", err))
		}

		_, _, err = stub.SplitCompositeKey(k.Key)
		if err != nil {
			return shim.Error(fmt.Sprintf("Error splitting composit key: %s", err))
		}

		signStatus := InvitationSignStatus{}
		err = json.Unmarshal(k.Value, &signStatus)
		if err != nil {
			return shim.Error(fmt.Sprintf("Error Unmarshal k.Value: %s, err: %s", k.Value, err))
		}
		// signStatus.Inviter = partials[0]
		// signStatus.Invitee = partials[1]
		// signStatus.Signer = partials[2]

		signStatusList = append(signStatusList, signStatus)
	}

	logger.Infof("signStatusList: %+v\n", signStatusList)
	data, err := json.Marshal(signStatusList)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error marshaling: %s", err))
	}
	logger.Infof("===============End GetInvitationSignStatus============")
	return shim.Success(data)
}

func (t *PublicChaincode) UpdateChainOrgInfo(stub shim.ChaincodeStubInterface, chainId, orgName, data string) pb.Response {
	logger.Infof("start updateChainOrgInfo, chainid:%s, orgName:%s, data:%s", chainId, orgName, data)

	key, err := stub.CreateCompositeKey(channelOrgPrefix, []string{chainId, orgName})
	if err != nil {
		return shim.Error(fmt.Sprintf("Error creating composit key, err: %s, prefix: %s, []string: %s", err, orgPrefix, []string{chainId, orgName}))
	}

	err = stub.PutState(key, []byte(data))
	if err != nil {
		return shim.Error(fmt.Sprintf("Error updateChainOrgInfo: %s", err))
	}
	logger.Infof("end updateChainOrgInfo")
	return shim.Success([]byte("Successfully UpdateChainOrgInfo"))
}

func (t *PublicChaincode) GetChainOrgInfo(stub shim.ChaincodeStubInterface, chainId string, orgName string) pb.Response {
	logger.Infof("===============Start GetChainOrgInfo============, chainId: %s, orgName: %s", chainId, orgName)

	key, err := stub.CreateCompositeKey(channelOrgPrefix, []string{chainId, orgName})
	if err != nil {
		return shim.Error(fmt.Sprintf("Error creating composit key, err: %s, prefix: %s, []string: %s", err, orgPrefix, []string{chainId, orgName}))
	}

	data, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error getting chain org info: %s", err))
	}
	logger.Infof("get state data:%s", data)
	logger.Infof("===============End GetChainOrgInfo============")

	return shim.Success(data)
}

func InviterExist(stub shim.ChaincodeStubInterface, chainId, inviter string) bool {
	logger.Infof("===============Start InviterExist, chainId: %s, inviter: %s============", chainId, inviter)

	key, err := stub.CreateCompositeKey(orgPrefix, []string{chainId, inviter})
	if err != nil {
		logger.Errorf("Error creating composit key, err: %s", err)
		return false
	}

	value, err := stub.GetState(key)
	if err != nil {
		logger.Errorf("Error getting org info, err: %s", err)
		return false
	}
	if value == nil {
		return false
	}
	logger.Infof("===============End InviterExist============")
	return true
}

func InviteeExist(stub shim.ChaincodeStubInterface, chainId, invitee string) bool {
	logger.Infof("===============Start InviterExist, chainId: %s, invitee: %s============", chainId, invitee)

	key, err := stub.CreateCompositeKey(orgPrefix, []string{chainId, invitee})
	if err != nil {
		logger.Errorf("Error creating composit key, err: %s", err)
		return true
	}

	value, err := stub.GetState(key)
	if err != nil {
		logger.Errorf("Error getting org info, err: %s", err)
		return true
	}
	if value == nil {
		return false
	}
	logger.Infof("===============End InviterExist============")
	return true
}

func HasBeenInvited(stub shim.ChaincodeStubInterface, chainId, invitee string) bool {
	logger.Infof("===============Start InvitationExist, chainId: %s, invitee: %s============", chainId, invitee)
	iter, err := stub.GetStateByPartialCompositeKey(invitePrefix, []string{chainId})
	if err != nil {
		logger.Errorf("Error creating composit key, err: %s", err)
		return true
	}
	defer iter.Close()

	for iter.HasNext() {
		k, err := iter.Next()
		if err != nil {
			logger.Errorf("Error getting next state: %s", err)
			return true
		}
		_, partials, err := stub.SplitCompositeKey(k.Key)
		if err != nil {
			logger.Errorf("Error splitting composit key, err: %s", err)
			return true
		}
		// make sure, invitee has not been invited to this chain
		if invitee == partials[2] {
			return true
		}
	}
	logger.Infof("===============End InvitationExist============")
	return false
}

func main() {
	err := shim.Start(new(PublicChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
