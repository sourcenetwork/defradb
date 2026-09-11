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
	"bytes"
	"context"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	car "github.com/ipld/go-car/v2"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
)

// filterAllowsReplication returns true when no filter is configured or when
// the configured filter returns true for this document.
func (p *P2P) filterAllowsReplication(
	ctx context.Context,
	collectionID string,
	docID string,
	block *coreblock.Block,
	carData []byte,
) bool {
	if p.replicationFilter == nil {
		return true
	}
	fields := extractFieldsFromBlock(block, carData)
	return p.replicationFilter.AllowReplication(ctx, collectionID, docID, fields)
}

// extractFieldsFromBlock returns the document's field values keyed by field name.
//
// A composite block carries no value of its own, only links to the field blocks that do, so the
// values come from the accompanying CAR. Without one the linked blocks have not arrived and only
// the block's own value is returned. A field is absent unless its value was decoded, so a caller
// can tell a value it does not have from one that is null.
func extractFieldsFromBlock(block *coreblock.Block, carData []byte) map[string]any {
	if block == nil {
		return nil
	}

	fields := make(map[string]any)
	addFieldValue(fields, block)

	if len(block.Links) == 0 || len(carData) == 0 {
		return fields
	}
	addLinkedFieldValues(fields, block.Links, carData)

	return fields
}

// addFieldValue decodes the value a field block holds. The name comes from the block's own delta,
// so a linked block that names no field contributes nothing.
func addFieldValue(fields map[string]any, block *coreblock.Block) {
	name := block.Delta.GetFieldName()
	if name == "" {
		return
	}
	// An encrypted delta holds ciphertext, which would decode to a value the document does
	// not have.
	if block.Encryption != nil {
		return
	}
	var value any
	if err := cbor.Unmarshal(block.Delta.GetData(), &value); err != nil {
		return
	}
	fields[name] = value
}

// addLinkedFieldValues reads the linked blocks out of the CAR, which holds one commit: the root
// block and the fields it links. A CAR that cannot be read is left to the import path to report,
// so the filter just decides on the fields it did read.
func addLinkedFieldValues(fields map[string]any, links []coreblock.DAGLink, carData []byte) {
	wanted := make(map[cid.Cid]struct{}, len(links))
	for _, link := range links {
		wanted[link.Cid] = struct{}{}
	}

	reader, err := car.NewBlockReader(bytes.NewReader(carData))
	if err != nil {
		return
	}

	for len(wanted) > 0 {
		carBlock, err := reader.Next()
		if err != nil {
			return
		}
		blockCID := carBlock.Cid()
		if _, ok := wanted[blockCID]; !ok {
			continue
		}
		delete(wanted, blockCID)

		decoded, err := coreblock.GetFromBytes(carBlock.RawData())
		if err != nil {
			continue
		}
		addFieldValue(fields, decoded)
	}
}
