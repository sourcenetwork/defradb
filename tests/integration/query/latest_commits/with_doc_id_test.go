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
							"cid": "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
							"links": []map[string]any{
								{
									"cid":  "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
									"name": "name",
								},
								{
									"cid":  "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
									"name": "age",
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
							"cid":             "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
							"schemaVersionId": "bafkreicprhqxzlw3akyssz2v6pifwfueavp7jq2yj3dghapi3qcq6achs4",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
