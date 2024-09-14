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
	"io"
	"os"
	"strings"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/crypto"
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
// DocEncryptor is a session-bound, i.e. once a user requests to create (or update) a document or a node
// receives an UpdateEvent on a document (or any other event) a new DocEncryptor is created and stored
// in the context, so that the same DocEncryptor can be used by other object down the call chain.
type DocEncryptor struct {
	conf          immutable.Option[DocEncConfig]
	ctx           context.Context
	generatedKeys map[genK][]byte
}

type genK struct {
	docID     string
	fieldName immutable.Option[string]
}

func newDocEncryptor(ctx context.Context) *DocEncryptor {
	return &DocEncryptor{ctx: ctx, generatedKeys: make(map[genK][]byte)}
}

// SetConfig sets the configuration for the document encryptor.
func (d *DocEncryptor) SetConfig(conf immutable.Option[DocEncConfig]) {
	d.conf = conf
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
	plainText, encryptionKey []byte,
) ([]byte, error) {
	var cipherText []byte
	var err error
	if len(plainText) > 0 {
		cipherText, _, err = crypto.EncryptAES(plainText, encryptionKey, nil, true)
	}

	return cipherText, err
}

// Decrypt decrypts the given cipherText that is associated with the given docID and fieldName.
// If the corresponding encryption key is not found, it returns nil.
func (d *DocEncryptor) Decrypt(
	cipherText, encKey []byte,
) ([]byte, error) {
	if len(encKey) == 0 {
		return nil, nil
	}
	return crypto.DecryptAES(nil, cipherText, encKey, nil)
}

// getGeneratedKeyFor returns the generated key for the given docID and fieldName.
func (d *DocEncryptor) getGeneratedKeyFor(
	docID string,
	fieldName immutable.Option[string],
) []byte {
	return d.generatedKeys[genK{docID, fieldName}]
}

// GetOrGenerateEncryptionKey returns the generated encryption key for the given docID, (optional) fieldName.
// If the key is not generated before, it generates a new key and stores it.
func (d *DocEncryptor) GetOrGenerateEncryptionKey(
	docID string,
	fieldName immutable.Option[string],
) ([]byte, error) {
	encryptionKey := d.getGeneratedKeyFor(docID, fieldName)
	if len(encryptionKey) > 0 {
		return encryptionKey, nil
	}

	return d.generateEncryptionKey(docID, fieldName)
}

// generateEncryptionKey generates a new encryption key for the given docID and fieldName.
func (d *DocEncryptor) generateEncryptionKey(
	docID string,
	fieldName immutable.Option[string],
) ([]byte, error) {
	if !shouldEncryptIndividualField(d.conf, fieldName) {
		fieldName = immutable.None[string]()
	}

	if !shouldEncryptDocField(d.conf, fieldName) {
		return nil, nil
	}

	encryptionKey, err := generateEncryptionKeyFunc(docID, fieldName)
	if err != nil {
		return nil, err
	}

	d.generatedKeys[genK{docID, fieldName}] = encryptionKey

	return encryptionKey, nil
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
