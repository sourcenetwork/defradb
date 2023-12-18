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
		Description: "Simple latest commits query with docID",
		Request: `query {
					latestCommits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
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
				"cid": "bafybeib5dcp5ccsm3r3sey6td23r7qgd2xjqnx4gfxss6hd77zhjl44gbq",
				"links": []map[string]any{
					{
						"cid":  "bafybeiaayywpuodsyqnp6vbnftzpzoke4iwofeyuta3wyfv6xpjenf7le4",
						"name": "age",
					},
					{
						"cid":  "bafybeiam53sbmu6zqvullm53sq2hlrdc4icifbcznmltg5olijyrdnjhtm",
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
		Description: "Simple latest commits query with docID and schema versiion id field",
		Request: `query {
					latestCommits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
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
				"cid":             "bafybeib5dcp5ccsm3r3sey6td23r7qgd2xjqnx4gfxss6hd77zhjl44gbq",
				"schemaVersionId": "bafkreiayhdsgzhmrz6t5d3x2cgqqbdjt7aqgldtlkmxn5eibg542j3n6ea",
			},
		},
	}

	executeTestCase(t, test)
}
