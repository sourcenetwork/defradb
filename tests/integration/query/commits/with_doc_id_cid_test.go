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
							cid: "bafyreido4fwolghako5ogh4jcy6tr3butjicfwubk27uyuimlm366rtdmy"
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
							docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7",
							cid: "bafyreido4fwolghako5ogh4jcy6tr3butjicfwubk27uyuimlm366rtdmy"
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreido4fwolghako5ogh4jcy6tr3butjicfwubk27uyuimlm366rtdmy",
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
			// depth is pretty arbitrary here, as long as its big enough to cover the updates
			// from the target cid (ie >=2)
			testUtils.Request{
				Request: ` {
						commits(
							docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7",
							cid: "bafyreido4fwolghako5ogh4jcy6tr3butjicfwubk27uyuimlm366rtdmy",
							depth: 5
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreido4fwolghako5ogh4jcy6tr3butjicfwubk27uyuimlm366rtdmy",
						},
						{
							"cid": "bafyreidtdklweht7ainl5rrdeqscr3cwr72sr4lehzrpmmnnbvnvstavnm",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
