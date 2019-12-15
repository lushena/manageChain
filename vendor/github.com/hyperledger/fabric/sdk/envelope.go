package sdk

import (
	"bytes"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/common/crypto"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/protos/common"
	ab "github.com/hyperledger/fabric/protos/orderer"
	pp "github.com/hyperledger/fabric/protos/peer"
	putils "github.com/hyperledger/fabric/protos/utils"
	"github.com/pkg/errors"
)

// CreateConfigUpdateEnvelopeBytes ...
// Returns: configUpdate, signatureHeader, toSignBytes, error
func CreateConfigUpdateEnvelopeBytes(creator []byte, channelConfigUpdate *common.ConfigUpdate) ([]byte, []byte, []byte, error) {
	configUpdate := putils.MarshalOrPanic(channelConfigUpdate)

	sigHeader, err := newSignatureHeaderWithCreator(creator)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "creating signature header failed")
	}

	signatureHeader := putils.MarshalOrPanic(sigHeader)

	toSignBytes := util.ConcatenateBytes(signatureHeader, configUpdate)

	return configUpdate, signatureHeader, toSignBytes, nil
}

// CreateDeliverEnvelopeBytes ...
func CreateDeliverEnvelopeBytes(chainID string, info *ab.SeekInfo, creator []byte) ([]byte, error) {
	payloadChannelHeader := putils.MakeChannelHeader(common.HeaderType_DELIVER_SEEK_INFO, int32(0), chainID, uint64(0))
	payloadSignatureHeader, err := newSignatureHeaderWithCreator(creator)
	if err != nil {
		logger.Error("Error creating signatureHeader", err)
		return nil, err
	}
	data, err := proto.Marshal(info)
	if err != nil {
		logger.Error("Error marshaling data", err)
		return nil, err
	}

	paylBytes := putils.MarshalOrPanic(&common.Payload{
		Header: putils.MakePayloadHeader(payloadChannelHeader, payloadSignatureHeader),
		Data:   data,
	})

	return paylBytes, nil
}

// CreateChannelEnvelopeBytes ...
func CreateChannelEnvelopeBytes(chainID string, creator []byte, configUpdate []byte, sigs []*common.ConfigSignature) ([]byte, error) {
	newConfigUpdateEnv := &common.ConfigUpdateEnvelope{
		ConfigUpdate: configUpdate,
		Signatures:   sigs,
	}
	payloadChannelHeader := putils.MakeChannelHeader(common.HeaderType_CONFIG_UPDATE, 0, chainID, 0)
	payloadSignatureHeader, err := newSignatureHeaderWithCreator(creator)
	if err != nil {
		return nil, err
	}

	data, err := proto.Marshal(newConfigUpdateEnv)
	if err != nil {
		return nil, err
	}

	paylBytes := putils.MarshalOrPanic(&common.Payload{
		Header: putils.MakePayloadHeader(payloadChannelHeader, payloadSignatureHeader),
		Data:   data,
	})

	return paylBytes, nil

}

// CreateChaincodeEnvelopeBytesFromBytes ...
func CreateChaincodeEnvelopeBytesFromBytes(proposalBytes []byte, responses [][]byte) ([]byte, error) {
	proposal := &pp.Proposal{}
	err := proto.Unmarshal(proposalBytes, proposal)
	if err != nil {
		logger.Error("Error unmarshing proposal", err)
		return nil, err
	}

	resps := []*pp.ProposalResponse{}
	for _, respBytes := range responses {
		resp := &pp.ProposalResponse{}
		err := proto.Unmarshal(respBytes, resp)
		if err != nil {
			logger.Error("Error unmarshaling proposalResponse", err)
			return nil, err
		}
		resps = append(resps, resp)
	}
	return CreateChaincodeEnvelopeBytes(proposal, resps)

}

// CreateChaincodeEnvelopeBytes ...
func CreateChaincodeEnvelopeBytes(proposal *pp.Proposal, resps []*pp.ProposalResponse) ([]byte, error) {
	if len(resps) == 0 {
		return nil, fmt.Errorf("At least one proposal response is necessary")
	}

	// the original header
	hdr, err := putils.GetHeader(proposal.Header)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal the proposal header")
	}

	// the original payload
	pPayl, err := putils.GetChaincodeProposalPayload(proposal.Payload)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal the proposal payload")
	}

	// get header extensions so we have the visibility field
	hdrExt, err := putils.GetChaincodeHeaderExtension(hdr)
	if err != nil {
		return nil, err
	}

	// ensure that all actions are bitwise equal and that they are successful
	var a1 []byte
	for n, r := range resps {
		if n == 0 {
			a1 = r.Payload
			if r.Response.Status != 200 {
				return nil, fmt.Errorf("Proposal response was not successful, error code %d, msg %s", r.Response.Status, r.Response.Message)
			}
			continue
		}

		if bytes.Compare(a1, r.Payload) != 0 {
			return nil, fmt.Errorf("ProposalResponsePayloads do not match")
		}
	}

	// fill endorsements
	endorsements := make([]*pp.Endorsement, len(resps))
	for n, r := range resps {
		endorsements[n] = r.Endorsement
	}

	// create ChaincodeEndorsedAction
	cea := &pp.ChaincodeEndorsedAction{ProposalResponsePayload: resps[0].Payload, Endorsements: endorsements}

	// obtain the bytes of the proposal payload that will go to the transaction
	propPayloadBytes, err := putils.GetBytesProposalPayloadForTx(pPayl, hdrExt.PayloadVisibility)
	if err != nil {
		return nil, err
	}

	// serialize the chaincode action payload
	cap := &pp.ChaincodeActionPayload{ChaincodeProposalPayload: propPayloadBytes, Action: cea}
	capBytes, err := putils.GetBytesChaincodeActionPayload(cap)
	if err != nil {
		return nil, err
	}

	// create a transaction
	taa := &pp.TransactionAction{Header: hdr.SignatureHeader, Payload: capBytes}
	taas := make([]*pp.TransactionAction, 1)
	taas[0] = taa
	tx := &pp.Transaction{Actions: taas}

	// serialize the tx
	txBytes, err := putils.GetBytesTransaction(tx)
	if err != nil {
		return nil, err
	}

	// create the payload
	payl := &common.Payload{Header: hdr, Data: txBytes}
	return putils.GetBytesPayload(payl)
}

func newSignatureHeaderWithCreator(creator []byte) (*common.SignatureHeader, error) {
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		logger.Error("Error creating random nonce", err)
		return nil, err
	}
	return &common.SignatureHeader{
		Creator: creator,
		Nonce:   nonce,
	}, nil
}
