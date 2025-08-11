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
				Results: map[string]any{
					"commits": []map[string]any{},
				},
				ExpectedError: "missing cid",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndCidForDifferentDocWithUpdate(t *testing.T) {
	test := testUtils.TestCase{
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
							cid: "bafyreiale6qsjc7qewod3c6h2odwamfwcf7vt4zlqtw7ldcm57xdkgxja4"
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{},
				},
				ExpectedError: "cid does not belong to document",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndCidWithUpdate(t *testing.T) {
	test := testUtils.TestCase{
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
							cid: "bafyreiale6qsjc7qewod3c6h2odwamfwcf7vt4zlqtw7ldcm57xdkgxja4"
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreiale6qsjc7qewod3c6h2odwamfwcf7vt4zlqtw7ldcm57xdkgxja4",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndCidWithUpdateAndDepth(t *testing.T) {
	test := testUtils.TestCase{
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
			// depth is pretty arbitrary here, as long as its big enough to cover the updates
			// from the target cid (ie >=2)
			testUtils.Request{
				Request: ` {
						commits(
							docID: "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
							cid: "bafyreiale6qsjc7qewod3c6h2odwamfwcf7vt4zlqtw7ldcm57xdkgxja4",
							depth: 5
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreiale6qsjc7qewod3c6h2odwamfwcf7vt4zlqtw7ldcm57xdkgxja4",
						},
						{
							"cid": "bafyreia2vlbfkcbyogdjzmbqcjneabwwwtw7ti2xbd7yor5mbu2sk4pcoy",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
