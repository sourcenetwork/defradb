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

	"github.com/sourcenetwork/defradb/datastore/mocks"
	"github.com/sourcenetwork/defradb/internal/core"
)

var testErr = errors.New("test error")

var docID = "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3"

func getPlainText() []byte {
	return []byte("test")
}

func getCipherText(t *testing.T) []byte {
	cipherText, err := EncryptAES(getPlainText(), []byte(testEncryptionKey))
	assert.NoError(t, err)
	return cipherText
}

func TestEncryptorEncrypt_IfStorageReturnsError_Error(t *testing.T) {
	enc := newDocEncryptor(context.Background())
	st := mocks.NewDSReaderWriter(t)
	enc.SetStore(st)

	st.EXPECT().Get(mock.Anything, mock.Anything).Return(nil, testErr)

	_, err := enc.Encrypt(docID, 0, []byte("test"))

	assert.ErrorIs(t, err, testErr)
}

func TestEncryptorEncrypt_IfNoKeyFoundInStorage_ShouldGenerateKeyStoreItAndReturnCipherText(t *testing.T) {
	enc := newDocEncryptor(context.Background())
	st := mocks.NewDSReaderWriter(t)
	enc.EnableKeyGeneration()
	enc.SetStore(st)

	st.EXPECT().Get(mock.Anything, mock.Anything).Return(nil, ds.ErrNotFound)

	storeKey := core.NewEncStoreDocKey(docID, 0)

	st.EXPECT().Put(mock.Anything, storeKey.ToDS(), []byte(testEncryptionKey)).Return(nil)

	cipherText, err := enc.Encrypt(docID, 0, getPlainText())

	assert.NoError(t, err)
	assert.Equal(t, getCipherText(t), cipherText)
}

func TestEncryptorEncrypt_IfKeyFoundInStorage_ShouldUseItToReturnCipherText(t *testing.T) {
	enc := newDocEncryptor(context.Background())
	st := mocks.NewDSReaderWriter(t)
	enc.EnableKeyGeneration()
	enc.SetStore(st)

	st.EXPECT().Get(mock.Anything, mock.Anything).Return([]byte(testEncryptionKey), nil)

	cipherText, err := enc.Encrypt(docID, 0, getPlainText())

	assert.NoError(t, err)
	assert.Equal(t, getCipherText(t), cipherText)
}

func TestEncryptorEncrypt_IfStorageFailsToStoreEncryptionKey_ReturnError(t *testing.T) {
	enc := newDocEncryptor(context.Background())
	st := mocks.NewDSReaderWriter(t)
	enc.EnableKeyGeneration()
	enc.SetStore(st)

	st.EXPECT().Get(mock.Anything, mock.Anything).Return(nil, ds.ErrNotFound)

	st.EXPECT().Put(mock.Anything, mock.Anything, mock.Anything).Return(testErr)

	_, err := enc.Encrypt(docID, 0, getPlainText())

	assert.ErrorIs(t, err, testErr)
}

func TestEncryptorEncrypt_IfKeyGenerationIsNotEnabled_ShouldReturnPlainText(t *testing.T) {
	enc := newDocEncryptor(context.Background())
	st := mocks.NewDSReaderWriter(t)
	// we don call enc.EnableKeyGeneration()
	enc.SetStore(st)

	st.EXPECT().Get(mock.Anything, mock.Anything).Return(nil, ds.ErrNotFound)

	cipherText, err := enc.Encrypt(docID, 0, getPlainText())

	assert.NoError(t, err)
	assert.Equal(t, getPlainText(), cipherText)
}

func TestEncryptorEncrypt_IfNoStorageProvided_Error(t *testing.T) {
	enc := newDocEncryptor(context.Background())
	enc.EnableKeyGeneration()
	// we don call enc.SetStore(st)

	_, err := enc.Encrypt(docID, 0, getPlainText())

	assert.ErrorIs(t, err, ErrNoStorageProvided)
}

func TestEncryptorDecrypt_IfNoStorageProvided_Error(t *testing.T) {
	enc := newDocEncryptor(context.Background())
	enc.EnableKeyGeneration()
	// we don call enc.SetStore(st)

	_, err := enc.Decrypt(docID, 0, getPlainText())

	assert.ErrorIs(t, err, ErrNoStorageProvided)
}

func TestEncryptorDecrypt_IfStorageReturnsError_Error(t *testing.T) {
	enc := newDocEncryptor(context.Background())
	st := mocks.NewDSReaderWriter(t)
	enc.SetStore(st)

	st.EXPECT().Get(mock.Anything, mock.Anything).Return(nil, testErr)

	_, err := enc.Decrypt(docID, 0, []byte("test"))

	assert.ErrorIs(t, err, testErr)
}

func TestEncryptorDecrypt_IfKeyFoundInStorage_ShouldUseItToReturnPlainText(t *testing.T) {
	enc := newDocEncryptor(context.Background())
	st := mocks.NewDSReaderWriter(t)
	enc.EnableKeyGeneration()
	enc.SetStore(st)

	st.EXPECT().Get(mock.Anything, mock.Anything).Return([]byte(testEncryptionKey), nil)

	plainText, err := enc.Decrypt(docID, 0, getCipherText(t))

	assert.NoError(t, err)
	assert.Equal(t, getPlainText(), plainText)
}

func TestEncryptorDecrypt_IfNoKeyFoundInStorage_ShouldGenerateKeyStoreItAndReturnCipherText(t *testing.T) {
	enc := newDocEncryptor(context.Background())
	st := mocks.NewDSReaderWriter(t)
	enc.EnableKeyGeneration()
	enc.SetStore(st)

	st.EXPECT().Get(mock.Anything, mock.Anything).Return(nil, ds.ErrNotFound)

	storeKey := core.NewEncStoreDocKey(docID, 0)

	st.EXPECT().Put(mock.Anything, storeKey.ToDS(), []byte(testEncryptionKey)).Return(nil)

	cipherText, err := enc.Encrypt(docID, 0, getPlainText())

	assert.NoError(t, err)
	assert.Equal(t, getCipherText(t), cipherText)
}
