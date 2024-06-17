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

	"github.com/sourcenetwork/defradb/datastore"
)

// docEncContextKey is the key type for document encryption context values.
type docEncContextKey struct{}

// TryGetContextDocEnc returns a document encryption and a bool indicating if
// it was retrieved from the given context.
func TryGetContextEncryptor(ctx context.Context) (*DocEncryptor, bool) {
	enc, ok := ctx.Value(docEncContextKey{}).(*DocEncryptor)
	return enc, ok
}

func getContextWithDocEnc(ctx context.Context) (context.Context, *DocEncryptor) {
	enc, ok := TryGetContextEncryptor(ctx)
	if !ok {
		enc = newDocEncryptor(ctx)
		ctx = context.WithValue(ctx, docEncContextKey{}, enc)
	}
	return ctx, enc
}

func Context(ctx context.Context) context.Context {
	ctx, _ = getContextWithDocEnc(ctx)
	return ctx
}

func ContextWithKey(ctx context.Context, encryptionKey []byte) context.Context {
	ctx, encryptor := getContextWithDocEnc(ctx)
	encryptor.SetKey(encryptionKey)
	return ctx
}

func ContextWithStore(ctx context.Context, txn datastore.Txn) context.Context {
	ctx, encryptor := getContextWithDocEnc(ctx)
	encryptor.SetStore(txn.Encstore())
	return ctx
}

func EncryptDoc(ctx context.Context, docID string, fieldID uint32, plainText []byte) ([]byte, error) {
	enc, ok := TryGetContextEncryptor(ctx)
	if !ok {
		return plainText, nil
	}
	return enc.Encrypt(docID, fieldID, plainText)
}

func DecryptDoc(ctx context.Context, docID string, fieldID uint32, cipherText []byte) ([]byte, error) {
	enc, ok := TryGetContextEncryptor(ctx)
	if !ok {
		return cipherText, nil
	}
	return enc.Decrypt(docID, fieldID, cipherText)
}
