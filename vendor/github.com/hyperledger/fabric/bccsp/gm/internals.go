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

	"github.com/hyperledger/fabric/bccsp"
)

// KeyGenerator is a BCCSP-like interface that provides key generation algorithms
type KeyGenerator interface {

	// KeyGen generates a key using opts.
	KeyGen(opts bccsp.KeyGenOpts) (k bccsp.Key, err error)
}

// KeyImporter is a BCCSP-like interface that provides key import algorithms
type KeyImporter interface {

	// KeyImport imports a key from its raw representation using opts.
	// The opts argument should be appropriate for the primitive used.
	KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (k bccsp.Key, err error)
}

// Encryptor is a BCCSP-like interface that provides encryption algorithms
type Encryptor interface {

	// Encrypt encrypts plaintext using key k.
	// The opts argument should be appropriate for the algorithm used.
	Encrypt(k bccsp.Key, plaintext []byte, opts bccsp.EncrypterOpts) (ciphertext []byte, err error)
}

// Decryptor is a BCCSP-like interface that provides decryption algorithms
type Decryptor interface {

	// Decrypt decrypts ciphertext using key k.
	// The opts argument should be appropriate for the algorithm used.
	Decrypt(k bccsp.Key, ciphertext []byte, opts bccsp.DecrypterOpts) (plaintext []byte, err error)
}

