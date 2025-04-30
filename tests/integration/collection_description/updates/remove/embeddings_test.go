// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package remove

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColDescrUpdate_RemoveVectorEmbedding_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						about: String
						name_v: [Float32!] @embedding(fields: ["name", "about"], provider: "openai", model: "text-embedding-3-small",  url: "https://api.openai.com/v1")
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "/bafkreigf66gyhrju7qebw6wxe7qrnzqfegcqxizp5jsk3qnnpv3ronrcza/VectorEmbeddings/0"
						}
					]
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionDescription{
					{
						Name:             "Users",
						IsMaterialized:   true,
						VectorEmbeddings: []client.VectorEmbeddingDescription{},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
