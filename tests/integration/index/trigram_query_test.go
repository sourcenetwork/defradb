// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// trigramQueryNames are the documents every query test in this file runs against. They cover
// an uppercase value, values sharing a substring, and a value too short to have a trigram.
var trigramQueryNames = []string{
	"Islam", "ISLAMABAD", "Andy", "Shahzad", "Fred", "Orpheus", "Chris", "Jo", "John",
}

// assertTrigramQuery runs one filter over the same documents twice: once on a collection with
// a trigram index on name, and once on a collection without one. Both must return expected.
//
// This is the assertion that matters for a trigram index. It is a candidate generator, and
// every candidate is rechecked against the original filter, so leaving the answer unchanged
// is its whole contract - whether it narrowed the scan or gave up and fell back to one.
func assertTrigramQuery(t *testing.T, filter string, expected []string) {
	t.Helper()

	results := make([]map[string]any, 0, len(expected))
	for _, name := range expected {
		results = append(results, map[string]any{"name": name})
	}

	for _, indexDirective := range []string{"", "@index(kind: TRIGRAM)"} {
		actions := []any{
			&action.AddCollection{
				SDL: `type User { name: String ` + indexDirective + ` }`,
			},
		}
		for _, name := range trigramQueryNames {
			actions = append(actions, &action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "` + name + `"}`,
			})
		}
		actions = append(actions, &action.Request{
			Request:           `query { User(filter: ` + filter + `) { name } }`,
			Results:           map[string]any{"User": results},
			NonOrderedResults: true,
		})

		testUtils.ExecuteTestCase(t, testUtils.TestCase{Actions: actions})
	}
}

func TestTrigramIndexQuery_WithLike_ReturnsSameAsFullScan(t *testing.T) {
	// Wildcards on both sides: a substring anywhere in the value.
	assertTrigramQuery(t, `{name: {_like: "%sla%"}}`, []string{"Islam"})
	// A leading wildcard: the literal must be at the end of the value.
	assertTrigramQuery(t, `{name: {_like: "%eus"}}`, []string{"Orpheus"})
	// A trailing wildcard: the literal must be at the start of the value.
	assertTrigramQuery(t, `{name: {_like: "Chr%"}}`, []string{"Chris"})
	// A wildcard in the middle: both runs must be present.
	assertTrigramQuery(t, `{name: {_like: "Sha%zad"}}`, []string{"Shahzad"})
	// No wildcard at all is an equality test on the whole value.
	assertTrigramQuery(t, `{name: {_like: "Fred"}}`, []string{"Fred"})
	// The index holds lowercase, so this one over-returns and the recheck drops the
	// uppercase document again.
	assertTrigramQuery(t, `{name: {_like: "%SLA%"}}`, []string{"ISLAMABAD"})
	assertTrigramQuery(t, `{name: {_like: "%zzz%"}}`, nil)
}

func TestTrigramIndexQuery_WithILike_ReturnsSameAsFullScan(t *testing.T) {
	assertTrigramQuery(t, `{name: {_ilike: "%sla%"}}`, []string{"Islam", "ISLAMABAD"})
	assertTrigramQuery(t, `{name: {_ilike: "%SLA%"}}`, []string{"Islam", "ISLAMABAD"})
	assertTrigramQuery(t, `{name: {_ilike: "islamabad"}}`, []string{"ISLAMABAD"})
	assertTrigramQuery(t, `{name: {_ilike: "chr%"}}`, []string{"Chris"})
}

func TestTrigramIndexQuery_WithRegex_ReturnsSameAsFullScan(t *testing.T) {
	assertTrigramQuery(t, `{name: {_regex: "sla"}}`, []string{"Islam"})
	// Anchors match the empty string, so they neither add to nor take from the trigrams.
	assertTrigramQuery(t, `{name: {_regex: "^Chris$"}}`, []string{"Chris"})
	assertTrigramQuery(t, `{name: {_regex: "^Islam"}}`, []string{"Islam"})
	assertTrigramQuery(t, `{name: {_regex: "eus$"}}`, []string{"Orpheus"})
	// Alternation becomes an OR of the branches' trigrams.
	assertTrigramQuery(t, `{name: {_regex: "Chris|Fred"}}`, []string{"Chris", "Fred"})
	assertTrigramQuery(t, `{name: {_regex: "^(Chris|Orpheus)$"}}`, []string{"Chris", "Orpheus"})
	assertTrigramQuery(t, `{name: {_regex: "(?i)islam"}}`, []string{"Islam", "ISLAMABAD"})
	assertTrigramQuery(t, `{name: {_regex: "Isl.m"}}`, []string{"Islam"})
	assertTrigramQuery(t, `{name: {_regex: "zzz"}}`, nil)
}

// A pattern with no run of three characters yields no trigrams, so the query has to fall back
// to a full scan. It must still answer correctly, including for documents too short to have
// an index entry of their own.
func TestTrigramIndexQuery_WithPatternTooShortForTrigrams_ReturnsSameAsFullScan(t *testing.T) {
	assertTrigramQuery(t, `{name: {_like: "%Jo%"}}`, []string{"Jo", "John"})
	assertTrigramQuery(t, `{name: {_like: "Jo"}}`, []string{"Jo"})
	assertTrigramQuery(t, `{name: {_ilike: "%jo%"}}`, []string{"Jo", "John"})
	assertTrigramQuery(t, `{name: {_regex: "^Jo$"}}`, []string{"Jo"})
	assertTrigramQuery(t, `{name: {_regex: "^J.$"}}`, []string{"Jo"})
	assertTrigramQuery(t, `{name: {_regex: "n$"}}`, []string{"John"})
	// Every document matches, and a value under three bytes has no index entry at all, so
	// this is only correct if the query gave up on the index entirely.
	assertTrigramQuery(t, `{name: {_regex: ".*"}}`, trigramQueryNames)
}

// A negated form cannot use the index, since no candidate set can express "every document
// that does not match", but it must still return the right documents.
func TestTrigramIndexQuery_WithNegatedForm_ReturnsSameAsFullScan(t *testing.T) {
	assertTrigramQuery(t, `{name: {_nlike: "%sla%"}}`,
		[]string{"ISLAMABAD", "Andy", "Shahzad", "Fred", "Orpheus", "Chris", "Jo", "John"})
	assertTrigramQuery(t, `{name: {_nilike: "%sla%"}}`,
		[]string{"Andy", "Shahzad", "Fred", "Orpheus", "Chris", "Jo", "John"})
	assertTrigramQuery(t, `{_not: {name: {_like: "%sla%"}}}`,
		[]string{"ISLAMABAD", "Andy", "Shahzad", "Fred", "Orpheus", "Chris", "Jo", "John"})
	assertTrigramQuery(t, `{_not: {name: {_regex: "^(Chris|Fred|Jo|John|Andy|Orpheus)$"}}}`,
		[]string{"Islam", "ISLAMABAD", "Shahzad"})
}

func TestTrigramIndexQuery_WithCompoundFilter_ReturnsSameAsFullScan(t *testing.T) {
	assertTrigramQuery(t, `{_and: [{name: {_like: "%sla%"}}, {name: {_like: "%lam"}}]}`,
		[]string{"Islam"})
	assertTrigramQuery(t, `{_or: [{name: {_like: "%sla%"}}, {name: {_like: "%hri%"}}]}`,
		[]string{"Islam", "Chris"})
	// One branch of the _or has no usable trigram, so the whole filter has to fall back.
	assertTrigramQuery(t, `{_or: [{name: {_like: "%sla%"}}, {name: {_like: "%Jo%"}}]}`,
		[]string{"Islam", "Jo", "John"})
	assertTrigramQuery(t, `{_and: [{name: {_like: "%a%"}}, {name: {_regex: "^Sha"}}]}`,
		[]string{"Shahzad"})
	assertTrigramQuery(t, `{_and: [{name: {_ilike: "%sla%"}}, {_not: {name: {_like: "%SLA%"}}}]}`,
		[]string{"Islam"})
	assertTrigramQuery(t, `{_or: [{name: {_regex: "^Chris"}}, {_not: {name: {_like: "%a%"}}}]}`,
		[]string{"ISLAMABAD", "Andy", "Fred", "Orpheus", "Chris", "Jo", "John"})
}

// A trigram index scans in the order of the trigrams, not the order of the values, so it must
// not be treated as satisfying an ordering on its own field.
func TestTrigramIndexQuery_WithOrderOnTheIndexedField_ReturnsOrderedResults(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @index(kind: TRIGRAM)
					}
				`,
			},
			&action.AddDoc{CollectionID: 0, Doc: `{"name": "banana"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name": "apricot"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name": "cherry"}`},
			&action.Request{
				Request: `query {
					User(filter: {name: {_regex: "a|e"}}, order: {name: ASC}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "apricot"},
						{"name": "banana"},
						{"name": "cherry"},
					},
				},
			},
			&action.Request{
				Request: `query {
					User(order: {name: DESC}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "cherry"},
						{"name": "banana"},
						{"name": "apricot"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// An _or branch on another field cannot be answered by the trigram index, so the whole filter
// has to fall back to a scan. The candidate set of the branch it can answer would be missing
// every document matched only by the other branch.
func TestTrigramIndexQuery_WithOrBranchOnAnotherField_ReturnsSameAsFullScan(t *testing.T) {
	for _, filter := range []string{
		`{_or: [{name: {_like: "%sla%"}}, {age: {_eq: 30}}]}`,
		`{_and: [{_or: [{name: {_like: "%sla%"}}, {age: {_eq: 30}}]}]}`,
	} {
		test := testUtils.TestCase{
			Actions: []any{
				&action.AddCollection{
					SDL: `
						type User {
							name: String @index(kind: TRIGRAM)
							age: Int
						}
					`,
				},
				&action.AddDoc{CollectionID: 0, Doc: `{"name": "Islam", "age": 20}`},
				&action.AddDoc{CollectionID: 0, Doc: `{"name": "Andy", "age": 30}`},
				&action.AddDoc{CollectionID: 0, Doc: `{"name": "Fred", "age": 40}`},
				&action.Request{
					Request: `query { User(filter: ` + filter + `) { name } }`,
					Results: map[string]any{
						"User": []map[string]any{
							{"name": "Islam"},
							{"name": "Andy"},
						},
					},
					NonOrderedResults: true,
				},
			},
		}

		testUtils.ExecuteTestCase(t, test)
	}
}

// The point of the index is that it narrows the scan, so check that it is actually reached
// and that it gives up when it should.
func TestTrigramIndexQuery_UsesTheIndexOnlyWhenTheFilterYieldsTrigrams(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @index(kind: TRIGRAM)
					}
				`,
			},
			&action.AddDoc{CollectionID: 0, Doc: `{"name": "Islam"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name": "ISLAMABAD"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name": "Chris"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name": "Jo"}`},
			&action.Request{
				// Both documents hold the trigram "sla", and only one survives the recheck.
				Request: makeExplainQuery(`query {
					User(filter: {name: {_like: "%sla%"}}) {
						name
					}
				}`),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
			&action.Request{
				// "isl" and "sla" are both required, so both posting lists are walked in
				// step: two reads per document that survives the intersection.
				Request: makeExplainQuery(`query {
					User(filter: {name: {_regex: "isla"}}) {
						name
					}
				}`),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(4),
			},
			&action.Request{
				// Too short for a trigram, so the index is not read at all.
				Request: makeExplainQuery(`query {
					User(filter: {name: {_like: "%Jo%"}}) {
						name
					}
				}`),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
			&action.Request{
				// A negated form is never served by the index.
				Request: makeExplainQuery(`query {
					User(filter: {name: {_nlike: "%sla%"}}) {
						name
					}
				}`),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
