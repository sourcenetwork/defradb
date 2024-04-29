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

// PromptFunc is a callback used to retrieve the user's password.
type PromptFunc func(s string) ([]byte, error)

// Keyring provides a simple set/get interface for a keyring service.
type Keyring interface {
	// Set stores the given key in the keystore under the given name.
	Set(name string, key []byte) error
	// Get returns the key with the given name from the keystore.
	Get(name string) ([]byte, error)
	// Delete removes the key with the given name from the keystore.
	Delete(name string) error
}

// Open attempts to open the keyring file from the given directory.
//
// If the directory is an empty string the system keystore will be used instead.
func Open(dir string, service string, prompt PromptFunc) (Keyring, error) {
	if dir == "" {
		return openSystemKeyring(service), nil
	}
	password, err := prompt("Enter keystore password:")
	if err != nil {
		return nil, err
	}
	return openFileKeyring(dir, password)
}
