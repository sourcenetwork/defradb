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

package test_acp_dac_index

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// bm25Actions sets up a collection whose bio field carries a BM25 index, with the two best
// matches for "alpha" private to identity 1 and the next two readable by anyone.
func bm25Actions(requests ...any) []any {
	return append([]any{
		testUtils.AddDACPolicy{
			Identity: testUtils.ClientIdentity(1),
			Policy:   userPolicy,
		},
		&action.AddCollection{
			SDL: `
				type Users @policy(
					id: "{{.Policy0}}",
					resource: "users"
				) {
					name: String
					bio: String @index(kind: BM25)
				}
			`,
		},
		&action.AddDoc{
			Identity: testUtils.ClientIdentity(1),
			Doc:      `{"name": "private best", "bio": "alpha alpha alpha"}`,
		},
		&action.AddDoc{
			Identity: testUtils.ClientIdentity(1),
			Doc:      `{"name": "private second", "bio": "alpha alpha"}`,
		},
		&action.AddDoc{Doc: `{"name": "public best", "bio": "alpha"}`},
		&action.AddDoc{Doc: `{"name": "public second", "bio": "alpha beta gamma"}`},
		&action.AddDoc{Doc: `{"name": "no match", "bio": "beta"}`},
	}, requests...)
}

func bm25Results(names ...string) map[string]any {
	documents := make([]map[string]any, 0, len(names))
	for _, name := range names {
		documents = append(documents, map[string]any{
			"name": name,
			"rank": gomega.BeNumerically(">", 0.0),
		})
	}
	return map[string]any{"Users": documents}
}

// A limit still returns that many documents when the best matches are ones the caller may not
// read. The documents access control removes are dropped below the limit node, which goes on
// pulling until it has what it asked for, so the count is not quietly short.
func TestACPWithBM25Index_UponQueryingWithLimitWithoutIdentity_ReturnsThePermittedBest(t *testing.T) {
	test := testUtils.TestCase{
		Actions: bm25Actions(&action.Request{
			Request: `query {
				Users(order: {_alias: {rank: DESC}}, limit: 2) {
					name
					rank: _bm25(bio: {query: "alpha"})
				}
			}`,
			Results: bm25Results("public best", "public second"),
		}),
	}

	testUtils.ExecuteTestCase(t, test)
}

// The same query with the identity that may read them puts the two private documents on top.
func TestACPWithBM25Index_UponQueryingWithLimitWithIdentity_ReturnsTheBest(t *testing.T) {
	test := testUtils.TestCase{
		Actions: bm25Actions(&action.Request{
			Identity: testUtils.ClientIdentity(1),
			Request: `query {
				Users(order: {_alias: {rank: DESC}}, limit: 2) {
					name
					rank: _bm25(bio: {query: "alpha"})
				}
			}`,
			Results: bm25Results("private best", "private second"),
		}),
	}

	testUtils.ExecuteTestCase(t, test)
}
