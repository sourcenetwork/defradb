// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package se

import (
	"context"
	"fmt"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	secore "github.com/sourcenetwork/defradb/internal/se/core"
)

// BlockData represents the data needed for SE processing
type BlockData struct {
	IsComposite  bool
	IsCollection bool
	FieldName    string
	Data         []byte
}

// ProcessBlock handles SE for a block
func ProcessBlock(
	ctx context.Context,
	txn datastore.Txn,
	block *coreblock.Block,
) error {
	seCtx, ok := ctx.Value(contextKey{}).(*Context)
	if !ok {
		return nil
	}

	if block.Delta.IsComposite() || block.Delta.IsCollection() {
		return nil
	}

	fieldName := block.Delta.GetFieldName()

	var encIdx *client.EncryptedIndexDescription
	for _, idx := range seCtx.config.EncryptedFields {
		if idx.FieldName == fieldName {
			encIdx = &idx
			break
		}
	}

	if encIdx == nil {
		return nil
	}

	var tag []byte
	var err error

	switch encIdx.Type {
	case client.EncryptedIndexTypeEquality:
		tag, err = secore.GenerateEqualityTag(
			seCtx.config.Key,
			seCtx.config.CollectionID,
			fieldName,
			block.Delta.GetData(),
		)
	default:
		return fmt.Errorf("unsupported index type: %s", encIdx.Type)
	}

	if err != nil {
		return err
	}

	artifact := secore.Artifact{
		Type:         secore.ArtifactTypeEqualityTag,
		CollectionID: seCtx.config.CollectionID,
		FieldName:    fieldName,
		Tag:          tag,
		Operation:    secore.OperationAdd,
		DocID:        seCtx.doc.ID().String(),
	}

	seCtx.artifacts = append(seCtx.artifacts, artifact)
	return nil
}
