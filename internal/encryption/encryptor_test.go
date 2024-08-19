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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/datastore/mocks"
	"github.com/sourcenetwork/defradb/internal/core"
)

var testErr = errors.New("test error")

const docID = "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3"

var fieldName = immutable.Some("name")
var noFieldName = immutable.None[string]()

func getPlainText() []byte {
	return []byte("test")
}

func getEncKey(fieldName immutable.Option[string]) []byte {
	key, _ := generateTestEncryptionKey(docID, fieldName)
	return key
}

func getKeyID(fieldName immutable.Option[string]) string {
	key := getEncKey(fieldName)
	cid, _ := crypto.GenerateCid(key)
	return cid.String()
}

func makeStoreKey(docID string, fieldName immutable.Option[string]) core.EncStoreDocKey {
	return core.NewEncStoreDocKey(docID, fieldName, getKeyID(fieldName))
}

func getCipherText(t *testing.T, fieldName immutable.Option[string]) []byte {
	cipherText, _, err := crypto.EncryptAES(getPlainText(), getEncKey(fieldName), nil, true)
	assert.NoError(t, err)
	return cipherText
}

func newDefaultEncryptor(t *testing.T) (*DocEncryptor, *mocks.DSReaderWriter) {
	return newEncryptorWithConfig(t, DocEncConfig{IsDocEncrypted: true})
}

func newEncryptorWithConfig(t *testing.T, conf DocEncConfig) (*DocEncryptor, *mocks.DSReaderWriter) {
	enc := newDocEncryptor(context.Background())
	st := mocks.NewDSReaderWriter(t)
	enc.SetConfig(immutable.Some(conf))
	enc.SetStore(st)
	return enc, st
}

func TestEncryptorEncrypt_IfStorageReturnsError_Error(t *testing.T) {
	enc, st := newDefaultEncryptor(t)

	st.EXPECT().Get(mock.Anything, mock.Anything).Return(nil, testErr)

	_, err := enc.Encrypt(makeStoreKey(docID, fieldName), []byte("test"))

	assert.ErrorIs(t, err, testErr)
}

func TestEncryptorEncrypt_IfKeyWithFieldFoundInStorage_ShouldUseItToReturnCipherText(t *testing.T) {
	enc, st := newEncryptorWithConfig(t, DocEncConfig{EncryptedFields: []string{fieldName.Value()}})

	storeKey := makeStoreKey(docID, fieldName)
	st.EXPECT().Get(mock.Anything, storeKey.ToDS()).Return(getEncKey(fieldName), nil)

	cipherText, err := enc.Encrypt(makeStoreKey(docID, fieldName), getPlainText())

	assert.NoError(t, err)
	assert.Equal(t, getCipherText(t, fieldName), cipherText)
}

func TestEncryptorEncrypt_IfKeyFoundInStorage_ShouldUseItToReturnCipherText(t *testing.T) {
	enc, st := newDefaultEncryptor(t)

	st.EXPECT().Get(mock.Anything, mock.Anything).Return(getEncKey(noFieldName), nil)

	cipherText, err := enc.Encrypt(makeStoreKey(docID, noFieldName), getPlainText())

	assert.NoError(t, err)
	assert.Equal(t, getCipherText(t, noFieldName), cipherText)
}

func TestEncryptorEncrypt_IfStorageFailsToStoreEncryptionKey_ReturnError(t *testing.T) {
	enc, st := newDefaultEncryptor(t)

	st.EXPECT().Put(mock.Anything, mock.Anything, mock.Anything).Return(testErr)

	_, _, err := enc.GetOrGenerateEncryptionKey(docID, noFieldName)
	assert.ErrorIs(t, err, testErr)
}

func TestEncryptorEncrypt_IfKeyGenerationIsNotEnabled_ShouldReturnPlainText(t *testing.T) {
	enc, _ := newDefaultEncryptor(t)
	enc.SetConfig(immutable.None[DocEncConfig]())

	encStoreKey, encryptionKey, err := enc.GetOrGenerateEncryptionKey(docID, noFieldName)
	assert.NoError(t, err)
	assert.Len(t, encryptionKey, 0)
	assert.Equal(t, encStoreKey, immutable.None[core.EncStoreDocKey]())
}

func TestEncryptorEncrypt_IfNoStorageProvided_Error(t *testing.T) {
	enc, _ := newDefaultEncryptor(t)
	enc.SetStore(nil)

	_, err := enc.Encrypt(makeStoreKey(docID, fieldName), getPlainText())

	assert.ErrorIs(t, err, ErrNoStorageProvided)
}

func TestEncryptorDecrypt_IfNoStorageProvided_Error(t *testing.T) {
	enc, _ := newDefaultEncryptor(t)
	enc.SetStore(nil)

	_, err := enc.Decrypt(makeStoreKey(docID, fieldName), getPlainText())

	assert.ErrorIs(t, err, ErrNoStorageProvided)
}

func TestEncryptorDecrypt_IfStorageReturnsError_Error(t *testing.T) {
	enc, st := newDefaultEncryptor(t)

	st.EXPECT().Get(mock.Anything, mock.Anything).Return(nil, testErr)

	_, err := enc.Decrypt(makeStoreKey(docID, fieldName), []byte("test"))

	assert.ErrorIs(t, err, testErr)
}

func TestEncryptorDecrypt_IfKeyFoundInStorage_ShouldUseItToReturnPlainText(t *testing.T) {
	enc, st := newDefaultEncryptor(t)

	st.EXPECT().Get(mock.Anything, mock.Anything).Return(getEncKey(noFieldName), nil)

	plainText, err := enc.Decrypt(makeStoreKey(docID, fieldName), getCipherText(t, noFieldName))

	assert.NoError(t, err)
	assert.Equal(t, getPlainText(), plainText)
}

func TestEncryptDoc_IfContextHasNoEncryptor_ReturnNil(t *testing.T) {
	data, err := EncryptDoc(context.Background(), makeStoreKey(docID, fieldName), getPlainText())
	assert.Nil(t, data, "data should be nil")
	assert.NoError(t, err, "error should be nil")
}

func TestDecryptDoc_IfContextHasNoEncryptor_ReturnNil(t *testing.T) {
	data, err := DecryptDoc(
		context.Background(),
		makeStoreKey(docID, fieldName),
		getCipherText(t, fieldName),
	)
	assert.Nil(t, data, "data should be nil")
	assert.NoError(t, err, "error should be nil")
}
