// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package encryption

import (
	"context"
	"crypto/rand"
	"errors"
	"io"

	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
)

var generateEncryptionKeyFunc = generateEncryptionKey

const keyLength = 32 // 32 bytes for AES-256

const testEncryptionKey = "examplekey1234567890examplekey12"

// generateEncryptionKey generates a random AES key.
func generateEncryptionKey() ([]byte, error) {
	key := make([]byte, keyLength)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

// generateTestEncryptionKey generates a deterministic encryption key for testing.
func generateTestEncryptionKey() ([]byte, error) {
	return []byte(testEncryptionKey), nil
}

type DocEncryptor struct {
	shouldGenerateKey bool
	ctx               context.Context
	store             datastore.DSReaderWriter
}

func newDocEncryptor(ctx context.Context) *DocEncryptor {
	return &DocEncryptor{ctx: ctx}
}

func (d *DocEncryptor) EnableKeyGeneration() {
	d.shouldGenerateKey = true
}

func (d *DocEncryptor) SetStore(store datastore.DSReaderWriter) {
	d.store = store
}

func (d *DocEncryptor) Encrypt(docID string, fieldID uint32, plainText []byte) ([]byte, error) {
	encryptionKey, storeKey, err := d.fetchEncryptionKey(docID, fieldID)
	if err != nil {
		return nil, err
	}

	if len(encryptionKey) == 0 {
		if !d.shouldGenerateKey {
			return plainText, nil
		}

		encryptionKey, err = generateEncryptionKeyFunc()
		if err != nil {
			return nil, err
		}

		err = d.store.Put(d.ctx, storeKey.ToDS(), encryptionKey)
		if err != nil {
			return nil, err
		}
	}
	return EncryptAES(plainText, encryptionKey)
}

func (d *DocEncryptor) Decrypt(docID string, fieldID uint32, cipherText []byte) ([]byte, error) {
	encKey, _, err := d.fetchEncryptionKey(docID, fieldID)
	if err != nil {
		return nil, err
	}
	if len(encKey) == 0 {
		return nil, nil
	}
	return DecryptAES(cipherText, encKey)
}

// fetchEncryptionKey fetches the encryption key for the given docID and fieldID.
// If the key is not found, it returns an empty key.
func (d *DocEncryptor) fetchEncryptionKey(docID string, fieldID uint32) ([]byte, core.EncStoreDocKey, error) {
	storeKey := core.NewEncStoreDocKey(docID, fieldID)
	if d.store == nil {
		return nil, core.EncStoreDocKey{}, ErrNoStorageProvided
	}
	encryptionKey, err := d.store.Get(d.ctx, storeKey.ToDS())
	isNotFound := errors.Is(err, ds.ErrNotFound)
	if err != nil && !isNotFound {
		return nil, core.EncStoreDocKey{}, err
	}
	return encryptionKey, storeKey, nil
}

func EncryptDoc(ctx context.Context, docID string, fieldID uint32, plainText []byte) ([]byte, error) {
	enc, ok := TryGetContextEncryptor(ctx)
	if !ok {
		return nil, nil
	}
	return enc.Encrypt(docID, fieldID, plainText)
}

func DecryptDoc(ctx context.Context, docID string, fieldID uint32, cipherText []byte) ([]byte, error) {
	enc, ok := TryGetContextEncryptor(ctx)
	if !ok {
		return nil, nil
	}
	return enc.Decrypt(docID, fieldID, cipherText)
}
