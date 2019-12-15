package sdk

import (
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/msp"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

const (
	defaultWaitTimeout = 1 * time.Minute
)

var fnGetTransactionByID = []byte("GetTransactionByID")

// ValidTx returns whether this tx is valid or not by calling qscc
func (client *Client) ValidTx(chainID string, txID string, committer *Endpoint) (bool, error) {
	_, _, resps, err := client.Endorse(chainID, "qscc", [][]byte{fnGetTransactionByID, []byte(chainID), []byte(txID)}, nil, []*Endpoint{committer})
	if err != nil {
		logger.Error("Error calling GetTransactionByID in qscc", err)
		return false, err
	}
	resp := resps[0].Response
	if resp.Status != shim.OK {
		logger.Errorf("Error calling GetTransactionByID with message: %s", resp.Message)
		return false, errors.Errorf("response's status is not ok, message: %s", resp.Message)
	}

	tx := &pb.ProcessedTransaction{}
	err = proto.Unmarshal(resp.Payload, tx)
	if err != nil {
		logger.Error("Error unmarshaling ProcessedTransaction", err)
		return false, err
	}
	return tx.ValidationCode == int32(pb.TxValidationCode_VALID), nil
}

// WaitTx waits this tx to be processed by the committer
func (client *Client) WaitTx(chainID string, txID string, committer *Endpoint, timeout time.Duration) (bool, error) {
	return waitTx(chainID, txID, committer, client.signer, timeout)
}

// WaitTx returns whether this tx is valid or not and the error message
func waitTx(chainID string, txID string, committer *Endpoint, signer msp.SigningIdentity, timeout time.Duration) (bool, error) {
	iter, err := getNewCommittedFilteredBlocksByChannel(chainID, committer, signer)
	if err != nil {
		logger.Error("Error getting newly committed filtered blocks", err)
		return false, err
	}

	defer iter.Close()

	if timeout == time.Duration(0) {
		timeout = defaultTimeout
	}

	timer := time.AfterFunc(timeout, func() {
		logger.Errorf("Timeout waiting for the transaction: %s", txID)
		iter.Close()
	})
	defer timer.Stop()

	for {
		filteredBlock, err := iter.NextFilteredBlock()
		if err == ErrEOF {
			logger.Error("Error getting EOF when waiting committed blocks, that should never happen")
		}
		if err == ErrClosed {
			logger.Error("Stop receiving because the iterator is closed")
		}
		if err != nil {
			return false, err
		}
		for _, tx := range filteredBlock.FilteredTransactions {
			if tx.Txid == txID {
				return tx.TxValidationCode == pb.TxValidationCode_VALID, nil
			}
		}
	}
}
