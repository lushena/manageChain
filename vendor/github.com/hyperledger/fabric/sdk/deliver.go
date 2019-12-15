package sdk

import (
	"context"
	"math"

	cb "github.com/hyperledger/fabric/protos/common"
	ab "github.com/hyperledger/fabric/protos/orderer"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Errors ...
var (
	ErrEOF    = errors.New("EOF")
	ErrClosed = errors.New("block iterator is closed")
)

var (
	seekNewest = &ab.SeekPosition{
		Type: &ab.SeekPosition_Newest{
			Newest: &ab.SeekNewest{},
		},
	}
	seekOldest = &ab.SeekPosition{
		Type: &ab.SeekPosition_Oldest{
			Oldest: &ab.SeekOldest{},
		},
	}
	seekMax = seekSpecified(math.MaxUint64)
)

func seekSpecified(num uint64) *ab.SeekPosition {
	return &ab.SeekPosition{
		Type: &ab.SeekPosition_Specified{
			Specified: &ab.SeekSpecified{
				Number: num,
			},
		},
	}
}

func seekInfo(start, end *ab.SeekPosition) *ab.SeekInfo {
	return &ab.SeekInfo{
		Start:    start,
		Stop:     end,
		Behavior: ab.SeekInfo_BLOCK_UNTIL_READY,
	}
}

// BlockIterator ...
type BlockIterator struct {
	blockC  chan *cb.Block
	fblockC chan *pb.FilteredBlock
	errorC  chan error
	stopC   chan struct{}
	cancel  context.CancelFunc
}

// Close ...
func (br *BlockIterator) Close() {
	select {
	case <-br.stopC:
	default:
		br.cancel()
		close(br.stopC)
	}
}

// NextBlock ...
func (br *BlockIterator) NextBlock() (*cb.Block, error) {
	var block *cb.Block
	var err error

	ok := true

	select {
	case <-br.stopC:
		err = ErrClosed
	case err, ok = <-br.errorC:
	case block, ok = <-br.blockC:
	}

	if !ok {
		err = ErrClosed
	}

	return block, err
}

// NextFilteredBlock ...
func (br *BlockIterator) NextFilteredBlock() (*pb.FilteredBlock, error) {
	var fblock *pb.FilteredBlock
	var err error

	ok := true
	select {
	case <-br.stopC:
		err = ErrClosed
	case err, ok = <-br.errorC:
	case fblock, ok = <-br.fblockC:
	}

	if !ok {
		err = ErrClosed
	}

	return fblock, err

}

// DeliverClient delivers blocks from orderer
type DeliverClient struct {
	endpoint *Endpoint
}

// NewDeliverClient ...
func NewDeliverClient(deliver *Endpoint) *DeliverClient {
	return &DeliverClient{
		endpoint: deliver,
	}
}

// RequestBlock requests a single block once a time
func (dc *DeliverClient) RequestBlock(req *cb.Envelope) (*cb.Block, error) {
	de, conn, cancel, err := newAtomicBroadcastDeliverClient(dc.endpoint)
	if err != nil {
		logger.Error("Error creating deliver client", err)
		return nil, err
	}

	defer conn.Close()
	defer de.CloseSend()
	defer cancel()

	err = de.Send(req)
	if err != nil {
		logger.Error("Error sending block request", err)
		return nil, err
	}

	msg, err := de.Recv()
	if err != nil {
		return nil, errors.Wrap(err, "error receiving")
	}
	switch t := msg.Type.(type) {
	case *ab.DeliverResponse_Status:
		logger.Infof("Got status: %v", t)
		return nil, errors.Errorf("can't read the block: %v", t)
	case *ab.DeliverResponse_Block:
		logger.Infof("Received block: %v", t.Block.Header.Number)
		de.Recv() // Flush the success message
		return t.Block, nil
	default:
		return nil, errors.Errorf("response error: unknown type %T", t)
	}
}

// RequestBlocks requests blocks
func (dc *DeliverClient) RequestBlocks(req *cb.Envelope) (*BlockIterator, error) {
	de, conn, cancel, err := newAtomicBroadcastDeliverClient(dc.endpoint)
	if err != nil {
		logger.Error("Error creating deliver client", err)
		return nil, err
	}

	err = de.Send(req)
	if err != nil {
		logger.Error("Error sending block request", err)
		return nil, err
	}
	de.CloseSend()

	// receive ...
	blockC := make(chan *cb.Block)
	errorC := make(chan error)
	stopC := make(chan struct{})

	go func() {
		defer close(blockC)
		defer close(errorC)
		defer conn.Close()

		for {
			msg, err := de.Recv()
			if err != nil {
				select {
				case <-stopC:
					logger.Info("Exit receive loop ...")
				default:
					errorC <- errors.Wrap(err, "error receiving")
				}
				return
			}
			switch t := msg.Type.(type) {
			case *ab.DeliverResponse_Status:
				logger.Infof("Got status: %v", t)
				if t.Status == cb.Status_SUCCESS {
					errorC <- ErrEOF
				} else {
					errorC <- errors.Errorf("got status: %v", t)
				}
				return
			case *ab.DeliverResponse_Block:
				blockC <- t.Block
			default:
				errorC <- errors.Errorf("response error: unknown type %T", t)
				return
			}

		}

	}()

	return &BlockIterator{
		blockC: blockC,
		errorC: errorC,
		stopC:  stopC,
		cancel: cancel,
	}, nil

}

func newAtomicBroadcastDeliverClient(endpoint *Endpoint) (ab.AtomicBroadcast_DeliverClient, *grpc.ClientConn, context.CancelFunc, error) {
	conn, err := createConnection(endpoint)
	if err != nil {
		logger.Error("Error creating connection", err)
		return nil, nil, nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	de, err := ab.NewAtomicBroadcastClient(conn).Deliver(ctx)
	if err != nil {
		logger.Error("Error creating DeliverClient", err)
		conn.Close()
		cancel()
		return nil, nil, nil, err
	}
	return de, conn, cancel, nil
}

// PeerDeliveredClient ...
type PeerDeliveredClient struct {
	endpoint *Endpoint
}

// NewPeerDeliverClient ...
func NewPeerDeliverClient(deliver *Endpoint) *PeerDeliveredClient {
	return &PeerDeliveredClient{
		endpoint: deliver,
	}
}

// RequestFilteredBlocks ...
func (pdc *PeerDeliveredClient) RequestFilteredBlocks(req *cb.Envelope) (*BlockIterator, error) {
	dc, conn, cancel, err := newPeerDeliverFilteredClient(pdc.endpoint)
	if err != nil {
		logger.Error("Error creating DeliverFilteredClient", err)
		return nil, err
	}

	err = dc.Send(req)
	if err != nil {
		logger.Error("Error sending block request", err)
		return nil, err
	}
	dc.CloseSend()

	// receive ...
	fblockC := make(chan *pb.FilteredBlock)
	errorC := make(chan error)
	stopC := make(chan struct{})

	go func() {
		defer close(fblockC)
		defer close(errorC)
		defer conn.Close()

		for {
			msg, err := dc.Recv()
			if err != nil {
				select {
				case <-stopC:
					logger.Info("Exit receive loop ...")
				default:
					errorC <- errors.Wrap(err, "error receiving")
				}
				return
			}
			switch t := msg.Type.(type) {
			case *pb.DeliverResponse_Status:
				logger.Infof("Got status: %v", t)
				if t.Status == cb.Status_SUCCESS {
					errorC <- ErrEOF
				} else {
					errorC <- errors.Errorf("got status: %v", t)
				}
				return
			case *pb.DeliverResponse_FilteredBlock:
				fblockC <- t.FilteredBlock
			default:
				errorC <- errors.Errorf("response error: unknown type %T", t)
				return
			}

		}
	}()

	return &BlockIterator{
		fblockC: fblockC,
		errorC:  errorC,
		stopC:   stopC,
		cancel:  cancel,
	}, nil

}

func newPeerDeliverFilteredClient(endpoint *Endpoint) (pb.Deliver_DeliverFilteredClient, *grpc.ClientConn, context.CancelFunc, error) {

	conn, err := createConnection(endpoint)
	if err != nil {
		logger.Error("Error creating connection", err)
		return nil, nil, nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	dc, err := pb.NewDeliverClient(conn).DeliverFiltered(ctx)
	if err != nil {
		logger.Error("Error creating DeliverFilteredClient", err)
		conn.Close()
		cancel()
		return nil, nil, nil, err
	}
	return dc, conn, cancel, nil
}
