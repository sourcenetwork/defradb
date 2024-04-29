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

import (
	"os"
	"path/filepath"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/zalando/go-keyring"
)

var _ Keyring = (*fileKeyring)(nil)

var keyEncryptionAlgorithm = jwa.PBES2_HS512_A256KW

// fileKeyring is a keyring that stores keys in encrypted files.
type fileKeyring struct {
	// dir is the keystore root directory
	dir string
	// key is the keystore encryption key
	key []byte
}

func openFileKeyring(dir string, password []byte) (*fileKeyring, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &fileKeyring{
		dir: dir,
		key: password,
	}, nil
}

func (f *fileKeyring) Set(name string, key []byte) error {
	cipher, err := jwe.Encrypt(key, jwe.WithKey(keyEncryptionAlgorithm, f.key))
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(f.dir, name), cipher, 0755)
}

func (f *fileKeyring) Get(name string) ([]byte, error) {
	cipher, err := os.ReadFile(filepath.Join(f.dir, name))
	if os.IsNotExist(err) {
		return nil, keyring.ErrNotFound
	}
	return jwe.Decrypt(cipher, jwe.WithKey(keyEncryptionAlgorithm, f.key))
}

func (f *fileKeyring) Delete(user string) error {
	err := os.Remove(filepath.Join(f.dir, user))
	if os.IsNotExist(err) {
		return keyring.ErrNotFound
	}
	return err
}
