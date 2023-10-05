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
					// bae-80eded16-ee4b-5c9d-b33f-6a7b83958af2
					Doc: `{
						"name": "100 Go Mistakes to Avoid.",
						"rating": 4.8,
						"publisher_id": "bae-176ebdf0-77e7-5b2f-91ae-f620e37a29e3"
					}`,
				},
				testUtils.CreateDoc{
					// Authors
					CollectionID: 1,
					// bae-2f80f359-535d-508e-ba58-088a309ce3c3
					Doc: `{
						"name": "Teiva Harsanyi",
						"age": 48,
						"verified": true,
						"wrote_id": "bae-80eded16-ee4b-5c9d-b33f-6a7b83958af2"
					}`,
				},
				testUtils.CreateDoc{
					// Publishers
					CollectionID: 2,
					// bae-176ebdf0-77e7-5b2f-91ae-f620e37a29e3
					Doc: `{
						"name": "Manning Early Access Program (MEAP)",
						"address": "Online"
					}`,
				},
				testUtils.Request{
					Request: `mutation {
						delete_Author(docID: "bae-2f80f359-535d-508e-ba58-088a309ce3c3") {
							_docID
						}
					}`,
					Results: []map[string]any{
						{
							"_docID": "bae-2f80f359-535d-508e-ba58-088a309ce3c3",
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
					// bae-80eded16-ee4b-5c9d-b33f-6a7b83958af2
					Doc: `{
						"name": "100 Go Mistakes to Avoid.",
						"rating": 4.8,
						"publisher_id": "bae-176ebdf0-77e7-5b2f-91ae-f620e37a29e3"
					}`,
				},
				testUtils.CreateDoc{
					// Authors
					CollectionID: 1,
					// bae-2f80f359-535d-508e-ba58-088a309ce3c3
					Doc: `{
						"name": "Teiva Harsanyi",
						"age": 48,
						"verified": true,
						"wrote_id": "bae-80eded16-ee4b-5c9d-b33f-6a7b83958af2"
					}`,
				},
				testUtils.CreateDoc{
					// Publishers
					CollectionID: 2,
					// bae-176ebdf0-77e7-5b2f-91ae-f620e37a29e3
					Doc: `{
						"name": "Manning Early Access Program (MEAP)",
						"address": "Online"
					}`,
				},
				testUtils.Request{
					Request: `mutation {
						delete_Author(docID: "bae-2f80f359-535d-508e-ba58-088a309ce3c3") {
							AliasOfKey: _docID
						}
					}`,
					Results: []map[string]any{
						{
							"AliasOfKey": "bae-2f80f359-535d-508e-ba58-088a309ce3c3",
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
					// bae-80eded16-ee4b-5c9d-b33f-6a7b83958af2
					Doc: `{
						"name": "100 Go Mistakes to Avoid.",
						"rating": 4.8,
						"publisher_id": "bae-176ebdf0-77e7-5b2f-91ae-f620e37a29e3"
					}`,
				},
				testUtils.CreateDoc{
					// Authors
					CollectionID: 1,
					// bae-2f80f359-535d-508e-ba58-088a309ce3c3
					Doc: `{
						"name": "Teiva Harsanyi",
						"age": 48,
						"verified": true,
						"wrote_id": "bae-80eded16-ee4b-5c9d-b33f-6a7b83958af2"
					}`,
				},
				testUtils.CreateDoc{
					// Publishers
					CollectionID: 2,
					// bae-176ebdf0-77e7-5b2f-91ae-f620e37a29e3
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
						delete_Author(docID: "bae-2f80f359-535d-508e-ba58-088a309ce3c3") {
							Key: _docID
						}
					}`,
					Results: []map[string]any{
						{
							"Key": "bae-2f80f359-535d-508e-ba58-088a309ce3c3",
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
