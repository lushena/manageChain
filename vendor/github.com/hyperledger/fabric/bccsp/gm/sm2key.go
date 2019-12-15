/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package gm

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"fmt"

	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/sm/sm2"
)

type sm2PrivateKey struct {
	privKey *ecdsa.PrivateKey
	raw     []byte
}

// Bytes converts this key to its byte representation,
// if this operation is allowed.
func (k *sm2PrivateKey) Bytes() (raw []byte, err error) {
	return k.raw, nil
}

// SKI returns the subject key identifier of this key.
func (k *sm2PrivateKey) SKI() (ski []byte) {
	raw := elliptic.Marshal(k.privKey.Curve, k.privKey.PublicKey.X, k.privKey.PublicKey.Y)

	// Hash it
	hash := sha256.New()
	hash.Write(raw)
	return hash.Sum(nil)
}

// Symmetric returns true if this key is a symmetric key,
// false if this key is asymmetric
func (k *sm2PrivateKey) Symmetric() bool {
	return false
}

// Private returns true if this key is a private key,
// false otherwise.
func (k *sm2PrivateKey) Private() bool {
	return true
}

// PublicKey returns the corresponding public key part of an asymmetric public/private key pair.
// This method returns an error in symmetric key schemes.
func (k *sm2PrivateKey) PublicKey() (bccsp.Key, error) {
	return &sm2PublicKey{pubKey: &k.privKey.PublicKey}, nil
	//	return nil, errors.New("Cannot call this method on a symmetric key.")
}

type sm2PublicKey struct {
	pubKey *ecdsa.PublicKey
}

// Bytes converts this key to its byte representation,
// if this operation is allowed.
func (k *sm2PublicKey) Bytes() (raw []byte, err error) {
	raw, err = x509.MarshalPKIXPublicKey(k.pubKey)
	if err != nil {
		return nil, fmt.Errorf("Failed marshalling key [%s]", err)
	}
	return
}

// SKI returns the subject key identifier of this key.
func (k *sm2PublicKey) SKI() (ski []byte) {
	if k.pubKey == nil {
		return nil
	}

	// Marshall the public key
	raw := elliptic.Marshal(k.pubKey.Curve, k.pubKey.X, k.pubKey.Y)

	// Hash it
	hash := sha256.New()
	hash.Write(raw)
	return hash.Sum(nil)
}

// Symmetric returns true if this key is a symmetric key,
// false if this key is asymmetric
func (k *sm2PublicKey) Symmetric() bool {
	return false
}

// Private returns true if this key is a private key,
// false otherwise.
func (k *sm2PublicKey) Private() bool {
	return false
}

// PublicKey returns the corresponding public key part of an asymmetric public/private key pair.
// This method returns an error in symmetric key schemes.
func (k *sm2PublicKey) PublicKey() (bccsp.Key, error) {
	return k, nil
}

type sm2Signer struct{}

func (s *sm2Signer) Sign(k bccsp.Key, digest []byte, opts bccsp.SignerOpts) (signature []byte, err error) {
	keyBytes, _ := k.Bytes()

	key, err := sm2.LoadSM2PrivKeyFromBytes(keyBytes)
	if err != nil {
		return nil, err
	}
	defer sm2.FreeSm2EcKey(key)
	//return sm2.SM2Sign(key, digest)
	sig, err := sm2.SM2Sign(key, digest)
	// fmt.Println("sm2Sign", sig, err)
	return sig, err
}

func (v *sm2Signer) Verify(k bccsp.Key, signature, digest []byte, opts bccsp.SignerOpts) (valid bool, err error) {
	keyBytes, _ := k.Bytes()
	key, err := sm2.LoadSM2PubKeyFromByte(keyBytes)
	if err != nil {
		return false, err
	}
	defer sm2.FreeSm2EcKey(key)
	//return sm2.SM2Verify(key, digest, signature)
	verify, err := sm2.SM2Verify(key, digest, signature)
	// fmt.Println("sm2Verify ", verify, err)
	return verify, err
}
