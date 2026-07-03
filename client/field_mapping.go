// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

// CollectionFieldMapping contains the short ID mapping for a single collection,
// enabling cross-instance KV import by mapping source short IDs to FieldID CIDs.
type CollectionFieldMapping struct {
	// CollectionID is the content-addressed CID of the collection.
	CollectionID string `json:"collection_id"`
	// CollectionShortID is the source instance's short collection ID.
	CollectionShortID uint32 `json:"collection_short_id"`
	// FieldIDMapping maps source short field ID → FieldID (CID string).
	FieldIDMapping map[uint32]string `json:"field_id_mapping"`
}
