// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package core

import "strings"

// Trigrams returns the distinct trigrams of val: the value lowercased, then every 3-byte
// window of it, stride 1 and overlapping, with no padding.
//
// The windows are bytes, not runes. A multi-byte rune is therefore split across windows,
// which is safe because a trigram index only ever produces candidates and the real pattern
// match is re-run on each one.
//
// A value shorter than 3 bytes produces no trigrams at all, so the document holding it has
// no entry in the index and cannot be found through it. That is intended: padding short
// values to reach three bytes would put every one of them under the same handful of keys.
//
// It lives here rather than beside the index because the write path (internal/db) and the
// read path (internal/db/fetcher) must derive the same trigrams from the same value, and
// internal/db imports internal/db/fetcher.
func Trigrams(val string) []string {
	val = strings.ToLower(val)
	if len(val) < 3 {
		return nil
	}
	// One entry per distinct trigram per document. A repeated trigram would otherwise write
	// the same key twice, and the second delete of it would report the index as corrupt.
	seen := make(map[string]struct{}, len(val)-2)
	result := make([]string, 0, len(val)-2)
	for i := 0; i+3 <= len(val); i++ {
		trigram := val[i : i+3]
		if _, ok := seen[trigram]; ok {
			continue
		}
		seen[trigram] = struct{}{}
		result = append(result, trigram)
	}
	return result
}
