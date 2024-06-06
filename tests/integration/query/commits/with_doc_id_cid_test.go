// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package commits

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsWithDocIDAndCidForDifferentDoc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID and cid, for different doc",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: ` {
						commits(
							docID: "bae-not-this-doc",
							cid: "bafybeica4js2abwqjjrz7dcialbortbz32uxp7ufxu7yljbwvmhjqqxzny"
						) {
							cid
						}
					}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndCidForDifferentDocWithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID and cid, for different doc with update",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: ` {
						commits(
							docID: "bae-not-this-doc",
							cid: "bafybeica4js2abwqjjrz7dcialbortbz32uxp7ufxu7yljbwvmhjqqxzny"
						) {
							cid
						}
					}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndCidWithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID and cid, with update",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: ` {
						commits(
							docID: "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
							cid: "bafyreif632ewkphjjwxcthemgbkgtm25faw22mvw7eienu5gnazrao33ba"
						) {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafyreif632ewkphjjwxcthemgbkgtm25faw22mvw7eienu5gnazrao33ba",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
