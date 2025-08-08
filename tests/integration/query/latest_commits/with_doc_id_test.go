// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package latest_commits

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryLatestCommitsWithDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple latest commits query with docID",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
				Request: `query {
					latestCommits(docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7") {
						cid
						links {
							cid
							name
						}
					}
				}`,
				Results: map[string]any{
					"latestCommits": []map[string]any{
						{
							"cid": "bafyreidtdklweht7ainl5rrdeqscr3cwr72sr4lehzrpmmnnbvnvstavnm",
							"links": []map[string]any{
								{
									"cid":  "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
									"name": "age",
								},
								{
									"cid":  "bafyreiht7yhnnrgbwgyu5toe3exvpkovzrefzr6midu5secnlr546oel3q",
									"name": "name",
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryLatestCommitsWithDocIDWithSchemaVersionIDField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
				Request: `query {
					latestCommits(docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7") {
						cid
						schemaVersionId
					}
				}`,
				Results: map[string]any{
					"latestCommits": []map[string]any{
						{
							"cid":             "bafyreidtdklweht7ainl5rrdeqscr3cwr72sr4lehzrpmmnnbvnvstavnm",
							"schemaVersionId": "bafyreigk2gtae2irmijtkb7z736r3lpssqv7cvmbrp3p6x6ouw7nakc4nm",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryLatestCommits_WithDocIDAndAliased_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple latest commits query with docID and aliased",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
				Request: `query {
					history: latestCommits(docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7") {
						cid
						links {
							cid
							name
						}
					}
				}`,
				Results: map[string]any{
					"history": []map[string]any{
						{
							"cid": "bafyreidtdklweht7ainl5rrdeqscr3cwr72sr4lehzrpmmnnbvnvstavnm",
							"links": []map[string]any{
								{
									"cid":  "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
									"name": "age",
								},
								{
									"cid":  "bafyreiht7yhnnrgbwgyu5toe3exvpkovzrefzr6midu5secnlr546oel3q",
									"name": "name",
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
