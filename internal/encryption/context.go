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

import "context"

// docEncContextKey is the key type for document encryption context values.
type docEncContextKey struct{}

// TryGetContextDocEnc returns a document encryption and a bool indicating if
// it was retrieved from the given context.
func TryGetContextDocEnc(ctx context.Context) (*DocCipher, bool) {
	d, ok := ctx.Value(docEncContextKey{}).(*DocCipher)
	return d, ok
}

func SetDocEncContext(ctx context.Context, encryptionKey string) context.Context {
	cipher, ok := TryGetContextDocEnc(ctx)
	if !ok {
		cipher = NewDocCipher()
		ctx = context.WithValue(ctx, docEncContextKey{}, cipher)
	}
	cipher.setKey(encryptionKey)
	return ctx
}
