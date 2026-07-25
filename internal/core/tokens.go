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

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Tokens returns the terms of val and how many times each occurs: val lowercased, split on
// every character that is not a letter or a digit, with single-character tokens dropped.
//
// There is no stemming and no stop-word list. A stemmer is a dependency that silently
// invalidates every index built with the previous version of it, and BM25's inverse document
// frequency already de-weights a term that appears in most documents, which is what a stop-word
// list is for.
//
// It lives here rather than beside the index because the write path (internal/db) and the read
// path (internal/db/fetcher) must derive the same terms from the same value, and internal/db
// imports internal/db/fetcher. A query tokenized differently from the documents matches nothing
// and reports no error, so the two must be the same function rather than the same idea written
// twice.
func Tokens(val string) map[string]int {
	tokens := map[string]int{}
	for _, word := range strings.FieldsFunc(strings.ToLower(val), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		if utf8.RuneCountInString(word) > 1 {
			tokens[word]++
		}
	}
	return tokens
}
