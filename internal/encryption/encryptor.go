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
	conf          immutable.Option[DocEncConfig]
	ctx           context.Context
	store         datastore.DSReaderWriter
	generatedKeys []core.EncStoreDocKey
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

// Encrypt encrypts the given plainText with the encryption key that is associated with the given docID,
// fieldName and key id.
func (d *DocEncryptor) Encrypt(
	encStoreKey core.EncStoreDocKey,
	plainText []byte,
) ([]byte, error) {
	if d.store == nil {
		return nil, ErrNoStorageProvided
	}

	encryptionKey, err := d.fetchByEncStoreKey(encStoreKey)
	if err != nil {
		return nil, err
	}

	var cipherText []byte
	if len(plainText) > 0 {
		cipherText, _, err = crypto.EncryptAES(plainText, encryptionKey, nil, true)
	}

	return cipherText, err
}

// Decrypt decrypts the given cipherText that is associated with the given docID and fieldName.
// If the corresponding encryption key is not found, it returns nil.
func (d *DocEncryptor) Decrypt(
	encStoreKey core.EncStoreDocKey,
	cipherText []byte,
) ([]byte, error) {
	if d.store == nil {
		return nil, ErrNoStorageProvided
	}
	encKey, err := d.fetchByEncStoreKey(encStoreKey)
	if err != nil {
		return nil, err
	}
	if len(encKey) == 0 {
		return nil, nil
	}
	return crypto.DecryptAES(nil, cipherText, encKey, nil)
}

func (d *DocEncryptor) fetchByEncStoreKey(storeKey core.EncStoreDocKey) ([]byte, error) {
	encryptionKey, err := d.store.Get(d.ctx, storeKey.ToDS())
	isNotFound := errors.Is(err, ds.ErrNotFound)
	if err != nil {
		if isNotFound {
			return nil, nil
		}
		return nil, err
	}

	return encryptionKey, nil
}

func (d *DocEncryptor) storeByEncStoreKey(storeKey core.EncStoreDocKey, encryptionKey []byte) error {
	return d.store.Put(d.ctx, storeKey.ToDS(), encryptionKey)
}

// GetKey returns the encryption key for the given docID, (optional) fieldName and block height.
func (d *DocEncryptor) GetKey(encStoreKey core.EncStoreDocKey) ([]byte, error) {
	if d.store == nil {
		return nil, ErrNoStorageProvided
	}
	encryptionKey, err := d.fetchByEncStoreKey(encStoreKey)
	if err != nil {
		return nil, err
	}
	return encryptionKey, nil
}

// getGeneratedKeyFor returns the generated key for the given docID and fieldName.
func (d *DocEncryptor) getGeneratedKeyFor(
	docID string,
	fieldName immutable.Option[string],
) (immutable.Option[core.EncStoreDocKey], []byte) {
	for _, key := range d.generatedKeys {
		if key.DocID == docID && key.FieldName == fieldName {
			fetchByEncStoreKey, err := d.fetchByEncStoreKey(key)
			if err != nil {
				return immutable.None[core.EncStoreDocKey](), nil
			}
			return immutable.Some(key), fetchByEncStoreKey
		}
	}
	return immutable.None[core.EncStoreDocKey](), nil
}

// GetOrGenerateEncryptionKey returns the generated encryption key for the given docID, (optional) fieldName.
// If the key is not generated before, it generates a new key and stores it.
func (d *DocEncryptor) GetOrGenerateEncryptionKey(
	docID string,
	fieldName immutable.Option[string],
) (immutable.Option[core.EncStoreDocKey], []byte, error) {
	encStoreKey, encryptionKey := d.getGeneratedKeyFor(docID, fieldName)
	if encStoreKey.HasValue() {
		return encStoreKey, encryptionKey, nil
	}

	return d.generateEncryptionKey(docID, fieldName)
}

// generateEncryptionKey generates a new encryption key for the given docID and fieldName.
func (d *DocEncryptor) generateEncryptionKey(
	docID string,
	fieldName immutable.Option[string],
) (immutable.Option[core.EncStoreDocKey], []byte, error) {
	encStoreKey := core.NewEncStoreDocKey(docID, fieldName, "")
	if !shouldEncryptIndividualField(d.conf, fieldName) {
		encStoreKey.FieldName = immutable.None[string]()
	}

	if !shouldEncryptDocField(d.conf, encStoreKey.FieldName) {
		return immutable.None[core.EncStoreDocKey](), nil, nil
	}

	encryptionKey, err := generateEncryptionKeyFunc(encStoreKey.DocID, encStoreKey.FieldName)
	if err != nil {
		return immutable.None[core.EncStoreDocKey](), nil, err
	}

	keyID, err := crypto.GenerateCid(encryptionKey)
	if err != nil {
		return immutable.None[core.EncStoreDocKey](), nil, err
	}
	encStoreKey.KeyID = keyID.String()

	err = d.storeByEncStoreKey(encStoreKey, encryptionKey)
	if err != nil {
		return immutable.None[core.EncStoreDocKey](), nil, err
	}

	d.generatedKeys = append(d.generatedKeys, encStoreKey)

	return immutable.Some(encStoreKey), encryptionKey, nil
}

// SaveKey saves the given encryption key for the given docID, (optional) fieldName and block height.
func (d *DocEncryptor) SaveKey(encStoreKey core.EncStoreDocKey, encryptionKey []byte) error {
	if d.store == nil {
		return ErrNoStorageProvided
	}
	return d.storeByEncStoreKey(encStoreKey, encryptionKey)
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

func init() {
	arg := os.Args[0]
	// If the binary is a test binary, use a deterministic nonce.
	// TODO: We should try to find a better way to detect this https://github.com/sourcenetwork/defradb/issues/2801
	if strings.HasSuffix(arg, ".test") ||
		strings.Contains(arg, "/defradb/tests/") ||
		strings.Contains(arg, "/__debug_bin") {
		generateEncryptionKeyFunc = generateTestEncryptionKey
	}
}
