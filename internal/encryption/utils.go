// Copyright 2025 Democratized Data Foundation
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

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
)

// DecryptBlock returns the result of decrypting the given block.
func DecryptBlock(
	ctx context.Context,
	block *coreblock.Block,
	encBlock *coreblock.Encryption,
) (*coreblock.Block, error) {
	_, encryptor := EnsureContextWithEncryptor(ctx)

	if block.Delta.IsComposite() || block.Delta.IsCollection() {
		// for composite blocks there is nothing to decrypt
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
