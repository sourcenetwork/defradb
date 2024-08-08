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
	"os"
	"strings"

	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
)

var generateEncryptionKeyFunc = generateEncryptionKey

const keyLength = 32 // 32 bytes for AES-256

const testEncryptionKey = "examplekey1234567890examplekey12"

// generateEncryptionKey generates a random AES key.
func generateEncryptionKey(_ string, _ immutable.Option[string]) ([]byte, error) {
	key := make([]byte, keyLength)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

// generateTestEncryptionKey generates a deterministic encryption key for testing.
// While testing, we also want to make sure different keys are generated for different docs and fields
// and that's why we use the docID and fieldName to generate the key.
func generateTestEncryptionKey(docID string, fieldName immutable.Option[string]) ([]byte, error) {
	return []byte(fieldName.Value() + docID + testEncryptionKey)[0:keyLength], nil
}

// DocEncryptor is a document encryptor that encrypts and decrypts individual document fields.
// It acts based on the configuration [DocEncConfig] provided and data stored in the provided store.
// It uses [core.EncStoreDocKey] to store and retrieve encryption keys.
type DocEncryptor struct {
	conf  immutable.Option[DocEncConfig]
	ctx   context.Context
	store datastore.DSReaderWriter
	cache map[core.EncStoreDocKey][]byte
}

func newDocEncryptor(ctx context.Context) *DocEncryptor {
	return &DocEncryptor{ctx: ctx, cache: make(map[core.EncStoreDocKey][]byte)}
}

// SetConfig sets the configuration for the document encryptor.
func (d *DocEncryptor) SetConfig(conf immutable.Option[DocEncConfig]) {
	d.conf = conf
}

// SetStore sets the store for the document encryptor.
func (d *DocEncryptor) SetStore(store datastore.DSReaderWriter) {
	d.store = store
}

func shouldEncryptIndividualField(conf immutable.Option[DocEncConfig], fieldName immutable.Option[string]) bool {
	if !conf.HasValue() || !fieldName.HasValue() {
		return false
	}
	for _, field := range conf.Value().EncryptedFields {
		if field == fieldName.Value() {
			return true
		}
	}
	return false
}

func shouldEncryptDocField(conf immutable.Option[DocEncConfig], fieldName immutable.Option[string]) bool {
	if !conf.HasValue() {
		return false
	}
	if conf.Value().IsDocEncrypted {
		return true
	}
	if !fieldName.HasValue() {
		return false
	}
	for _, field := range conf.Value().EncryptedFields {
		if field == fieldName.Value() {
			return true
		}
	}
	return false
}

// Encrypt encrypts the given plainText that is associated with the given docID, fieldName and block height.
// If the current configuration is set to encrypt the given key individually, it will encrypt it with a new key.
// Otherwise, it will use document-level encryption key.
func (d *DocEncryptor) Encrypt(
	docID string,
	fieldName immutable.Option[string],
	blockHeight uint64,
	plainText []byte,
) ([]byte, error) {
	if d.store == nil {
		return nil, ErrNoStorageProvided
	}
	encryptionKey, err := d.fetchByEncStoreKey(core.NewEncStoreDocKey(docID, fieldName, blockHeight))
	if err != nil {
		return nil, err
	}

	if len(encryptionKey) == 0 {
		if !shouldEncryptIndividualField(d.conf, fieldName) {
			fieldName = immutable.None[string]()
		}

		if !shouldEncryptDocField(d.conf, fieldName) {
			return plainText, nil
		}

		encryptionKey, err = generateEncryptionKeyFunc(docID, fieldName)
		if err != nil {
			return nil, err
		}

		err = d.storeByEncStoreKey(core.NewEncStoreDocKey(docID, fieldName, blockHeight), encryptionKey)
		if err != nil {
			return nil, err
		}
	}
	cipherText, _, err := crypto.EncryptAES(plainText, encryptionKey, nil, true)
	return cipherText, err
}

// Decrypt decrypts the given cipherText that is associated with the given docID and fieldName.
// If the corresponding encryption key is not found, it returns nil.
func (d *DocEncryptor) Decrypt(
	docID string,
	fieldName immutable.Option[string],
	blockHeight uint64,
	cipherText []byte,
) ([]byte, error) {
	if d.store == nil {
		return nil, ErrNoStorageProvided
	}
	encKey, err := d.fetchByEncStoreKey(core.NewEncStoreDocKey(docID, fieldName, blockHeight))
	if err != nil {
		return nil, err
	}
	if len(encKey) == 0 {
		return nil, nil
	}
	return crypto.DecryptAES(nil, cipherText, encKey, nil)
}

func (d *DocEncryptor) fetchByEncStoreKey(storeKey core.EncStoreDocKey) ([]byte, error) {
	if encryptionKey, ok := d.cache[storeKey]; ok {
		return encryptionKey, nil
	}
	encryptionKey, err := d.store.Get(d.ctx, storeKey.ToDS())
	isNotFound := errors.Is(err, ds.ErrNotFound)
	if err != nil {
		if isNotFound {
			return nil, nil
		}
		return nil, err
	}

	d.cache[storeKey] = encryptionKey
	return encryptionKey, nil
}

func (d *DocEncryptor) storeByEncStoreKey(storeKey core.EncStoreDocKey, encryptionKey []byte) error {
	d.cache[storeKey] = encryptionKey
	return d.store.Put(d.ctx, storeKey.ToDS(), encryptionKey)
}

// GetKey returns the encryption key for the given docID, (optional) fieldName and block height.
func (d *DocEncryptor) GetKey(docID string, fieldName immutable.Option[string], blockHeight uint64) ([]byte, error) {
	if d.store == nil {
		return nil, ErrNoStorageProvided
	}
	encryptionKey, err := d.fetchByEncStoreKey(core.NewEncStoreDocKey(docID, fieldName, blockHeight))
	if err != nil {
		return nil, err
	}
	return encryptionKey, nil
}

// SaveKey saves the given encryption key for the given docID, (optional) fieldName and block height.
func (d *DocEncryptor) SaveKey(
	docID string,
	fieldName immutable.Option[string],
	blockHeight uint64,
	encryptionKey []byte,
) error {
	if d.store == nil {
		return ErrNoStorageProvided
	}
	return d.storeByEncStoreKey(core.NewEncStoreDocKey(docID, fieldName, blockHeight), encryptionKey)
}

// EncryptDoc encrypts the given plainText that is associated with the given docID, fieldName and block height with
// encryptor in the context.
// If the current configuration is set to encrypt the given key individually, it will encrypt it with a new key.
// Otherwise, it will use document-level encryption key.
func EncryptDoc(
	ctx context.Context,
	docID string,
	fieldName immutable.Option[string],
	blockHeight uint64,
	plainText []byte,
) ([]byte, error) {
	enc, ok := TryGetContextEncryptor(ctx)
	if !ok {
		return nil, nil
	}
	return enc.Encrypt(docID, fieldName, blockHeight, plainText)
}

// DecryptDoc decrypts the given cipherText that is associated with the given docID and fieldName with
// encryptor in the context.
// If fieldName is not provided, it will try to decrypt with the document-level key. Otherwise, it will try to
// decrypt with the field-level key.
func DecryptDoc(
	ctx context.Context,
	docID string,
	fieldName immutable.Option[string],
	blockHeight uint64,
	cipherText []byte,
) ([]byte, error) {
	enc, ok := TryGetContextEncryptor(ctx)
	if !ok {
		return nil, nil
	}
	return enc.Decrypt(docID, fieldName, blockHeight, cipherText)
}

// ShouldEncryptDocField returns true if the given field should be encrypted based on the context config.
func ShouldEncryptDocField(ctx context.Context, fieldName immutable.Option[string]) bool {
	return shouldEncryptDocField(GetContextConfig(ctx), fieldName)
}

// ShouldEncryptIndividualField returns true if the given field should be encrypted individually based on
// the context config.
func ShouldEncryptIndividualField(ctx context.Context, fieldName immutable.Option[string]) bool {
	return shouldEncryptIndividualField(GetContextConfig(ctx), fieldName)
}

// SaveKey saves the given encryption key for the given docID, (optional) fieldName and block height with
// encryptor in the context.
func SaveKey(
	ctx context.Context,
	docID string,
	fieldName immutable.Option[string],
	blockHeight uint64,
	encryptionKey []byte,
) error {
	enc, ok := TryGetContextEncryptor(ctx)
	if !ok {
		return nil
	}
	return enc.SaveKey(docID, fieldName, blockHeight, encryptionKey)
}

// GetKey returns the encryption key for the given docID, (optional) fieldName and block height with encryptor
// in the context.
func GetKey(ctx context.Context, docID string, fieldName immutable.Option[string], blockHeight uint64) ([]byte, error) {
	enc, ok := TryGetContextEncryptor(ctx)
	if !ok {
		return nil, nil
	}
	return enc.GetKey(docID, fieldName, blockHeight)
}

func init() {
	arg := os.Args[0]
	// If the binary is a test binary, use a deterministic nonce.
	// TODO: We should try to find a better way to detect this https://github.com/sourcenetwork/defradb/issues/2801
	if strings.HasSuffix(arg, ".test") || strings.Contains(arg, "/defradb/tests/") {
		generateEncryptionKeyFunc = generateTestEncryptionKey
	}
}
