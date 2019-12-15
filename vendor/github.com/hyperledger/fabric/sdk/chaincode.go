package sdk

import (
	"github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/peer/chaincode"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/pkg/errors"
)

var (
	defaultESCC = []byte("escc")
	defaultVSCC = []byte("vscc")
)

// InstantiateChaincode ...
func (client *Client) InstantiateChaincode(chainID string, name string, version string, input [][]byte, policy string, collection []byte, endorser *Endpoint, casters []*Endpoint) error {
	return instantiateChaincode(chainID, name, version, input, policy, collection, endorser, casters, client.signer)
}

func instantiateChaincode(chainID string, name string, version string, input [][]byte, policy string, collection []byte, endorser *Endpoint, casters []*Endpoint, signer msp.SigningIdentity) error {
	cds := createChaincodeDeploymentSpec(name, version, "", nil, input)
	creator, err := signer.Serialize()
	if err != nil {
		logger.Error("Error serializing", err)
		return err
	}

	// policy
	var policyBytes []byte
	if policy != "" {
		p, err := cauthdsl.FromString(policy)
		if err != nil {
			return errors.Errorf("invalid policy %s", policy)
		}
		policyBytes = utils.MarshalOrPanic(p)
	}

	// collection
	var collectionBytes []byte
	if collection != nil {
		collectionBytes, err = chaincode.GetCollectionConfigFromBytes(collection)
		if err != nil {
			return errors.Errorf("get collection config from bytes error: %s", err)
		}
	}

	prop, _, err := utils.CreateDeployProposalFromCDS(chainID, cds, creator, policyBytes, defaultESCC, defaultVSCC, collectionBytes)
	if err != nil {
		logger.Error("Error creating deployProposal", err)
		return err
	}
	propBytes, err := utils.GetBytesProposal(prop)
	if err != nil {
		logger.Error("Error marshaling proposal", err)
		return err
	}

	sig, err := signer.Sign(propBytes)
	if err != nil {
		logger.Error("Error signning proposal", err)
		return err
	}

	resps, err := Endorse(propBytes, sig, []*Endpoint{endorser})
	if err != nil {
		logger.Error("Error endorsing", err)
		return err
	}
	payload, err := CreateChaincodeEnvelopeBytes(prop, resps)
	if err != nil {
		logger.Error("Error creating ChaincodeEnvelopeBytes", err)
		return err

	}
	signature, err := signer.Sign(payload)
	if err != nil {
		logger.Error("Error signning payload", err)
		return err
	}
	for _, caster := range casters {
		if err = Broadcast(payload, signature, caster); err == nil {
			return nil
		}
		logger.Error("Error broadcasting", err)
	}
	return errors.New("failed broadcasting after try all orderers")
}

// InstallChaincode ...
func (client *Client) InstallChaincode(name string, version string, ccPath string, code []byte, endorsers []*Endpoint) error {
	return installChaincode(name, version, ccPath, code, endorsers, client.signer)
}

func installChaincode(name string, version string, ccPath string, code []byte, endorsers []*Endpoint, signer msp.SigningIdentity) error {
	cds := createChaincodeDeploymentSpec(name, version, ccPath, code, nil)
	creator, err := signer.Serialize()
	if err != nil {
		logger.Error("Error serializing", err)
		return err
	}
	prop, _, err := utils.CreateInstallProposalFromCDS(cds, creator)
	if err != nil {
		logger.Error("Error creating installProposal", err)
		return err
	}

	propBytes, err := utils.GetBytesProposal(prop)
	if err != nil {
		logger.Error("Error marshaling proposal", err)
		return err
	}

	sig, err := signer.Sign(propBytes)
	if err != nil {
		logger.Error("Error signning proposal", err)
		return err
	}

	_, err = Endorse(propBytes, sig, endorsers)
	return err
}

func createChaincodeDeploymentSpec(name string, version string, ccPath string, code []byte, input [][]byte) *pb.ChaincodeDeploymentSpec {
	spec := &pb.ChaincodeSpec{
		ChaincodeId: &pb.ChaincodeID{
			Path:    ccPath,
			Name:    name,
			Version: version,
		},
		Input: &pb.ChaincodeInput{
			Args: input,
		},
		Type: pb.ChaincodeSpec_GOLANG,
	}

	return &pb.ChaincodeDeploymentSpec{
		ChaincodeSpec: spec,
		CodePackage:   code,
	}
}
