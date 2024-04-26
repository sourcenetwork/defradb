// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"crypto/rand"
	"errors"
	"syscall"

	"github.com/99designs/keyring"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	peerKeyName             = "peer_key"
	badgerEncryptionKeyName = "badger_encryption_key"
)

// openKeyring attempts to open the keyring file from the given directory.
//
// If the directory is an empty string the best option for the current OS will be used.
func openKeyring(cmd *cobra.Command, dir string) (keyring.Keyring, error) {
	var allowedBackends []keyring.BackendType
	if dir != "" {
		// only allow file backend if a directory is specified
		allowedBackends = append(allowedBackends, keyring.FileBackend)
	}

	prompt := keyring.PromptFunc(func(s string) (string, error) {
		cmd.Print(s)
		pass, err := term.ReadPassword(int(syscall.Stdin))
		return string(pass), err
	})

	return keyring.Open(keyring.Config{
		AllowedBackends:         allowedBackends,
		ServiceName:             "defradb",
		KeychainName:            "defradb",
		KeychainPasswordFunc:    prompt,
		FilePasswordFunc:        prompt,
		FileDir:                 dir,
		KeyCtlScope:             "user",
		KeyCtlPerm:              0, // TODO
		KWalletAppID:            "defradb",
		KWalletFolder:           "defradb",
		LibSecretCollectionName: "defradb",
		PassPrefix:              "defradb",
		WinCredPrefix:           "defradb",
	})
}

// generateAES256 generates a new random AES-256 bit encryption key.
func generateAES256() ([]byte, error) {
	data := make([]byte, 32)
	_, err := rand.Read(data)
	return data, err
}

// loadOrGenerateAES256 attempts to load the AES-256 bit key with the given name.
//
// If the key does not exist a new random key is generated and stored in the keyring.
func loadOrGenerateAES256(kr keyring.Keyring, name string) ([]byte, error) {
	item, err := kr.Get(name)
	if err == nil {
		return item.Data, nil
	}
	if !errors.Is(err, keyring.ErrKeyNotFound) {
		return nil, err
	}
	key, err := generateAES256()
	if err != nil {
		return nil, err
	}
	err = kr.Set(keyring.Item{
		Key:  name,
		Data: key,
	})
	if err != nil {
		return nil, err
	}
	return key, nil
}

// generateEd25519 generates a new random Ed25519 private key.
func generateEd25519() (crypto.PrivKey, error) {
	key, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
	return key, err
}

// loadOrGenerateEd25519 attempts to load the Ed25519 private key with the given name.
//
// If the key does not exist a new random key is generated and stored in the keyring.
func loadOrGenerateEd25519(kr keyring.Keyring, name string) (crypto.PrivKey, error) {
	item, err := kr.Get(name)
	if err == nil {
		return crypto.UnmarshalPrivateKey(item.Data)
	}
	if !errors.Is(err, keyring.ErrKeyNotFound) {
		return nil, err
	}
	key, err := generateEd25519()
	if err != nil {
		return nil, err
	}
	data, err := crypto.MarshalPrivateKey(key)
	if err != nil {
		return nil, err
	}
	err = kr.Set(keyring.Item{
		Key:  name,
		Data: data,
	})
	if err != nil {
		return nil, err
	}
	return key, nil
}
