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
package factory

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/bccsp/gm"
)

const (
	GMBasedFactoryName = "GM"
)

// SWFactory is the factory of the software-based BCCSP.
type GMFactory struct{}

// Name returns the name of this factory
func (f *GMFactory) Name() string {
	return GMBasedFactoryName
}

// Get returns an instance of BCCSP using Opts.
func (f *GMFactory) Get(config *FactoryOpts) (bccsp.BCCSP, error) {
	// Validate arguments
	if config == nil {
		return nil, errors.New("Invalid config. It must not be nil.")
	}
	if config.ProviderName == "GM" {
		swOpts := config.SwOpts

		var ks bccsp.KeyStore
		if swOpts.Ephemeral == true {
			ks = gm.NewDummyKeyStore()
		} else if swOpts.FileKeystore != nil {
			fks, err := gm.NewFileBasedKeyStore(nil, swOpts.FileKeystore.KeyStorePath, false)
			if err != nil {
				return nil, fmt.Errorf("Failed to initialize software key store: %s", err)
			}
			ks = fks
		} else {
			// Default to DummyKeystore
			ks = gm.NewDummyKeyStore()
		}

		//fks, err := gm.NewFileBasedKeyStore(nil, swOpts.FileKeystore.KeyStorePath, false)
		//fks, err := gm.NewFileBasedKeyStore(nil, "", false)
		/*
			if err!=nil{
				logger.Error("[hzyangwenlong] error !!!!! when Get GMFactory keystore",err)
			}
		*/
		return gm.New(ks)
	}

	return nil, errors.New("Invalid config. It will set to gm")
}
