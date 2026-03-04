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

package remove

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdate_RemoveVectorEmbedding_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						about: String
						name_v: [Float32!] @embedding(fields: ["name", "about"], provider: "openai", model: "text-embedding-3-small",  url: "https://api.openai.com/v1")
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "/Users/VectorEmbeddings/0"
						}
					]
				`,
			},
			&action.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:             "Users",
						IsMaterialized:   true,
						IsActive:         true,
						VectorEmbeddings: []client.VectorEmbeddingDescription{},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
