package sdk

import (
	pcommon "github.com/hyperledger/fabric/protos/common"
	pp "github.com/hyperledger/fabric/protos/peer"
	putils "github.com/hyperledger/fabric/protos/utils"
)

// CreateChaincodeProposal ...
func CreateChaincodeProposal(chainID string, chaincode string, args [][]byte, transient map[string][]byte, creator []byte) (string, *pp.Proposal, error) {
	input := &pp.ChaincodeInput{
		Args: args,
	}

	spec := &pp.ChaincodeSpec{
		Type:        pp.ChaincodeSpec_GOLANG,
		ChaincodeId: &pp.ChaincodeID{Name: chaincode},
		Input:       input,
	}

	invocation := &pp.ChaincodeInvocationSpec{ChaincodeSpec: spec}

	txID := ""
	prop, txID, err := putils.CreateChaincodeProposalWithTxIDAndTransient(pcommon.HeaderType_ENDORSER_TRANSACTION, chainID, invocation, creator, txID, transient)
	if err != nil {
		logger.Error("Error creating proposal", err)
		return "", nil, err
	}
	return txID, prop, nil

}

// CreateChaincodeProposalBytes ...
func CreateChaincodeProposalBytes(chainID string, chaincode string, args [][]byte, transient map[string][]byte, creator []byte) (string, []byte, error) {
	txID, prop, err := CreateChaincodeProposal(chainID, chaincode, args, transient, creator)
	if err != nil {
		logger.Error("Error creating ChaincodeProposal", err)
		return "", nil, err
	}
	bytes, err := putils.GetBytesProposal(prop)
	if err != nil {
		logger.Error("Error getting bytes of proposal", err)
		return "", nil, err
	}
	return txID, bytes, nil
}
