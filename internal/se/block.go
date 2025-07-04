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
	block *coreblock.Block,
) error {
	seCtx, ok := ctx.Value(contextKey{}).(*Context)
	if !ok {
		return nil
	}

	artifact, err := GenerateArtifactFromBlock(
		block,
		seCtx.config.CollectionID,
		seCtx.doc.ID().String(),
		seCtx.config.EncryptedFields,
		seCtx.config.Key,
	)
	if err != nil {
		return err
	}

	if artifact != nil {
		seCtx.artifacts = append(seCtx.artifacts, *artifact)
	}

	return nil
}

// GenerateArtifactFromBlock generates an SE artifact from a block if it contains an encrypted field.
// This is a helper function used by ProcessBlock and can be used by other components.
//
// Parameters:
//   - block: The block to process
//   - collectionID: Collection ID the block belongs to
//   - docID: Document ID the block belongs to
//   - encryptedFields: List of encrypted field configurations
//   - seKey: SE key for tag generation
//
// Returns nil, nil if the block doesn't contain an encrypted field
func GenerateArtifactFromBlock(
	block *coreblock.Block,
	collectionID string,
	docID string,
	encryptedFields []client.EncryptedIndexDescription,
	seKey []byte,
) (*secore.Artifact, error) {
	if block.Delta.IsComposite() || block.Delta.IsCollection() {
		return nil, nil
	}

	fieldName := block.Delta.GetFieldName()

	var encIdx *client.EncryptedIndexDescription
	for _, idx := range encryptedFields {
		if idx.FieldName == fieldName {
			encIdx = &idx
			break
		}
	}

	if encIdx == nil {
		return nil, nil
	}

	var tag []byte
	var err error
	switch encIdx.Type {
	case client.EncryptedIndexTypeEquality:
		tag, err = secore.GenerateEqualityTag(
			seKey,
			collectionID,
			fieldName,
			block.Delta.GetData(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to generate equality tag for field %s: %w", fieldName, err)
		}
	default:
		return nil, fmt.Errorf("unsupported index type: %s", encIdx.Type)
	}

	artifact := &secore.Artifact{
		Type:         secore.ArtifactTypeEqualityTag,
		CollectionID: collectionID,
		FieldName:    fieldName,
		SearchTag:    tag,
		Operation:    secore.OperationAdd,
		DocID:        docID,
		IndexID:      fmt.Sprintf("%s_%s", collectionID, fieldName), // Generate index ID from collection and field
	}

	return artifact, nil
}
