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

	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
)

type DocEncryptor struct {
	encryptionKey []byte
	ctx           context.Context
	store         datastore.DSReaderWriter
}

func newDocEncryptor(ctx context.Context) *DocEncryptor {
	return &DocEncryptor{ctx: ctx}
}

func (d *DocEncryptor) SetKey(encryptionKey []byte) {
	d.encryptionKey = encryptionKey
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
		if len(d.encryptionKey) == 0 {
			return plainText, nil
		}
		if d.store != nil {
			err = d.store.Put(d.ctx, storeKey.ToDS(), d.encryptionKey)
			if err != nil {
				return nil, err
			}
		}
		encryptionKey = d.encryptionKey
	}
	return EncryptAES(plainText, encryptionKey)
}

func (d *DocEncryptor) Decrypt(docID string, fieldID uint32, cipherText []byte) ([]byte, error) {
	encKey, _, err := d.fetchEncryptionKey(docID, fieldID)
	if err != nil {
		return nil, err
	}
	if len(encKey) == 0 {
		return cipherText, nil
	}
	return DecryptAES(cipherText, encKey)
}

func (d *DocEncryptor) fetchEncryptionKey(docID string, fieldID uint32) ([]byte, core.EncStoreDocKey, error) {
	storeKey := core.NewEncStoreDocKey(docID, fieldID)
	if d.store == nil {
		return nil, core.EncStoreDocKey{}, nil
	}
	encryptionKey, err := d.store.Get(d.ctx, storeKey.ToDS())
	isNotFound := errors.Is(err, ds.ErrNotFound)
	if err != nil && !isNotFound {
		return nil, core.EncStoreDocKey{}, err
	}
	return encryptionKey, storeKey, nil
}
