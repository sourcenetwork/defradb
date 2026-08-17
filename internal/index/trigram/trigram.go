// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package trigram contains the storage-independent extraction and conservative query compiler used
// by DefraDB's trigram index.
package trigram

import "strings"

// Extract returns the distinct trigrams of val: the value lowercased, then every overlapping
// three-byte window with no padding. Bytes rather than runes are intentional: the index only
// produces candidates, and the original predicate is always reapplied to each candidate.
func Extract(val string) []string {
	val = strings.ToLower(val)
	if len(val) < 3 {
		return nil
	}
	seen := make(map[string]struct{}, len(val)-2)
	result := make([]string, 0, len(val)-2)
	for i := 0; i+3 <= len(val); i++ {
		value := val[i : i+3]
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
