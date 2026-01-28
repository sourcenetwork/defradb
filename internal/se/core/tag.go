// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package secore

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

// GenerateEqualityTag creates a deterministic search tag for equality queries
//
// SECURITY NOTE: This function generates the SAME tag for the same value across
// ALL documents in the collection for the same identity. This enables efficient
// equality search but reveals when multiple documents share the same field value
// (frequency analysis). For fields with low cardinality (e.g., boolean, status codes),
// consider the privacy implications.
//
// The tag is computed as: HMAC-SHA256(key, "eq:identity:collectionID:fieldName" || value)[:16]
func GenerateEqualityTag(
	key []byte,
	identityID string,
	collectionID string,
	fieldName string,
	value []byte,
) []byte {
	// Domain separation explanation:
	// - "eq" indicates equality search (vs future range/prefix)
	// - identityID isolates tags per identity on shared remote nodes
	// - collectionID ensures tags are unique per collection
	// - fieldName ensures tags are unique per field
	// This prevents cross-identity, cross-field and cross-collection tag collisions
	domainSeparator := fmt.Sprintf("eq:%s:%s:%s", identityID, collectionID, fieldName)

	h := hmac.New(sha256.New, key)
	h.Write([]byte(domainSeparator))
	h.Write(value)
	tag := h.Sum(nil)

	// Truncate to 16 bytes for storage and network efficiency.
	// HMAC's security doesn't degrade linearly with truncation and (128 bits) is explicitly approved
	// by cryptographic standards providing good collision resistance even with billions of documents.
	return tag[:16]
}
