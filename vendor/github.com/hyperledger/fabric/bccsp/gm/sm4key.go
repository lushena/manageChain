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
	"crypto/sha256"
	"github.com/hyperledger/fabric/sm/sm4"
	"github.com/hyperledger/fabric/bccsp"
)

type sm4PrivateKey struct {
	key	[]byte
}

// Bytes converts this key to its byte representation,
// if this operation is allowed.
func (k *sm4PrivateKey) Bytes() (raw []byte, err error) {
	return k.key, nil
}

// SKI returns the subject key identifier of this key.
func (k *sm4PrivateKey) SKI() (ski []byte) {
	// Hash it
	hash := sha256.New()
	hash.Write(k.key)
	return hash.Sum(nil)
}

// Symmetric returns true if this key is a symmetric key,
// false if this key is asymmetric
func (k *sm4PrivateKey) Symmetric() bool {
	return true
}

// Private returns true if this key is a private key,
// false otherwise.
func (k *sm4PrivateKey) Private() bool {
	return true
}

// PublicKey returns the corresponding public key part of an asymmetric public/private key pair.
// This method returns an error in symmetric key schemes.
func (k *sm4PrivateKey) PublicKey() (bccsp.Key, error) {
	return &sm4PrivateKey{key:k.key}, nil
}



type sm4Encryptor struct{}

func (* sm4Encryptor) Encrypt(k bccsp.Key, plaintext []byte, opts bccsp.EncrypterOpts) (ciphertext []byte, err error) {
	keyBytes,err := k.Bytes()
	if err != nil{
		return nil,err
	}
	return sm4.Encrypt(keyBytes,plaintext)
}

type sm4Decryptor struct{}

func (*sm4Decryptor) Decrypt(k bccsp.Key, ciphertext []byte, opts bccsp.DecrypterOpts) (plaintext []byte, err error) {
	keyBytes,err := k.Bytes()
	if err != nil{
		return nil,err
	}
	return sm4.Decrypt(keyBytes,ciphertext)
}
