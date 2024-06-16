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

import "github.com/sourcenetwork/defradb/datastore"

type DocEncryptor struct {
	encryptionKey string
	store         datastore.DSReaderWriter
}

func newDocEncryptor() *DocEncryptor {
	return &DocEncryptor{}
}

func (d *DocEncryptor) SetKey(encryptionKey string) {
	d.encryptionKey = encryptionKey
}

func (d *DocEncryptor) SetStore(store datastore.DSReaderWriter) {
	d.store = store
}

func (d *DocEncryptor) Encrypt(docID string, fieldID int, plainText []byte) ([]byte, error) {
	return EncryptAES(plainText, []byte(d.encryptionKey))
}

func (d *DocEncryptor) Decrypt(docID string, fieldID int, cipherText []byte) ([]byte, error) {
	return DecryptAES(cipherText, []byte(d.encryptionKey))
}
