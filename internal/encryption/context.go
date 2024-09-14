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

	"github.com/sourcenetwork/immutable"
)

// docEncContextKey is the key type for document encryption context values.
type docEncContextKey struct{}

// configContextKey is the key type for encryption context values.
type configContextKey struct{}

// GetEncryptorFromContext returns a document encryptor from the given context.
// It returns nil if no encryptor exists in the context.
func GetEncryptorFromContext(ctx context.Context) *DocEncryptor {
	enc, ok := ctx.Value(docEncContextKey{}).(*DocEncryptor)
	if ok {
		setConfig(ctx, enc)
	}
	return enc
}

func setConfig(ctx context.Context, enc *DocEncryptor) {
	enc.SetConfig(GetContextConfig(ctx))
	enc.ctx = ctx
}

// EnsureContextWithEncryptor returns a context with a document encryptor and the
// document encryptor itself. If the context already has an encryptor, it
// returns the context and encryptor as is. Otherwise, it creates a new
// document encryptor and stores it in the context.
func EnsureContextWithEncryptor(ctx context.Context) (context.Context, *DocEncryptor) {
	enc := GetEncryptorFromContext(ctx)
	if enc == nil {
		enc = newDocEncryptor(ctx)
		ctx = context.WithValue(ctx, docEncContextKey{}, enc)
	}
	return ctx, enc
}

// GetContextConfig returns the doc encryption config from the given context.
func GetContextConfig(ctx context.Context) immutable.Option[DocEncConfig] {
	encConfig, ok := ctx.Value(configContextKey{}).(DocEncConfig)
	if ok {
		return immutable.Some(encConfig)
	}
	return immutable.None[DocEncConfig]()
}

// SetContextConfig returns a new context with the doc encryption config set.
func SetContextConfig(ctx context.Context, encConfig DocEncConfig) context.Context {
	ctx, _ = EnsureContextWithEncryptor(ctx)
	return context.WithValue(ctx, configContextKey{}, encConfig)
}

// SetContextConfigFromParams returns a new context with the doc encryption config created from given params.
// If encryptDoc is false, and encryptFields is empty, the context is returned as is.
func SetContextConfigFromParams(ctx context.Context, encryptDoc bool, encryptFields []string) context.Context {
	if encryptDoc || len(encryptFields) > 0 {
		conf := DocEncConfig{EncryptedFields: encryptFields}
		if encryptDoc {
			conf.IsDocEncrypted = true
		}
		ctx = SetContextConfig(ctx, conf)
	}
	return ctx
}
