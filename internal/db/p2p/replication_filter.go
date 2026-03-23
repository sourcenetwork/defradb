// Copyright 2026 Democratized Data Foundation
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

	"github.com/fxamacker/cbor/v2"

	"github.com/sourcenetwork/corelog"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
)

// extractFieldsFromCARBlocks extracts field name-value pairs from decoded CAR blocks.
// It iterates through all blocks looking for LWW and Counter deltas that contain
// field-level data, decodes their CBOR values, and returns a map of field names to values.
func extractFieldsFromCARBlocks(regularBlocks []coreblock.Block) map[string]any {
	fields := make(map[string]any)

	for i := range regularBlocks {
		block := &regularBlocks[i]
		delta := block.Delta

		var fieldName string
		var data []byte

		if delta.LWWDelta != nil {
			fieldName = delta.LWWDelta.FieldName
			data = delta.LWWDelta.Data
		} else if delta.CounterDelta != nil {
			fieldName = delta.CounterDelta.FieldName
			data = delta.CounterDelta.Data
		}

		if fieldName == "" || len(data) == 0 {
			continue
		}

		var value any
		if err := cbor.Unmarshal(data, &value); err != nil {
			// If CBOR decoding fails, store raw bytes
			value = data
		}

		fields[fieldName] = value
	}

	return fields
}

// parseCARBlocksForFilter parses the regular blocks from a CAR file into coreblock.Block
// structs for field extraction. Only blocks that successfully decode are returned.
func parseCARBlocksForFilter(parsed *parsedCAR) []coreblock.Block {
	var decoded []coreblock.Block

	for _, blk := range parsed.regularBlocks {
		block, err := coreblock.GetFromBytes(blk.RawData())
		if err != nil {
			continue
		}
		decoded = append(decoded, *block)
	}

	return decoded
}

// filterCARDocument checks whether a single CAR document should be replicated.
// It extracts field values from the parsed CAR blocks and calls the filter.
func (p *P2P) filterCARDocument(
	ctx context.Context,
	collectionID string,
	docID string,
	parsed *parsedCAR,
) bool {
	if p.replicationFilter == nil {
		return true
	}

	decoded := parseCARBlocksForFilter(parsed)
	fields := extractFieldsFromCARBlocks(decoded)

	return p.filterAllowsReplication(ctx, collectionID, docID, fields)
}

// filterAllowsReplication checks if the replication filter allows the given document.
// Returns true if no filter is set or if the filter allows the document.
func (p *P2P) filterAllowsReplication(
	ctx context.Context,
	collectionID string,
	docID string,
	fields map[string]any,
) bool {
	if p.replicationFilter == nil {
		return true
	}

	// Resolve human-readable collection name from cached collection lookup.
	// The hasCollection call that precedes filtering populates this cache.
	filterID := collectionID
	if col := p.getCachedCollection(collectionID); col != nil {
		filterID = col.Name()
	}

	allowed := p.replicationFilter.AllowReplication(ctx, filterID, docID, fields)
	if !allowed {
		log.Info("Document rejected by replication filter",
			corelog.String("Collection", filterID),
			corelog.String("DocID", docID))
	}
	return allowed
}

// filterNonCARDocument checks whether a non-CAR document should be replicated.
// After DAG sync, all blocks (including field-level) are in the blockstore.
// For composite blocks, it follows the Links to fetch child field blocks and
// extract their values. For field-level blocks, it extracts the value directly.
func (p *P2P) filterNonCARDocument(
	ctx context.Context,
	collectionID string,
	docID string,
	block *coreblock.Block,
) bool {
	if p.replicationFilter == nil {
		return true
	}

	fields := make(map[string]any)

	if block.Delta.IsComposite() {
		// Composite block: follow Links to get field-level blocks from the blockstore
		bstore := p.db.Multistore().Blockstore()
		for _, link := range block.Links {
			if link.Name == "" || link.Name == "_head" {
				continue
			}
			childRaw, err := bstore.Get(ctx, link.Link.Cid)
			if err != nil {
				continue
			}
			childBlock, err := coreblock.GetFromBytes(childRaw.RawData())
			if err != nil {
				continue
			}
			if childBlock.Delta.LWWDelta != nil && len(childBlock.Delta.LWWDelta.Data) > 0 {
				var value any
				if err := cbor.Unmarshal(childBlock.Delta.LWWDelta.Data, &value); err == nil {
					fields[link.Name] = value
				}
			} else if childBlock.Delta.CounterDelta != nil && len(childBlock.Delta.CounterDelta.Data) > 0 {
				var value any
				if err := cbor.Unmarshal(childBlock.Delta.CounterDelta.Data, &value); err == nil {
					fields[link.Name] = value
				}
			}
		}
	} else if block.Delta.LWWDelta != nil && block.Delta.LWWDelta.FieldName != "" {
		// Field-level block: extract value directly
		var value any
		if err := cbor.Unmarshal(block.Delta.LWWDelta.Data, &value); err == nil {
			fields[block.Delta.LWWDelta.FieldName] = value
		}
	}

	return p.filterAllowsReplication(ctx, collectionID, docID, fields)
}
