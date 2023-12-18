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

func TestQueryCommitsWithDockeyAndUnknownField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID and unknown field",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", fieldId: "not a field") {
							cid
						}
					}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndUnknownFieldId(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID and unknown field id",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", fieldId: "999999") {
							cid
						}
					}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (should return all commits for docID-field).
func TestQueryCommitsWithDockeyAndField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID and field",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", fieldId: "Age") {
							cid
						}
					}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (Users should not be specifying field ids).
func TestQueryCommitsWithDockeyAndFieldId(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID and field id",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", fieldId: "1") {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeiazsz3twea2uxpen6452qqa7qnzp2xildfxliidhqk632jpvbixkm",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (Users should not be specifying field ids).
func TestQueryCommitsWithDockeyAndCompositeFieldId(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID and field id",
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", fieldId: "C") {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeihbcl2ijavd6vdcj4vgunw4q5qt5itmumxw7iy7fhoqfsuvkpkqeq",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
