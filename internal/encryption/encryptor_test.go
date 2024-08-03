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

	ds "github.com/ipfs/go-datastore"
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

func getCipherText(t *testing.T, fieldName immutable.Option[string]) []byte {
	cipherText, err := crypto.EncryptAES(getPlainText(), getEncKey(fieldName))
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

	_, err := enc.Encrypt(docID, fieldName, []byte("test"))

	assert.ErrorIs(t, err, testErr)
}

func TestEncryptorEncrypt_IfStorageReturnsErrorOnSecondCall_Error(t *testing.T) {
	enc, st := newDefaultEncryptor(t)

	st.EXPECT().Get(mock.Anything, mock.Anything).Return(nil, ds.ErrNotFound).Once()
	st.EXPECT().Get(mock.Anything, mock.Anything).Return(nil, testErr)

	_, err := enc.Encrypt(docID, fieldName, []byte("test"))

	assert.ErrorIs(t, err, testErr)
}

func TestEncryptorEncrypt_WithEmptyFieldNameIfNoKeyFoundInStorage_ShouldGenerateKeyStoreItAndReturnCipherText(t *testing.T) {
	enc, st := newDefaultEncryptor(t)

	storeKey := core.NewEncStoreDocKey(docID, noFieldName)

	st.EXPECT().Get(mock.Anything, storeKey.ToDS()).Return(nil, ds.ErrNotFound)
	st.EXPECT().Put(mock.Anything, storeKey.ToDS(), getEncKey(noFieldName)).Return(nil)

	cipherText, err := enc.Encrypt(docID, noFieldName, getPlainText())

	assert.NoError(t, err)
	assert.Equal(t, getCipherText(t, noFieldName), cipherText)
}

func TestEncryptorEncrypt_IfNoFieldEncRequestedAndNoKeyInStorage_GenerateKeyStoreItAndReturnCipherText(t *testing.T) {
	enc, st := newDefaultEncryptor(t)

	docStoreKey := core.NewEncStoreDocKey(docID, noFieldName).ToDS()
	fieldStoreKey := core.NewEncStoreDocKey(docID, fieldName).ToDS()

	st.EXPECT().Get(mock.Anything, fieldStoreKey).Return(nil, ds.ErrNotFound)
	st.EXPECT().Get(mock.Anything, docStoreKey).Return(nil, ds.ErrNotFound)
	st.EXPECT().Put(mock.Anything, docStoreKey, getEncKey(noFieldName)).Return(nil)

	cipherText, err := enc.Encrypt(docID, fieldName, getPlainText())

	assert.NoError(t, err)
	assert.Equal(t, getCipherText(t, noFieldName), cipherText)
}

func TestEncryptorEncrypt_IfNoKeyWithFieldFoundInStorage_ShouldGenerateKeyStoreItAndReturnCipherText(t *testing.T) {
	enc, st := newEncryptorWithConfig(t, DocEncConfig{EncryptedFields: []string{fieldName.Value()}})

	storeKey := core.NewEncStoreDocKey(docID, fieldName)

	st.EXPECT().Get(mock.Anything, storeKey.ToDS()).Return(nil, ds.ErrNotFound)
	st.EXPECT().Put(mock.Anything, storeKey.ToDS(), getEncKey(fieldName)).Return(nil)

	cipherText, err := enc.Encrypt(docID, fieldName, getPlainText())

	assert.NoError(t, err)
	assert.Equal(t, getCipherText(t, fieldName), cipherText)
}

func TestEncryptorEncrypt_IfKeyWithFieldFoundInStorage_ShouldUseItToReturnCipherText(t *testing.T) {
	enc, st := newEncryptorWithConfig(t, DocEncConfig{EncryptedFields: []string{fieldName.Value()}})

	storeKey := core.NewEncStoreDocKey(docID, fieldName)
	st.EXPECT().Get(mock.Anything, storeKey.ToDS()).Return(getEncKey(fieldName), nil)

	cipherText, err := enc.Encrypt(docID, fieldName, getPlainText())

	assert.NoError(t, err)
	assert.Equal(t, getCipherText(t, fieldName), cipherText)
}

func TestEncryptorEncrypt_IfKeyFoundInStorage_ShouldUseItToReturnCipherText(t *testing.T) {
	enc, st := newDefaultEncryptor(t)

	st.EXPECT().Get(mock.Anything, mock.Anything).Return(getEncKey(noFieldName), nil)

	cipherText, err := enc.Encrypt(docID, noFieldName, getPlainText())

	assert.NoError(t, err)
	assert.Equal(t, getCipherText(t, noFieldName), cipherText)
}

func TestEncryptorEncrypt_IfStorageFailsToStoreEncryptionKey_ReturnError(t *testing.T) {
	enc, st := newDefaultEncryptor(t)

	st.EXPECT().Get(mock.Anything, mock.Anything).Return(nil, ds.ErrNotFound)

	st.EXPECT().Put(mock.Anything, mock.Anything, mock.Anything).Return(testErr)

	_, err := enc.Encrypt(docID, fieldName, getPlainText())

	assert.ErrorIs(t, err, testErr)
}

func TestEncryptorEncrypt_IfKeyGenerationIsNotEnabled_ShouldReturnPlainText(t *testing.T) {
	enc, st := newDefaultEncryptor(t)
	enc.SetConfig(immutable.None[DocEncConfig]())

	st.EXPECT().Get(mock.Anything, mock.Anything).Return(nil, ds.ErrNotFound)

	cipherText, err := enc.Encrypt(docID, fieldName, getPlainText())

	assert.NoError(t, err)
	assert.Equal(t, getPlainText(), cipherText)
}

func TestEncryptorEncrypt_IfNoStorageProvided_Error(t *testing.T) {
	enc, _ := newDefaultEncryptor(t)
	enc.SetStore(nil)

	_, err := enc.Encrypt(docID, fieldName, getPlainText())

	assert.ErrorIs(t, err, ErrNoStorageProvided)
}

func TestEncryptorDecrypt_IfNoStorageProvided_Error(t *testing.T) {
	enc, _ := newDefaultEncryptor(t)
	enc.SetStore(nil)

	_, err := enc.Decrypt(docID, fieldName, getPlainText())

	assert.ErrorIs(t, err, ErrNoStorageProvided)
}

func TestEncryptorDecrypt_IfStorageReturnsError_Error(t *testing.T) {
	enc, st := newDefaultEncryptor(t)

	st.EXPECT().Get(mock.Anything, mock.Anything).Return(nil, testErr)

	_, err := enc.Decrypt(docID, fieldName, []byte("test"))

	assert.ErrorIs(t, err, testErr)
}

func TestEncryptorDecrypt_IfKeyFoundInStorage_ShouldUseItToReturnPlainText(t *testing.T) {
	enc, st := newDefaultEncryptor(t)

	st.EXPECT().Get(mock.Anything, mock.Anything).Return(getEncKey(noFieldName), nil)

	plainText, err := enc.Decrypt(docID, fieldName, getCipherText(t, noFieldName))

	assert.NoError(t, err)
	assert.Equal(t, getPlainText(), plainText)
}

func TestEncryptDoc_IfContextHasNoEncryptor_ReturnNil(t *testing.T) {
	data, err := EncryptDoc(context.Background(), docID, fieldName, getPlainText())
	assert.Nil(t, data, "data should be nil")
	assert.NoError(t, err, "error should be nil")
}

func TestDecryptDoc_IfContextHasNoEncryptor_ReturnNil(t *testing.T) {
	data, err := DecryptDoc(context.Background(), docID, fieldName, getCipherText(t, fieldName))
	assert.Nil(t, data, "data should be nil")
	assert.NoError(t, err, "error should be nil")
}
