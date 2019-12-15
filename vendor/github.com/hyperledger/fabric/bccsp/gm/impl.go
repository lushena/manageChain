package gm

import (
	"hash"
	"reflect"
	"github.com/pkg/errors"
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/sm/sm3"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/bccsp/utils/ecies"
)

var (
	logger = flogging.MustGetLogger("bccsp_gm")
)

type impl struct {
	ks bccsp.KeyStore
	keyImporters  map[reflect.Type]KeyImporter
	signer *sm2Signer
	verifier *sm2Signer
	keyGenerators map[reflect.Type]KeyGenerator
	encryptors    map[reflect.Type]Encryptor
	decryptors    map[reflect.Type]Decryptor
}

func New(keyStore bccsp.KeyStore) (bccsp.BCCSP, error) {

	impl := &impl{
			ks:keyStore,
		signer:&sm2Signer{},
			verifier:&sm2Signer{},
		}
	keyImporters := make(map[reflect.Type]KeyImporter)
	keyImporters[reflect.TypeOf(&bccsp.ECDSAPKIXPublicKeyImportOpts{})] = &ecdsaPKIXPublicKeyImportOptsKeyImporter{}
	keyImporters[reflect.TypeOf(&bccsp.ECDSAPrivateKeyImportOpts{})] = &ecdsaPrivateKeyImportOptsKeyImporter{}
	keyImporters[reflect.TypeOf(&bccsp.ECDSAGoPublicKeyImportOpts{})] = &ecdsaGoPublicKeyImportOptsKeyImporter{}
	keyImporters[reflect.TypeOf(&bccsp.X509PublicKeyImportOpts{})] = &x509PublicKeyImportOptsKeyImporter{bccsp: impl}
	keyImporters[reflect.TypeOf(&bccsp.SM4ImportKeyOpts{})] = &sm4ImportKeyOptsKeyImporter{}

	// Set the encryptors
	encryptors := make(map[reflect.Type]Encryptor)
	encryptors[reflect.TypeOf(&sm4PrivateKey{})] = &sm4Encryptor{}

	// Set the decryptors
	decryptors := make(map[reflect.Type]Decryptor)
	decryptors[reflect.TypeOf(&sm4PrivateKey{})] = &sm4Decryptor{}
	keyGenerators := make(map[reflect.Type]KeyGenerator)
	keyGenerators[reflect.TypeOf(&bccsp.SM4KeyGenOpts{})] = &sm4KeyGenerator{}

	impl.keyGenerators = keyGenerators
	impl.encryptors = encryptors
	impl.decryptors = decryptors
	impl.keyImporters = keyImporters
	return impl, nil
}

// KeyGen generates a key using opts.
func (csp *impl) KeyGen(opts bccsp.KeyGenOpts) (k bccsp.Key, err error) {
	// Validate arguments
	if opts == nil {
		return nil, errors.Errorf("BCCSP BadRequest Invalid Opts parameter. It must not be nil.")
	}

	keyGenerator, found := csp.keyGenerators[reflect.TypeOf(opts)]
	if !found {
		return nil, errors.Errorf("BCCSP NotFound Unsupported 'KeyGenOpts' provided [%v]", opts)
	}

	k, err = keyGenerator.KeyGen(opts)
	if err != nil {
		return nil, errors.Wrapf(err,"BCCSP Internal Failed generating key with opts [%v]", opts)
	}
	return k,nil
}

// KeyDeriv derives a key from k using opts.
// The opts argument should be appropriate for the primitive used.
func (csp *impl) KeyDeriv(k bccsp.Key, opts bccsp.KeyDerivOpts) (dk bccsp.Key, err error) {
	return nil, nil
}

// KeyImport imports a key from its raw representation using opts.
// The opts argument should be appropriate for the primitive used.
func (csp *impl) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (k bccsp.Key, err error) {
	// Validate arguments
	if raw == nil {
		return nil, errors.Errorf("BCCSP BadRequest Invalid raw. It must not be nil.")
	}
	if opts == nil {
		return nil, errors.Errorf("BCCSP BadRequest Invalid opts. It must not be nil.")
	}

	keyImporter, found := csp.keyImporters[reflect.TypeOf(opts)]
	if !found {
		return nil, errors.Errorf("BCCSP NotFound Unsupported 'KeyImportOpts' provided [%v]", opts)
	}

	k, err = keyImporter.KeyImport(raw, opts)
	if err != nil {
		return nil, errors.Wrapf(err,"BCCSP Internal Failed importing key with opts [%v]", opts)
	}

	// If the key is not Ephemeral, store it.
	if !opts.Ephemeral() {
		// Store the key
		err = csp.ks.StoreKey(k)
		if err != nil {
			return nil, errors.Wrapf(err,"BCCSP Internal Failed storing imported key with opts [%v]", opts)
		}
	}

	return
}

// GetKey returns the key this CSP associates to
// the Subject Key Identifier ski.
func (csp *impl) GetKey(ski []byte) (k bccsp.Key, err error) {
	k, err = csp.ks.GetKey(ski)
	if err != nil {
		return nil, errors.Wrapf(err,"BCCSP Internal Failed getting key for SKI [%v]", ski)
	}

	return
}

// Hash hashes messages msg using options opts.
func (csp *impl) Hash(msg []byte, opts bccsp.HashOpts) (digest []byte, err error) {
	// Validate arguments
	if opts == nil {
		return nil, errors.Errorf("BCCSP BadRequest Invalid opts. It must not be nil.")
	}
	//logger.Debugf("[hzyangwenlong] this is gm Hash")
	//直接调用sm3来进行hash
	h := sm3.New()
	h.Write(msg)
	digest = h.Sum(nil)
	//return []byte("hahahah"), nil
	return digest,nil
}

// GetHash returns and instance of hash.Hash using options opts.
// If opts is nil then the default hash function is returned.
func (csp *impl) GetHash(opts bccsp.HashOpts) (hash.Hash, error) {
	//返回sm3的hash函数
	hash := sm3.New()
	return hash, nil
}

// Sign signs digest using key k.
// The opts argument should be appropriate for the primitive used.
//
// Note that when a signature of a hash of a larger message is needed,
// the caller is responsible for hashing the larger message and passing
// the hash (as digest).
func (csp *impl) Sign(k bccsp.Key, digest []byte, opts bccsp.SignerOpts) (signature []byte, err error) {
	// Validate arguments
	//调用sm2的签名
	//logger.Debugf("[hzyangwenlong] this is gm Sign")

	if k == nil {
		return nil, errors.Errorf("BCCSP BadRequest Invalid Key. It must not be nil.")
	}
	if len(digest) == 0 {
		return nil, errors.Errorf("BCCSP BadRequest Invalid digest. Cannot be empty.")
	}

	signer := csp.signer

	signature, err = signer.Sign(k, digest, opts)
	if err != nil {
		return nil, errors.Wrapf(err,"BCCSP Internal Failed signing with opts [%v]", opts)
	}

	return
	//return []byte("hahahah"), nil
}

// Verify verifies signature against key k and digest
func (csp *impl) Verify(k bccsp.Key, signature, digest []byte, opts bccsp.SignerOpts) (valid bool, err error) {
	//用sm2的公钥来验证
	//logger.Debugf("[hzyangwenlong] this is gm Verify")
	// Validate arguments
	if k == nil {
		return false, errors.Errorf("BCCSP BadRequest Invalid Key. It must not be nil.")
	}
	if len(signature) == 0 {
		return false, errors.Errorf("BCCSP BadRequest Invalid signature. Cannot be empty.")
	}
	if len(digest) == 0 {
		return false, errors.Errorf("BCCSP BadRequest Invalid digest. Cannot be empty.")
	}

	verifier := csp.verifier

	valid, err = verifier.Verify(k, signature, digest, opts)
	if err != nil {
		return false, errors.Wrapf(err,"BCCSP Internal Failed verifing with opts [%v]", opts)
	}

	return
}

// Encrypt encrypts plaintext using key k.
// The opts argument should be appropriate for the primitive used.
func (csp *impl) Encrypt(k bccsp.Key, plaintext []byte, opts bccsp.EncrypterOpts) (ciphertext []byte, err error) {
	// Validate arguments
	if k == nil {
		return nil, errors.Errorf("BCCSP BadRequest Invalid Key. It must not be nil.")
	}

	encryptor, found := csp.encryptors[reflect.TypeOf(k)]
	if !found {
		//用sm2来进行对称加密，里面采用sm4加密，具体看ecies
		keyBytes,_ := k.Bytes()
		key, err := ecies.ParseECPublicKey(keyBytes)
		if err != nil{
			return nil,err
		}
		return ecies.EciesEncrypt(key,plaintext,true)
	}

	return encryptor.Encrypt(k, plaintext, opts)

}

// Decrypt decrypts ciphertext using key k.
// The opts argument should be appropriate for the primitive used.
func (csp *impl) Decrypt(k bccsp.Key, ciphertext []byte, opts bccsp.DecrypterOpts) (plaintext []byte, err error) {
	// Validate arguments
	if k == nil {
		return nil, errors.Errorf("BCCSP BadRequest Invalid Key. It must not be nil.")
	}

	decryptor, found := csp.decryptors[reflect.TypeOf(k)]
	if !found {
		//用sm2来进行对称解密，里面采用sm4解密，具体看ecies
		keyBytes,_ := k.Bytes()

		key,err := ecies.ParseECPrivateKey(keyBytes)
		if err != nil{
			return nil,err
		}
		return ecies.EciesDecrypt(key,ciphertext,true)
	}

	plaintext, err = decryptor.Decrypt(k, ciphertext, opts)
	if err != nil {
		return nil, errors.Wrapf(err,"BCCSP Internal Failed decrypting with opts [%v]", opts)
	}
	return 
}
