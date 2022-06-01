// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package relation_delete

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	relationTests "github.com/sourcenetwork/defradb/tests/integration/mutation/relation"
)

type dataMap = map[string]interface{}

func TestExplainRelationalDeletionOfADocumentUsingSingleKey_Success(t *testing.T) {
	tests := []testUtils.QueryTestCase{

		{
			Description: "Explain relational delete of an updated document and an aliased _key name.",

			Query: `mutation @explain {
						delete_author(id: "bae-2f80f359-535d-508e-ba58-088a309ce3c3") {
							Key: _key
					}
				}`,

			Docs: map[int][]string{
				// Books
				0: {
					// bae-80eded16-ee4b-5c9d-b33f-6a7b83958af2
					(`{
						"name": "100 Go Mistakes to Avoid.",
						"rating": 4.8,
						"publisher_id": "bae-176ebdf0-77e7-5b2f-91ae-f620e37a29e3"
					}`),
				},

				// Authors
				1: {
					// bae-2f80f359-535d-508e-ba58-088a309ce3c3
					(`{
					"name": "Teiva Harsanyi",
					"age": 48,
					"verified": true,
					"wrote_id": "bae-80eded16-ee4b-5c9d-b33f-6a7b83958af2"
					}`),
				},

				// Publishers
				2: {
					// bae-176ebdf0-77e7-5b2f-91ae-f620e37a29e3
					(`{
						"name": "Manning Early Access Program (MEAP)",
						"address": "Online"
					}`),

					// bae-5c599633-d6d2-56ae-b3f0-1b65b4cee9fe
					(`{
						"name": "Manning Publications",
						"address": "Website"
					}`),
				},
			},

			Results: []dataMap{
				{
					"explain": dataMap{
						"selectTopNode": dataMap{
							"selectNode": dataMap{
								"deleteNode": dataMap{
									"filter": nil,
									"ids": []string{
										"bae-2f80f359-535d-508e-ba58-088a309ce3c3",
									},
								},
								"filter": nil,
							},
						},
					},
				},
			},

			ExpectedError: "",
		},
	}

	for _, test := range tests {
		relationTests.ExecuteTestCase(t, test)
	}
}
