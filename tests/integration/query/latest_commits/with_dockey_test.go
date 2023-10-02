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
				"cid": "bafybeiax37emgcmyjjsiae7kwqis675whyc73wth44amhcmsndfygfhl7m",
				"links": []map[string]any{
					{
						"cid":  "bafybeic4x7hxoh7yhqmvo7c3mqoyv6j7lnnajkt2hzf2j3mjaf6wmwwl6u",
						"name": "age",
					},
					{
						"cid":  "bafybeidd6rsya2q5gxaarx52da22ih5jdn5wgxsfehcuwquffgjvmdrh34",
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
				"cid":             "bafybeiax37emgcmyjjsiae7kwqis675whyc73wth44amhcmsndfygfhl7m",
				"schemaVersionId": "bafkreig2vbhp6udwzcn2grqjnr7b3qbzperexbdazyovgwvkotpq3ehmny",
			},
		},
	}

	executeTestCase(t, test)
}
