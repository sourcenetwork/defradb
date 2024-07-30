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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
)

var generateEncryptionKeyFunc = generateEncryptionKey

const keyLength = 32 // 32 bytes for AES-256

const testEncryptionKey = "examplekey1234567890examplekey12"

// generateEncryptionKey generates a random AES key.
func generateEncryptionKey(_, _ string) ([]byte, error) {
	key := make([]byte, keyLength)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

// generateTestEncryptionKey generates a deterministic encryption key for testing.
// While testing, we also want to make sure different keys are generated for different docs and fields
// and that's why we use the docID and fieldName to generate the key.
func generateTestEncryptionKey(docID, fieldName string) ([]byte, error) {
	return []byte(fieldName + docID + testEncryptionKey)[0:keyLength], nil
}

// DocEncryptor is a document encryptor that encrypts and decrypts individual document fields.
// It acts based on the configuration [DocEncConfig] provided and data stored in the provided store.
// It uses [core.EncStoreDocKey] to store and retrieve encryption keys.
type DocEncryptor struct {
	conf  immutable.Option[DocEncConfig]
	ctx   context.Context
	store datastore.DSReaderWriter
}

func newDocEncryptor(ctx context.Context) *DocEncryptor {
	return &DocEncryptor{ctx: ctx}
}

// SetConfig sets the configuration for the document encryptor.
func (d *DocEncryptor) SetConfig(conf immutable.Option[DocEncConfig]) {
	d.conf = conf
}

// SetStore sets the store for the document encryptor.
func (d *DocEncryptor) SetStore(store datastore.DSReaderWriter) {
	d.store = store
}

func shouldEncryptIndividualField(conf immutable.Option[DocEncConfig], fieldName string) bool {
	if !conf.HasValue() || fieldName == "" {
		return false
	}
	for _, field := range conf.Value().EncryptedFields {
		if field == fieldName {
			return true
		}
	}
	return false
}

func shouldEncryptField(conf immutable.Option[DocEncConfig], fieldName string) bool {
	if !conf.HasValue() {
		return false
	}
	if conf.Value().IsDocEncrypted {
		return true
	}
	if fieldName == "" {
		return false
	}
	for _, field := range conf.Value().EncryptedFields {
		if field == fieldName {
			return true
		}
	}
	return false
}

// Encrypt encrypts the given plainText that is associated with the given docID and fieldName.
// If the current configuration is set to encrypt the given key individually, it will encrypt it with a new key.
// Otherwise, it will use document-level encryption key.
func (d *DocEncryptor) Encrypt(docID, fieldName string, plainText []byte) ([]byte, error) {
	encryptionKey, err := d.fetchEncryptionKey(docID, fieldName)
	if err != nil {
		return nil, err
	}

	if len(encryptionKey) == 0 {
		if !shouldEncryptIndividualField(d.conf, fieldName) {
			fieldName = ""
		}

		if !shouldEncryptField(d.conf, fieldName) {
			return plainText, nil
		}

		encryptionKey, err = generateEncryptionKeyFunc(docID, fieldName)
		if err != nil {
			return nil, err
		}

		storeKey := core.NewEncStoreDocKey(docID, fieldName)
		err = d.store.Put(d.ctx, storeKey.ToDS(), encryptionKey)
		if err != nil {
			return nil, err
		}
	}
	return EncryptAES(plainText, encryptionKey)
}

// Decrypt decrypts the given cipherText that is associated with the given docID and fieldName.
// If the corresponding encryption key is not found, it returns nil.
func (d *DocEncryptor) Decrypt(docID, fieldName string, cipherText []byte) ([]byte, error) {
	encKey, err := d.fetchEncryptionKey(docID, fieldName)
	if err != nil {
		return nil, err
	}
	if len(encKey) == 0 {
		return nil, nil
	}
	return DecryptAES(cipherText, encKey)
}

// fetchEncryptionKey fetches the encryption key for the given docID and fieldName.
// If the key is not found, it returns an empty key.
func (d *DocEncryptor) fetchEncryptionKey(docID string, fieldName string) ([]byte, error) {
	if d.store == nil {
		return nil, ErrNoStorageProvided
	}
	// first we try to find field-level key
	storeKey := core.NewEncStoreDocKey(docID, fieldName)
	encryptionKey, err := d.store.Get(d.ctx, storeKey.ToDS())
	isNotFound := errors.Is(err, ds.ErrNotFound)
	if err != nil {
		if !isNotFound {
			return nil, err
		}
		// if previous fetch was for doc-level, there is nothing else to look for
		if fieldName == "" {
			return nil, nil
		}
		if shouldEncryptIndividualField(d.conf, fieldName) {
			return nil, nil
		}
		// try to find doc-level key
		storeKey.FieldName = ""
		encryptionKey, err = d.store.Get(d.ctx, storeKey.ToDS())
		isNotFound = errors.Is(err, ds.ErrNotFound)
		if err != nil && !isNotFound {
			return nil, err
		}
	}
	return encryptionKey, nil
}

func (d *DocEncryptor) GetDocKey(docID string, optFieldName immutable.Option[string]) ([]byte, error) {
	if d.store == nil {
		return nil, ErrNoStorageProvided
	}
	fieldName := ""
	if optFieldName.HasValue() {
		fieldName = optFieldName.Value()
	}
	storeKey := core.NewEncStoreDocKey(docID, fieldName)
	encryptionKey, err := d.store.Get(d.ctx, storeKey.ToDS())
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return nil, err
	}
	return encryptionKey, nil
}

func (d *DocEncryptor) SaveDocKey(docID string, encryptionKey []byte) error {
	if d.store == nil {
		return ErrNoStorageProvided
	}
	storeKey := core.NewEncStoreDocKey(docID, "")
	return d.store.Put(d.ctx, storeKey.ToDS(), encryptionKey)
}

// EncryptDoc encrypts the given plainText that is associated with the given docID and fieldName with
// encryptor in the context.
// If the current configuration is set to encrypt the given key individually, it will encrypt it with a new key.
// Otherwise, it will use document-level encryption key.
func EncryptDoc(ctx context.Context, docID string, fieldName string, plainText []byte) ([]byte, error) {
	enc, ok := TryGetContextEncryptor(ctx)
	if !ok {
		return nil, nil
	}
	return enc.Encrypt(docID, fieldName, plainText)
}

// DecryptDoc decrypts the given cipherText that is associated with the given docID and fieldName with
// encryptor in the context.
func DecryptDoc(ctx context.Context, docID string, fieldName string, cipherText []byte) ([]byte, error) {
	enc, ok := TryGetContextEncryptor(ctx)
	if !ok {
		return nil, nil
	}
	return enc.Decrypt(docID, fieldName, cipherText)
}

// ShouldEncryptField returns true if the given field should be encrypted based on the context config.
func ShouldEncryptField(ctx context.Context, fieldName string) bool {
	return shouldEncryptField(GetContextConfig(ctx), fieldName)
}

// ShouldEncryptIndividualField returns true if the given field should be encrypted individually based on
// the context config.
func ShouldEncryptIndividualField(ctx context.Context, fieldName string) bool {
	return shouldEncryptIndividualField(GetContextConfig(ctx), fieldName)
}

// SaveDocKey saves the given encryption key for the given docID and (optional) fieldName with encryptor in the context.
func SaveDocKey(ctx context.Context, docID string, fieldName immutable.Option[string], encryptionKey []byte) error {
	enc, ok := TryGetContextEncryptor(ctx)
	if !ok {
		return nil
	}
	return enc.SaveDocKey(docID, encryptionKey)
}

// GetDocKey returns the encryption key for the given docID and (optional) fieldName with encryptor in the context.
func GetDocKey(ctx context.Context, docID string, fieldName immutable.Option[string]) ([]byte, error) {
	enc, ok := TryGetContextEncryptor(ctx)
	if !ok {
		return nil, nil
	}
	return enc.GetDocKey(docID, fieldName)
}
