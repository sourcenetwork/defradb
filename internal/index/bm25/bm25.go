// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package bm25 contains DefraDB's storage-independent BM25 analyzer and scoring formula.
package bm25

import (
	"math"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/sourcenetwork/defradb/client"
)

// TokenFrequencies lowercases val, splits on every non-letter/non-digit rune, drops
// single-rune tokens, and returns each retained token's frequency. There is deliberately no
// stemming or stop-word list; this exact analyzer is part of the persisted index format.
func TokenFrequencies(val string) map[string]int {
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

// ScoreTerm returns one query term's BM25 contribution for a document.
func ScoreTerm(
	termFrequency, documentFrequency, documentCount, documentLength uint64,
	averageDocumentLength float64,
	params client.BM25Params,
) float64 {
	if termFrequency == 0 || documentFrequency == 0 || documentCount == 0 || averageDocumentLength == 0 {
		return 0
	}
	tf := float64(termFrequency)
	df := float64(documentFrequency)
	n := float64(documentCount)
	dl := float64(documentLength)
	idf := math.Log(1 + (n-df+0.5)/(df+0.5))
	numerator := tf * (params.K1 + 1)
	denominator := tf + params.K1*(1-params.B+params.B*dl/averageDocumentLength)
	if denominator == 0 {
		return 0
	}
	return idf * numerator / denominator
}
