// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package fetcher

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The expected queries are written in the reference notation: juxtaposition is AND, '|' is
// OR, '+' matches everything (scan) and '-' matches nothing.
func TestRegexTrigramQuery(t *testing.T) {
	tests := []struct {
		pattern  string
		expected string
	}{
		// A literal is the AND of its overlapping trigrams, lowercased.
		{`abcdef`, `"abc" "bcd" "cde" "def"`},
		{`Abcdef`, `"abc" "bcd" "cde" "def"`},
		{`(abc)(def)`, `"abc" "bcd" "cde" "def"`},
		{`abc.*(def|ghi)`, `"abc" ("def"|"ghi")`},
		{`abc(def|ghi)`, `("abc" "bcd" "cde" "def")|("abc" "bcg" "cgh" "ghi")`},
		{`a+hello`, `"ahe" "ell" "hel" "llo"`},
		{`(a+hello|b+world)`, `("ahe" "ell" "hel" "llo")|("bwo" "orl" "rld" "wor")`},
		{`a*bbb`, `"bbb"`},
		{`a?bbb`, `("abb" "bbb")|("bbb")`},
		{`ab[cde]f`, `("abc" "bcf")|("abd" "bdf")|("abe" "bef")`},
		{`(abc|bac)de`, `("abc" "bcd" "cde")|("acd" "bac" "cde")`},
		{`ab(cab|cat)`, `("abc" "bca" "cab")|("abc" "bca" "cat")`},
		{`(a|ab)cde`, `("abc" "bcd" "cde")|("acd" "cde")`},
		{`[ab][cd][ef]`, `("ace"|"acf"|"ade"|"adf"|"bce"|"bcf"|"bde"|"bdf")`},
		{`a{3}bc`, `"aaa" "aab" "abc"`},

		// Anchors and word boundaries match the empty string, so they add nothing to the
		// query and take nothing away from it either.
		{`^abc`, `"abc"`},
		{`abc$`, `"abc"`},
		{`^abc$`, `"abc"`},
		{`\babc`, `"abc"`},
		{`abc\b`, `"abc"`},

		// The index stores lowercase, so a case-insensitive expression collapses to the same
		// query as the case-sensitive one rather than expanding into every case variant.
		{`(?i)abc`, `"abc"`},
		{`(?i)ABC`, `"abc"`},
		{`(?i)abcd`, `"abc" "bcd"`},
		{`(?i)abc|def`, `("abc"|"def")`},

		// A rune that case-folds to a letter it does not lowercase to still has to be allowed
		// for, or the index would miss documents the filter matches. Here 'ſ' folds to 's'
		// but lowercases to itself, so the query accepts either spelling of the first
		// trigram. The second branch holds the trigrams of "ſto", whose leading rune is two
		// bytes long.
		{`(?i)stone`, `"one" ("sto" "ton")|("ton" "\xbfto" "ſt")`},

		// Nothing can match, so nothing needs to be scanned.
		{`[^\s\S]`, `-`},

		// No required run of three or more characters, so the index cannot help and the
		// query has to fall back to a scan.
		{`ab.f`, `+`},
		{`ab[^cde]f`, `+`},
		{`.`, `+`},
		{`.*`, `+`},
		{`.*abc.*`, `"abc"`},
		{`\b`, `+`},
		{`()`, `+`},
		{`(a|b|c|d)(ef|g|hi|j)`, `+`},
		{`ab`, `+`},
		{`a*`, `+`},

		// A character class of more than 100 characters is deliberately overestimated as
		// "any character" rather than enumerated.
		{`a[\x{0100}-\x{0200}]c`, `+`},

		// An invalid expression is left for the filter itself to report, so it falls back to
		// a scan rather than failing here.
		{`a(b`, `+`},
	}

	for _, test := range tests {
		t.Run(test.pattern, func(t *testing.T) {
			assert.Equal(t, test.expected, regexTrigramQuery(test.pattern).String())
		})
	}
}

// A cross product of alternations is bounded rather than enumerated. The expression arrives
// in a user query, so this is a limit on what a single query can cost the server, not a
// tuning choice: enumerated, 32 groups of 5 would be 5^32 strings.
func TestRegexTrigramQuery_LargeCrossProduct_StaysBounded(t *testing.T) {
	for _, groups := range []int{2, 4, 8, 16, 32} {
		pattern := strings.Repeat("(abc|def|ghi|jkl|mno)", groups)
		q := regexTrigramQuery(pattern)

		assert.LessOrEqual(t, countTrigrams(q), 100*groups)
	}
}

func countTrigrams(q *trigramQuery) int {
	n := len(q.trigram)
	for _, sub := range q.sub {
		n += countTrigrams(sub)
	}
	return n
}

func TestLikeTrigramQuery(t *testing.T) {
	tests := []struct {
		pattern  string
		expected string
	}{
		// The three wildcard placements all require the value to contain the literal.
		{`%abcd%`, `"abc" "bcd"`},
		{`%abcd`, `"abc" "bcd"`},
		{`abcd%`, `"abc" "bcd"`},
		{`abcd`, `"abc" "bcd"`},

		// The index stores lowercase, so _like and _ilike produce the same query and the
		// case-sensitive form is narrowed again by the recheck.
		{`%ABCD%`, `"abc" "bcd"`},

		// A wildcard in the middle requires both runs.
		{`abcd%wxyz`, `"abc" "bcd" "wxy" "xyz"`},
		{`abcd%z`, `"abc" "bcd"`},
		{`a%wxyz`, `"wxy" "xyz"`},

		// connor's like treats a pattern with more than one interior wildcard as a plain
		// string, so the wildcards are part of the required literal here too.
		{`ab%cd%ef`, `"%cd" "%ef" "ab%" "b%c" "cd%" "d%e"`},

		// Too short to have a trigram, so the query falls back to a scan.
		{`%ab%`, `+`},
		{`%`, `+`},
		{`%%`, `+`},
		{``, `+`},
		{`a%b`, `+`},
	}

	for _, test := range tests {
		t.Run(test.pattern, func(t *testing.T) {
			assert.Equal(t, test.expected, likeTrigramQuery(test.pattern).String())
		})
	}
}
