// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_one_to_one

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestRelationalDeletionOfADocumentUsingSingleKey_Success(t *testing.T) {
	tests := []testUtils.TestCase{
		{
			Description: "Relational delete mutation where one element exists.",
			Actions: []any{
				testUtils.CreateDoc{
					// Books
					CollectionID: 0,
					// bae-8e8b2923-e167-5fd9-aee6-98267dd0ab40
					Doc: `{
						"name": "100 Go Mistakes to Avoid.",
						"rating": 4.8,
						"publisher_id": "bae-9c689bec-071e-5650-9378-bc11d5d3325c"
					}`,
				},
				testUtils.CreateDoc{
					// Authors
					CollectionID: 1,
					// bae-b4f1fb22-52f2-5e3d-950c-f6a4033d8f48
					Doc: `{
						"name": "Teiva Harsanyi",
						"age": 48,
						"verified": true,
						"wrote_id": "bae-8e8b2923-e167-5fd9-aee6-98267dd0ab40"
					}`,
				},
				testUtils.CreateDoc{
					// Publishers
					CollectionID: 2,
					// bae-9c689bec-071e-5650-9378-bc11d5d3325c
					Doc: `{
						"name": "Manning Early Access Program (MEAP)",
						"address": "Online"
					}`,
				},
				testUtils.Request{
					Request: `mutation {
						delete_Author(docID: "bae-b4f1fb22-52f2-5e3d-950c-f6a4033d8f48") {
							_docID
						}
					}`,
					Results: map[string]any{
						"delete_Author": []map[string]any{
							{
								"_docID": "bae-b4f1fb22-52f2-5e3d-950c-f6a4033d8f48",
							},
						},
					},
				},
			},
		},

		{
			Description: "Relational delete mutation with an aliased _docID name.",
			Actions: []any{
				testUtils.CreateDoc{
					// Books
					CollectionID: 0,
					// bae-8e8b2923-e167-5fd9-aee6-98267dd0ab40
					Doc: `{
						"name": "100 Go Mistakes to Avoid.",
						"rating": 4.8,
						"publisher_id": "bae-9c689bec-071e-5650-9378-bc11d5d3325c"
					}`,
				},
				testUtils.CreateDoc{
					// Authors
					CollectionID: 1,
					// bae-b4f1fb22-52f2-5e3d-950c-f6a4033d8f48
					Doc: `{
						"name": "Teiva Harsanyi",
						"age": 48,
						"verified": true,
						"wrote_id": "bae-8e8b2923-e167-5fd9-aee6-98267dd0ab40"
					}`,
				},
				testUtils.CreateDoc{
					// Publishers
					CollectionID: 2,
					// bae-9c689bec-071e-5650-9378-bc11d5d3325c
					Doc: `{
						"name": "Manning Early Access Program (MEAP)",
						"address": "Online"
					}`,
				},
				testUtils.Request{
					Request: `mutation {
						delete_Author(docID: "bae-b4f1fb22-52f2-5e3d-950c-f6a4033d8f48") {
							AliasOfKey: _docID
						}
					}`,
					Results: map[string]any{
						"delete_Author": []map[string]any{
							{
								"AliasOfKey": "bae-b4f1fb22-52f2-5e3d-950c-f6a4033d8f48",
							},
						},
					},
				},
			},
		},

		{
			Description: "Relational Delete of an updated document and an aliased _docID name.",
			Actions: []any{
				testUtils.CreateDoc{
					// Books
					CollectionID: 0,
					// bae-8e8b2923-e167-5fd9-aee6-98267dd0ab40
					Doc: `{
						"name": "100 Go Mistakes to Avoid.",
						"rating": 4.8,
						"publisher_id": "bae-9c689bec-071e-5650-9378-bc11d5d3325c"
					}`,
				},
				testUtils.CreateDoc{
					// Authors
					CollectionID: 1,
					// bae-b4f1fb22-52f2-5e3d-950c-f6a4033d8f48
					Doc: `{
						"name": "Teiva Harsanyi",
						"age": 48,
						"verified": true,
						"wrote_id": "bae-8e8b2923-e167-5fd9-aee6-98267dd0ab40"
					}`,
				},
				testUtils.CreateDoc{
					// Publishers
					CollectionID: 2,
					// bae-9c689bec-071e-5650-9378-bc11d5d3325c
					Doc: `{
						"name": "Manning Early Access Program (MEAP)",
						"address": "Online"
					}`,
				},
				testUtils.CreateDoc{
					// Publishers
					CollectionID: 2,
					// bae-5c599633-d6d2-56ae-b3f0-1b65b4cee9fe
					Doc: `{
						"name": "Manning Publications",
						"address": "Website"
					}`,
				},
				testUtils.UpdateDoc{
					CollectionID: 1,
					DocID:        0,
					Doc: `{
						"name": "Teiva Harsanyiiiiiiiiii",
						"age": 49
					}`,
				},
				testUtils.Request{
					Request: `mutation {
						delete_Author(docID: "bae-b4f1fb22-52f2-5e3d-950c-f6a4033d8f48") {
							Key: _docID
						}
					}`,
					Results: map[string]any{
						"delete_Author": []map[string]any{
							{
								"Key": "bae-b4f1fb22-52f2-5e3d-950c-f6a4033d8f48",
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		execute(t, test)
	}
}
