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
					latestCommits(docID: "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3", fieldId: "age") {
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
					latestCommits(docID: "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3", fieldId: "1") {
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
							"cid":   "bafyreif6dqbkr7t37jcjfxxrjnxt7cspxzvs7qwlbtjca57cc663he4s7e",
							"links": []map[string]any{},
						},
					},
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
					latestCommits(docID: "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3", fieldId: "C") {
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
							"cid": "bafyreia2vlbfkcbyogdjzmbqcjneabwwwtw7ti2xbd7yor5mbu2sk4pcoy",
							"links": []map[string]any{
								{
									"cid":  "bafyreif6dqbkr7t37jcjfxxrjnxt7cspxzvs7qwlbtjca57cc663he4s7e",
									"name": "age",
								},
								{
									"cid":  "bafyreigtnj6ntulcilkmin4pgukjwv3nwglqpiiyddz3dyfexdbltze7sy",
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
