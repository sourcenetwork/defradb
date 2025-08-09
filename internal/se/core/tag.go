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
func GenerateEqualityTag(
	key []byte,
	collectionID string,
	fieldName string,
	value []byte,
) ([]byte, error) {
	// Domain separation explanation:
	// - "eq" indicates equality search (vs future range/prefix)
	// - collectionID ensures tags are unique per collection
	// - fieldName ensures tags are unique per field
	// This prevents cross-field and cross-collection tag collisions
	// Note: This is HMAC input, not stored data
	domainSeparator := fmt.Sprintf("eq:%s:%s", collectionID, fieldName)

	// Compute HMAC-SHA256
	h := hmac.New(sha256.New, key)
	h.Write([]byte(domainSeparator))
	h.Write(value)
	tag := h.Sum(nil)

	// Truncate to 16 bytes for efficiency (128-bit security)
	return tag[:16], nil
}