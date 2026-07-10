// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"context"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
)

// filterAllowsReplication returns true when no filter is configured or when
// the configured filter returns true for this document.
func (p *P2P) filterAllowsReplication(
	ctx context.Context,
	collectionID string,
	docID string,
	block *coreblock.Block,
) bool {
	if p.replicationFilter == nil {
		return true
	}
	fields := extractFieldsFromBlock(block)
	return p.replicationFilter.AllowReplication(ctx, collectionID, docID, fields)
}

// extractFieldsFromBlock extracts field name → decoded value from a block's CRDT delta.
func extractFieldsFromBlock(block *coreblock.Block) map[string]any {
	if block == nil {
		return nil
	}

	fields := make(map[string]any)

	name := block.Delta.GetFieldName()
	if name != "" {
		fields[name] = block.Delta.GetData()
	}

	// Also extract from linked field blocks.
	for _, link := range block.Links {
		if link.Name != "" {
			// Use the link name as a best-effort field name indicator.
			fields[link.Name] = nil
		}
	}

	return fields
}
