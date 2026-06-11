// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package coreblock

import (
	"context"

	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/defradb/internal/encryption"
)

// LoadEncryptionBlock loads and decodes the encryption block referenced by encLink
// from the given link system. Callers are expected to wrap the returned error with
// any context-specific error type they need.
func LoadEncryptionBlock(
	ctx context.Context,
	encBlockLS linking.LinkSystem,
	encLink cidlink.Link,
) (*Encryption, error) {
	nd, err := encBlockLS.Load(linking.LinkContext{Ctx: ctx}, encLink, EncryptionSchemaPrototype)
	if err != nil {
		return nil, err
	}

	return GetEncryptionBlockFromNode(nd)
}

// DecryptBlock decrypts the delta of the given block using the supplied encryption
// block and returns a new block carrying the plaintext delta. Composite and
// collection blocks have no payload to decrypt and are returned unchanged.
//
// If the encryptor produces no plaintext (e.g. the key is not available), the
// returned block is nil and the returned error is nil — callers must treat this
// as "block not readable" rather than success.
func DecryptBlock(
	ctx context.Context,
	block *Block,
	encBlock *Encryption,
) (*Block, error) {
	_, encryptor := encryption.EnsureContextWithEncryptor(ctx)

	if block.Delta.IsComposite() || block.Delta.IsCollection() {
		// composite and collection blocks carry no encrypted payload
		return block, nil
	}

	bytes, err := encryptor.Decrypt(block.Delta.GetData(), encBlock.Key)
	if err != nil {
		return nil, err
	}
	if len(bytes) == 0 {
		return nil, nil
	}
	newBlock := block.Clone()
	newBlock.Delta.SetData(bytes)
	return newBlock, nil
}

// ProcessEncryptedBlock loads the encryption block and decrypts the delta if the
// given block is encrypted. Used by both the live merge path and the version
// replay path so they stay in lockstep.
//
// Returns:
//   - (block, true, nil)  — block is not encrypted, or decryption succeeded.
//   - (block, false, nil) — block is encrypted but the encryption block or key
//     is not available locally; the caller decides how to react.
//   - (nil, false, err)   — load or decrypt failed with a hard error.
func ProcessEncryptedBlock(
	ctx context.Context,
	encBlockLS linking.LinkSystem,
	block *Block,
) (*Block, bool, error) {
	if !block.IsEncrypted() {
		return block, true, nil
	}

	encBlock, err := LoadEncryptionBlock(ctx, encBlockLS, *block.Encryption)
	if err != nil {
		return nil, false, err
	}
	if encBlock == nil {
		return block, false, nil
	}

	plainTextBlock, err := DecryptBlock(ctx, block, encBlock)
	if err != nil {
		return nil, false, err
	}
	if plainTextBlock == nil {
		return block, false, nil
	}
	return plainTextBlock, true, nil
}
