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
	"github.com/sourcenetwork/immutable"
)

// docEncContextKey is the key type for document encryption context values.
type docEncContextKey struct{}

// configContextKey is the key type for encryption context values.
type configContextKey struct{}

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
	ctx, encryptor := getContextWithDocEnc(ctx)
	encryptor.EnableKeyGeneration()
	return ctx
}

func ContextWithStore(ctx context.Context, txn datastore.Txn) context.Context {
	ctx, encryptor := getContextWithDocEnc(ctx)
	encryptor.SetStore(txn.Encstore())
	return ctx
}

// GetContextConfig returns the doc encryption config from the given context.
func GetContextConfig(ctx context.Context) immutable.Option[DocEncConfig] {
	encConfig, ok := ctx.Value(configContextKey{}).(DocEncConfig)
	if ok {
		return immutable.Some(encConfig)
	}
	return immutable.None[DocEncConfig]()
}

// SetContextConfig returns a new context with the encryption value set.
func SetContextConfig(ctx context.Context, encConfig DocEncConfig) context.Context {
	return context.WithValue(ctx, configContextKey{}, encConfig)
}