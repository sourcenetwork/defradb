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

func TestRelationalDeletionOfADocumentUsingSingleKey_Success(t *testing.T) {
	tests := []testUtils.RequestTestCase{

		{
			Description: "Relational delete mutation where one element exists.",
			Request: `mutation {
						delete_author(id: "bae-2f80f359-535d-508e-ba58-088a309ce3c3") {
							_key
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
				}`)},
				// Authors
				1: {
					// bae-2f80f359-535d-508e-ba58-088a309ce3c3
					(`{
					"name": "Teiva Harsanyi",
					"age": 48,
					"verified": true,
					"wrote_id": "bae-80eded16-ee4b-5c9d-b33f-6a7b83958af2"
				}`)},
				// Publishers
				2: {
					// bae-176ebdf0-77e7-5b2f-91ae-f620e37a29e3
					(`{
					"name": "Manning Early Access Program (MEAP)",
					"address": "Online"
				}`)},
			},
			Results: []map[string]interface{}{
				{
					"_key": "bae-2f80f359-535d-508e-ba58-088a309ce3c3",
				},
			},
			ExpectedError: "",
		},

		{
			Description: "Relational delete mutation with an aliased _key name.",
			Request: `mutation {
						delete_author(id: "bae-2f80f359-535d-508e-ba58-088a309ce3c3") {
							AliasOfKey: _key
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
				}`)},
				// Authors
				1: {
					// bae-2f80f359-535d-508e-ba58-088a309ce3c3
					(`{
					"name": "Teiva Harsanyi",
					"age": 48,
					"verified": true,
					"wrote_id": "bae-80eded16-ee4b-5c9d-b33f-6a7b83958af2"
				}`)},
				// Publishers
				2: {
					// bae-176ebdf0-77e7-5b2f-91ae-f620e37a29e3
					(`{
					"name": "Manning Early Access Program (MEAP)",
					"address": "Online"
				}`)},
			},
			Results: []map[string]interface{}{
				{
					"AliasOfKey": "bae-2f80f359-535d-508e-ba58-088a309ce3c3",
				},
			},
			ExpectedError: "",
		},

		{
			Description: "Relational Delete of an updated document and an aliased _key name.",
			Request: `mutation {
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
			Updates: map[int][]string{
				0: {
					(`{
						"name": "Rust in Action.",
						"publisher_id": "bae-5c599633-d6d2-56ae-b3f0-1b65b4cee9fe"
					}`)},
			},
			Results: []map[string]interface{}{
				{
					"Key": "bae-2f80f359-535d-508e-ba58-088a309ce3c3",
				},
			},
			ExpectedError: "",
		},
	}

	for _, test := range tests {
		relationTests.ExecuteTestCase(t, test)
	}
}
