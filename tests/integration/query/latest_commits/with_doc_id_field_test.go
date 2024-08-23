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
							"cid":   "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
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
