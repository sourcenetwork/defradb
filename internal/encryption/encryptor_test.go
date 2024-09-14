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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/immutable"
)

func TestContext_NoEncryptor_ReturnsNil(t *testing.T) {
	ctx := context.Background()
	enc := GetEncryptorFromContext(ctx)
	assert.Nil(t, enc)
}

func TestContext_WithEncryptor_ReturnsEncryptor(t *testing.T) {
	ctx := context.Background()
	enc := newDocEncryptor(ctx)
	ctx = context.WithValue(ctx, docEncContextKey{}, enc)

	retrievedEnc := GetEncryptorFromContext(ctx)
	assert.NotNil(t, retrievedEnc)
	assert.Equal(t, enc, retrievedEnc)
}

func TestContext_EnsureEncryptor_CreatesNew(t *testing.T) {
	ctx := context.Background()
	newCtx, enc := EnsureContextWithEncryptor(ctx)

	assert.NotNil(t, enc)
	assert.NotEqual(t, ctx, newCtx)

	retrievedEnc := GetEncryptorFromContext(newCtx)
	assert.Equal(t, enc, retrievedEnc)
}

func TestContext_EnsureEncryptor_ReturnsExisting(t *testing.T) {
	ctx := context.Background()
	enc := newDocEncryptor(ctx)
	ctx = context.WithValue(ctx, docEncContextKey{}, enc)

	newCtx, retrievedEnc := EnsureContextWithEncryptor(ctx)
	assert.Equal(t, ctx, newCtx)
	assert.Equal(t, enc, retrievedEnc)
}

func TestConfig_GetFromContext_NoConfig(t *testing.T) {
	ctx := context.Background()
	config := GetContextConfig(ctx)
	assert.False(t, config.HasValue())
}

func TestConfig_GetFromContext_ReturnCurrentConfig(t *testing.T) {
	ctx := context.Background()
	expectedConfig := DocEncConfig{IsDocEncrypted: true, EncryptedFields: []string{"field1", "field2"}}
	ctx = context.WithValue(ctx, configContextKey{}, expectedConfig)

	config := GetContextConfig(ctx)
	assert.True(t, config.HasValue())
	assert.Equal(t, expectedConfig, config.Value())
}

func TestConfig_SetContextConfig_StoreConfig(t *testing.T) {
	ctx := context.Background()
	config := DocEncConfig{IsDocEncrypted: true, EncryptedFields: []string{"field1", "field2"}}

	newCtx := SetContextConfig(ctx, config)
	retrievedConfig := GetContextConfig(newCtx)

	assert.True(t, retrievedConfig.HasValue())
	assert.Equal(t, config, retrievedConfig.Value())
}

func TestConfig_SetFromParamsWithDocEncryption_StoreConfig(t *testing.T) {
	ctx := context.Background()
	newCtx := SetContextConfigFromParams(ctx, true, []string{"field1", "field2"})

	config := GetContextConfig(newCtx)
	assert.True(t, config.HasValue())
	assert.True(t, config.Value().IsDocEncrypted)
	assert.Equal(t, []string{"field1", "field2"}, config.Value().EncryptedFields)
}

func TestConfig_SetFromParamsWithFields_StoreConfig(t *testing.T) {
	ctx := context.Background()
	newCtx := SetContextConfigFromParams(ctx, false, []string{"field1", "field2"})

	config := GetContextConfig(newCtx)
	assert.True(t, config.HasValue())
	assert.False(t, config.Value().IsDocEncrypted)
	assert.Equal(t, []string{"field1", "field2"}, config.Value().EncryptedFields)
}

func TestConfig_SetFromParamsWithNoEncryptionSetting_NoConfig(t *testing.T) {
	ctx := context.Background()
	newCtx := SetContextConfigFromParams(ctx, false, nil)

	config := GetContextConfig(newCtx)
	assert.False(t, config.HasValue())
}

func TestEncryptor_EncryptDecrypt_SuccessfulRoundTrip(t *testing.T) {
	ctx := context.Background()
	enc := newDocEncryptor(ctx)
	enc.SetConfig(immutable.Some(DocEncConfig{EncryptedFields: []string{"field1"}}))

	plainText := []byte("Hello, World!")
	docID := "doc1"
	fieldName := immutable.Some("field1")

	key, err := enc.GetOrGenerateEncryptionKey(docID, fieldName)
	assert.NoError(t, err)
	assert.NotNil(t, key)

	cipherText, err := enc.Encrypt(plainText, key)
	assert.NoError(t, err)
	assert.NotEqual(t, plainText, cipherText)

	decryptedText, err := enc.Decrypt(cipherText, key)
	assert.NoError(t, err)
	assert.Equal(t, plainText, decryptedText)
}

func TestEncryptor_GetOrGenerateKey_ReturnsExistingKey(t *testing.T) {
	ctx := context.Background()
	enc := newDocEncryptor(ctx)
	enc.SetConfig(immutable.Some(DocEncConfig{EncryptedFields: []string{"field1"}}))

	docID := "doc1"
	fieldName := immutable.Some("field1")

	key1, err := enc.GetOrGenerateEncryptionKey(docID, fieldName)
	assert.NoError(t, err)
	assert.NotNil(t, key1)

	key2, err := enc.GetOrGenerateEncryptionKey(docID, fieldName)
	assert.NoError(t, err)
	assert.Equal(t, key1, key2)
}

func TestEncryptor_GenerateKey_DifferentKeysForDifferentFields(t *testing.T) {
	ctx := context.Background()
	enc := newDocEncryptor(ctx)
	enc.SetConfig(immutable.Some(DocEncConfig{EncryptedFields: []string{"field1", "field2"}}))

	docID := "doc1"
	fieldName1 := immutable.Some("field1")
	fieldName2 := immutable.Some("field2")

	key1, err := enc.GetOrGenerateEncryptionKey(docID, fieldName1)
	assert.NoError(t, err)
	assert.NotNil(t, key1)

	key2, err := enc.GetOrGenerateEncryptionKey(docID, fieldName2)
	assert.NoError(t, err)
	assert.NotNil(t, key2)

	assert.NotEqual(t, key1, key2)
}

func TestShouldEncryptField_WithDocEncryption_True(t *testing.T) {
	config := DocEncConfig{IsDocEncrypted: true}
	ctx := SetContextConfig(context.Background(), config)

	assert.True(t, ShouldEncryptDocField(ctx, immutable.Some("field1")))
	assert.True(t, ShouldEncryptDocField(ctx, immutable.Some("field2")))
}

func TestShouldEncryptField_WithFieldEncryption_TrueForMatchingField(t *testing.T) {
	config := DocEncConfig{EncryptedFields: []string{"field1"}}
	ctx := SetContextConfig(context.Background(), config)

	assert.True(t, ShouldEncryptDocField(ctx, immutable.Some("field1")))
	assert.False(t, ShouldEncryptDocField(ctx, immutable.Some("field2")))
}

func TestShouldEncryptIndividualField_WithDocEncryption_False(t *testing.T) {
	config := DocEncConfig{IsDocEncrypted: true}
	ctx := SetContextConfig(context.Background(), config)

	assert.False(t, ShouldEncryptIndividualField(ctx, immutable.Some("field1")))
	assert.False(t, ShouldEncryptIndividualField(ctx, immutable.Some("field2")))
}

func TestShouldEncryptIndividualField_WithFieldEncryption_TrueForMatchingField(t *testing.T) {
	config := DocEncConfig{EncryptedFields: []string{"field1"}}
	ctx := SetContextConfig(context.Background(), config)

	assert.True(t, ShouldEncryptIndividualField(ctx, immutable.Some("field1")))
	assert.False(t, ShouldEncryptIndividualField(ctx, immutable.Some("field2")))
}
