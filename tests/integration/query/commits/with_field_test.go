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

// This test is for documentation reasons only. This is not
// desired behaviour (should return all commits for dockey-field).
func TestQueryCommitsWithField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with field",
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
				Request: `query {
						commits (fieldId: "Age") {
							cid
						}
					}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (Users should not be specifying field ids).
func TestQueryCommitsWithFieldId(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with field id",
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
				Request: `query {
						commits (fieldId: "1") {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeic5oodfpnixl6uf4bi63m3eouuhj3gafudlsd4tqryhx2wy7rczoe",
					},
				},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (Users should not be specifying field ids).
func TestQueryCommitsWithCompositeFieldId(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey and field id",
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
				Request: `query {
						commits(fieldId: "C") {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeig3wrpwi6q7vjchizcwnenslasyxop6wey7jahbiszlubdglfq2fq",
					},
				},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (Users should not be specifying field ids).
func TestQueryCommitsWithCompositeFieldIdWithReturnedSchemaVersionId(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey and field id",
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
				Request: `query {
						commits(fieldId: "C") {
							cid
							schemaVersionId
						}
					}`,
				Results: []map[string]any{
					{
						"cid":             "bafybeig3wrpwi6q7vjchizcwnenslasyxop6wey7jahbiszlubdglfq2fq",
						"schemaVersionId": "bafkreicihc56up4gzd4pf6lsmg5fc7dugyuigoaywgtjwy5c2suvj5zhtm",
					},
				},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}
