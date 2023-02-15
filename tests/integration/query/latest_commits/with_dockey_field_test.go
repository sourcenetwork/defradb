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
func TestQueryLatestCommitsWithDocKeyAndFieldName(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple latest commits query with dockey and field name",
		Request: `query {
					latestCommits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", field: "Age") {
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
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Results: []map[string]any{},
	}

	executeTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (users should not be specifying field ids).
func TestQueryLatestCommitsWithDocKeyAndFieldId(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple latest commits query with dockey and field id",
		Request: `query {
					latestCommits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", field: "1") {
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
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Results: []map[string]any{
			{
				"cid":   "bafybeicovjpmtwu544e7hzgg7mcwabstmugesi3n62ju6kbimcsjqp23gu",
				"links": []map[string]any{},
			},
		},
	}

	executeTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (users should not be specifying field ids).
func TestQueryLatestCommitsWithDocKeyAndCompositeFieldId(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple latest commits query with dockey and composite field id",
		Request: `query {
					latestCommits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", field: "C") {
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
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Results: []map[string]any{
			{
				"cid": "bafybeignirnk6wuxtg2fzwfbvs26wmrldlhpqj243kwnxb4ewafbae23m4",
				"links": []map[string]any{
					{
						"cid":  "bafybeicovjpmtwu544e7hzgg7mcwabstmugesi3n62ju6kbimcsjqp23gu",
						"name": "Age",
					},
					{
						"cid":  "bafybeietvbhkavrhb6usprlsehh5cojgznzqv4zdah2bhbrmgc2ph3rxka",
						"name": "Name",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
