package sdk

import (
	"path"
	"sync"

	"github.com/golang/groupcache/lru"
	"github.com/hyperledger/fabric/bccsp/factory"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/msp/cache"
	logging "github.com/op/go-logging"
)

/*
Usage for interactive scene
  Query or Invoke:
	1. Call CreateChaincodeProposalBytes to get txID, proposalBytes, error
	2. Sign proposalBytes on user side, and send original proposalBytes and the signature
	3. Call EndorseToBytes to get proposalResponseBytes, responsePayloadBytes, error

	If you need to order that transaction, then:
	4. Call CreateChaincodeEnvelopeBytesFromBytes with proposalBytes and responsePayloadBytes to get envelopePayloadBytes
	5. Sign envelopePayloadBytes on user side, and send original envelopePayloadBytes and the signature
	6. Call Broadcast with envelopePayloadBytes and signature

  Create channel:
	1. Call CreateConfigUpdateEnvelopeBytes to get  configUpdate, signatureHeader, toSignBytes, error
	2. Sign toSignBytes on user side,
	3. Call CreateChannelEnvelopeBytes with chainID, creator, configUpdate, signature, signedSignature to get payloadBytes
	4. Sign payloadBytes on user side, and send original payloadBytes and the signature in Broadcast


*/

const pkgLogID = "sdk"

const mspCacheSize = 100

const keystore = "keystore"

var logger *logging.Logger
var mspCache *lru.Cache
var mspLock sync.Mutex

func init() {
	logger = flogging.MustGetLogger(pkgLogID)
	mspCache = lru.New(mspCacheSize)
}

// Client ...
// Hold peer/orderer's MSP
type Client struct {
	identity string
	mspInst  msp.MSP
	signer   msp.SigningIdentity
}

// NewClient ...
func NewClient(identity string, mspID string, dir string, gm bool) (*Client, error) {
	var config *factory.FactoryOpts
	if gm {
		config = &factory.FactoryOpts{
			ProviderName: "GM",
		}
	}
	mspInst, err := initializeMsp(identity, dir, mspID, config)
	if err != nil {
		logger.Error("Error initializing msp", err)
		return nil, err
	}

	signer, err := mspInst.GetDefaultSigningIdentity()
	if err != nil {
		logger.Error("Error getting defaultSigningIdentity", err)
		return nil, err
	}

	return &Client{
		identity: identity,
		mspInst:  mspInst,
		signer:   signer,
	}, nil
}

// Just support FABRIC msp with default factory opts
func initializeMsp(identity string, dir string, mspID string, bccspConfig *factory.FactoryOpts) (msp.MSP, error) {
	mspLock.Lock()
	defer mspLock.Unlock()

	if value, ok := mspCache.Get(identity); ok {
		logger.Debugf("Cached msp for identity: %s", identity)
		return value.(msp.MSP), nil
	}

	if bccspConfig == nil {
		bccspConfig = factory.GetDefaultOpts()
	}
	bccspConfig = msp.SetupBCCSPKeystoreConfig(bccspConfig, path.Join(dir, keystore))

	bccsp, err := factory.GetBCCSPFromOpts(bccspConfig)
	if err != nil {
		logger.Error("Error creating bccsp instance", err)
		return nil, err
	}

	conf, err := msp.GetLocalMspConfig(dir, bccspConfig, mspID)
	if err != nil {
		return nil, err
	}

	mspInst, err := msp.NewBccspMsp(msp.MSPv1_0, bccsp)
	if err != nil {
		return nil, err
	}
	mspInst, err = cache.New(mspInst)
	if err != nil {
		return nil, err
	}
	err = mspInst.Setup(conf)
	if err != nil {
		return nil, err
	}
	mspCache.Add(identity, mspInst)

	return mspInst, nil
}
