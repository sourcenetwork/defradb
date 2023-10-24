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

func TestQueryLatestCommitsWithDocKey(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple latest commits query with dockey",
		Request: `query {
					latestCommits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
						cid
						links {
							cid
							name
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"age": 21
				}`,
			},
		},
		Results: []map[string]any{
			{
				"cid": "bafybeihbcl2ijavd6vdcj4vgunw4q5qt5itmumxw7iy7fhoqfsuvkpkqeq",
				"links": []map[string]any{
					{
						"cid":  "bafybeiazsz3twea2uxpen6452qqa7qnzp2xildfxliidhqk632jpvbixkm",
						"name": "age",
					},
					{
						"cid":  "bafybeidzukbs36cwwhab4rkpi6jfhhxse2vjtc5tf767qda5valcinilmy",
						"name": "name",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryLatestCommitsWithDocKeyWithSchemaVersionIdField(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple latest commits query with dockey and schema versiion id field",
		Request: `query {
					latestCommits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
						cid
						schemaVersionId
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"age": 21
				}`,
			},
		},
		Results: []map[string]any{
			{
				"cid":             "bafybeihbcl2ijavd6vdcj4vgunw4q5qt5itmumxw7iy7fhoqfsuvkpkqeq",
				"schemaVersionId": "bafkreictcre4pylafzzoh5lpgbetdodunz4r6pz3ormdzzpsz2lqtp4v34",
			},
		},
	}

	executeTestCase(t, test)
}
