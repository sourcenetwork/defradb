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
					latestCommits(docID: "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3") {
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

func TestQueryLatestCommitsWithDocIDWithSchemaVersionIDField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple latest commits query with docID and schema versiion id field",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.Request{
				Request: `query {
					latestCommits(docID: "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3") {
						cid
						schemaVersionId
					}
				}`,
				Results: map[string]any{
					"latestCommits": []map[string]any{
						{
							"cid":             "bafyreia2vlbfkcbyogdjzmbqcjneabwwwtw7ti2xbd7yor5mbu2sk4pcoy",
							"schemaVersionId": "bafkreicprhqxzlw3akyssz2v6pifwfueavp7jq2yj3dghapi3qcq6achs4",
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
					history: latestCommits(docID: "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3") {
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
