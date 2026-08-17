// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
)

func TestParseDerivedIndexes_OnFields_ProducesTypedRequests(t *testing.T) {
	cases := []indexTestCase{
		{
			description: "full-text defaults",
			sdl:         `type article { body: String @fullTextIndex }`,
			targetDescriptions: []client.NewIndexRequest{{
				Fields: []client.IndexedFieldDescription{{Name: "body"}},
				FullText: &client.FullTextIndexDescription{
					Algorithm: client.FullTextAlgorithmBM25,
					BM25: &client.BM25Params{
						K1: client.DefaultBM25K1,
						B:  client.DefaultBM25B,
					},
				},
			}},
		},
		{
			description: "full-text explicit BM25 parameters",
			sdl:         `type article { body: String @fullTextIndex(BM25: {k1: 2, b: 0.5}) }`,
			targetDescriptions: []client.NewIndexRequest{{
				Fields: []client.IndexedFieldDescription{{Name: "body"}},
				FullText: &client.FullTextIndexDescription{
					Algorithm: client.FullTextAlgorithmBM25,
					BM25:      &client.BM25Params{K1: 2, B: 0.5},
				},
			}},
		},
		{
			description: "trigram",
			sdl:         `type article { title: String @trigramIndex }`,
			targetDescriptions: []client.NewIndexRequest{{
				Fields:  []client.IndexedFieldDescription{{Name: "title"}},
				Trigram: &client.TrigramIndexDescription{},
			}},
		},
	}

	for _, test := range cases {
		parseIndexAndTest(t, test)
	}
}
