// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keyring

// Keyring provides a simple set/get interface for a keyring service.
type Keyring interface {
	// Set stores the given key in the keystore under the given name.
	//
	// If a key with the given name already exists it will be overriden.
	Set(name string, key []byte) error
	// Get returns the key with the given name from the keystore.
	//
	// If a key with the given name does not exist `ErrNotFound` is returned.
	Get(name string) ([]byte, error)
	// Delete removes the key with the given name from the keystore.
	//
	// If a key with that name does not exist `ErrNotFound` is returned.
	Delete(name string) error
}
