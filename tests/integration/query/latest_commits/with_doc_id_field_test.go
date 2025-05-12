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

// This test is for documentation reasons only. This is not
// desired behaviour (it looks totally broken to me).
func TestQueryLatestCommitsWithDocIDAndFieldName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple latest commits query with docID and field name",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
				Request: `query {
					latestCommits(docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7", fieldName: "age") {
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
							"cid":   "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
							"links": []map[string]any{},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryLatestCommitsWithDocIDAndFieldId(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple latest commits query with docID and field id",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
				Request: `query {
					latestCommits(docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7", fieldName: "1") {
						cid
						links {
							cid
							name
						}
					}
				}`,
				Results: map[string]any{
					"latestCommits": []map[string]any{},
				},
			},
		},
	}

	executeTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (Users should not be specifying field ids).
func TestQueryLatestCommitsWithDocIDAndCompositeFieldId(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple latest commits query with docID and composite field id",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
				Request: `query {
					latestCommits(docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7", fieldName: "_C") {
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
