package sm2

// #cgo LDFLAGS: -L../../librarys -lsmcryptokit -lcrypto
// #include "../../include/sm/smcryptokit.h"
import "C"
import "unsafe"

import (
	"fmt"
	"math/big"

	sm3 "github.com/hyperledger/fabric/sm/sm3"
)

type Sm2EcKey struct {
	ecKey *C.EC_KEY
}

func LoadSM2PrivKey(path []byte) (*Sm2EcKey, error) {
	sm2EcKey := &Sm2EcKey{}
	sm2EcKey.ecKey = C.LoadSM2PrivKeyFromFile(unsafe.Pointer(&path[0]), C.int(len(path)))
	if sm2EcKey.ecKey == nil {
		return nil, fmt.Errorf("SM2NewEcKey error")
	}
	return sm2EcKey, nil
}

func LoadSM2PubKeyFromByte(pub []byte) (*Sm2EcKey, error) {
	sm2EcKey := &Sm2EcKey{}
	// fmt.Println("the LoadSM2PubKeyFromByte ", len(pub), pub)
	sm2EcKey.ecKey = C.LoadSM2PubKeyFromBytes(unsafe.Pointer(&pub[0]), C.int(len(pub)))

	if sm2EcKey.ecKey == nil {
		return nil, fmt.Errorf("LoadSM2PubKeyFromByte error")
	}
	//fmt.Println("the sm2EcKey ecKey is ",sm2EcKey.ecKey)
	return sm2EcKey, nil
}

func LoadSM2PrivKeyFromBytes(priv []byte) (*Sm2EcKey, error) {
	sm2EcKey := &Sm2EcKey{}
	sm2EcKey.ecKey = C.LoadSM2PrivKeyFromBytes(unsafe.Pointer(&priv[0]), C.int(len(priv)))
	if sm2EcKey.ecKey == nil {
		return nil, fmt.Errorf("LoadSM2PrivKeyFromByte error")
	}
	return sm2EcKey, nil
}

func NewSm2EcKey() (*Sm2EcKey, error) {
	sm2EcKey := &Sm2EcKey{}
	sm2EcKey.ecKey = C.SM2NewEcKey()
	if sm2EcKey.ecKey == nil {
		return nil, fmt.Errorf("SM2NewEcKey error")
	}
	return sm2EcKey, nil
}

func FreeSm2EcKey(sm2EcKey *Sm2EcKey) {
	if sm2EcKey.ecKey != nil {
		C.SM2FreeEcKey(sm2EcKey.ecKey)
	}
}

// SM2Sign signs
func SM2Sign(signKey interface{}, msg []byte) ([]byte, error) {
	sm2Eckey := signKey.(*Sm2EcKey)

	hash := sm3.New()
	hash.Write(msg)
	h := hash.Sum(nil)
	// fmt.Printf("hash of %s: %x\n", msg, h)

	sig := make([]byte, 128, 128)
	var sigLength uint32
	var csigLength C.uint
	res := C.SM2Sign(C.int(0), unsafe.Pointer(&h[0]), C.int(len(h)),
		unsafe.Pointer(&sig[0]), &csigLength, unsafe.Pointer(sm2Eckey.ecKey))

	sigLength = uint32(csigLength)
	if res == 0 {
		// fmt.Printf("SM2Sign error\n")
		return nil, fmt.Errorf("SM2Sign error")
	}

	// fmt.Printf("res:%v\n", res)
	// fmt.Printf("signature: 0x%x\n", sig[:sigLength])
	// fmt.Printf("sigLength: %v\n", sigLength)

	return sig[:sigLength], nil
}

func SM2SignDirect(signKey interface{}, msg []byte) (*big.Int, *big.Int, error) {
	sm2Eckey := signKey.(*Sm2EcKey)
	hash := sm3.New()
	hash.Write(msg)
	h := hash.Sum(nil)
	// fmt.Printf("hash of %s: %x\n", msg, h)

	r := make([]byte, 256, 256)
	s := make([]byte, 256, 256)
	//	r := make([]byte,32)
	//	s := make([]byte,32)
	var rLen, sLen C.int
	res := C.SM2SignDirect(C.int(0), unsafe.Pointer(&h[0]), C.int(len(h)),
		unsafe.Pointer(&r[0]), &rLen, unsafe.Pointer(&s[0]), &sLen,
		unsafe.Pointer(sm2Eckey.ecKey))

	if res == 0 {
		// fmt.Printf("SM2SignDirect error\n")
		return nil, nil, fmt.Errorf("SM2SignDirect error")
	}

	//fmt.Println(rLen,"--",r)
	//fmt.Println(sLen,"--",s)

	//rstr := string(r[:int(rLen)])
	//sstr := string(s[:int(sLen)])
	rByte := r[:int(rLen)]
	sByte := s[:int(sLen)]
	//fmt.Printf("rstr:   0x%v\n", rstr)
	//fmt.Printf("sstr:   0x%v\n", sstr)

	//ir, bres := new(big.Int).SetString(rstr, 16)
	ir := new(big.Int)
	ir.SetBytes(rByte)
	is := new(big.Int)
	is.SetBytes(sByte)
	//fmt.Println(is,ir)
	//	if !bres {
	//		return nil, nil, fmt.Errorf("convert string to bigint failed, rstr:%s", rstr)
	//	}
	//	is, bres := new(big.Int).SetString(sstr, 16)
	//	if !bres {
	//		return nil, nil, fmt.Errorf("convert string to bigint failed, sstr:%s", sstr)
	//	}

	return ir, is, nil
}

// SM2Verify verifies
func SM2Verify(verKey interface{}, msg, signature []byte) (bool, error) {
	if len(signature) == 0 {
		return false, fmt.Errorf("invalid length of signature:%v", len(signature))
	}

	sm2Eckey := verKey.(*Sm2EcKey)

	hash := sm3.New()
	hash.Write(msg)
	h := hash.Sum(nil)
	// fmt.Printf("hash of %s: %x\n", msg, h)

	res := C.SM2Verify(C.int(0), unsafe.Pointer(&h[0]), C.int(len(h)),
		unsafe.Pointer(&signature[0]), C.int(len(signature)), unsafe.Pointer(sm2Eckey.ecKey))

	if res == 1 {
		return true, nil
	} else if res == 0 {
		return false, nil
	} else {
		return false, fmt.Errorf("SM2Verify error, res:%v", res)
	}
}

func SM2VerifyDirect(verKey interface{}, msg []byte, r, s *big.Int) (bool, error) {
	if r == nil {
		return false, fmt.Errorf("SM2Verify r==nil")
	}

	if s == nil {
		return false, fmt.Errorf("SM2Verify s==nil")
	}

	sm2Eckey := verKey.(*Sm2EcKey)

	hash := sm3.New()
	hash.Write(msg)
	h := hash.Sum(nil)
	// fmt.Printf("hash of %s: %x\n", msg, h)

	//	hexr := []byte(fmt.Sprintf("%X", r))
	//	hexs := []byte(fmt.Sprintf("%X", s))
	hexr := r.Bytes()
	hexs := s.Bytes()
	//fmt.Printf("hexr:   0x%s\n", hexr)
	//fmt.Printf("len hexr: %v\n", len(hexr))
	//fmt.Printf("hexs:   0x%s\n", hexs)
	//fmt.Printf("len hexs: %v\n", len(hexs))

	res := C.SM2VerifyDirect(C.int(0), unsafe.Pointer(&h[0]), C.int(len(h)),
		unsafe.Pointer(&hexr[0]), C.int(len(hexr)),
		unsafe.Pointer(&hexs[0]), C.int(len(hexs)),
		unsafe.Pointer(sm2Eckey.ecKey))
	//fmt.Println("xxxxx",res)

	if res == 1 {
		return true, nil
	} else if res == 0 {
		return false, nil
	} else {
		return false, fmt.Errorf("SM2Verify error, res:%v", res)
	}
}

// VerifySignCapability tests signing capabilities
func VerifySignCapability(tempSK interface{}, certPK interface{}) error {
	/* TODO: reactive or remove
	msg := []byte("This is a message to be signed and verified by ECDSA!")

	sigma, err := ECDSASign(tempSK, msg)
	if err != nil {
		//		log.Errorf("Error signing [%s].", err.Error())

		return err
	}

	ok, err := ECDSAVerify(certPK, msg, sigma)
	if err != nil {
		//		log.Errorf("Error verifying [%s].", err.Error())

		return err
	}

	if !ok {
		//		log.Errorf("Signature not valid.")

		return errors.New("Signature not valid.")
	}

	//	log.Infof("Verifing signature capability...done")
	*/
	return nil
}
