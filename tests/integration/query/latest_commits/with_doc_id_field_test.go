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
	test := testUtils.RequestTestCase{
		Description: "Simple latest commits query with docID and field name",
		Request: `query {
					latestCommits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", fieldId: "age") {
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
		Results: []map[string]any{},
	}

	executeTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (Users should not be specifying field ids).
func TestQueryLatestCommitsWithDocIDAndFieldId(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple latest commits query with docID and field id",
		Request: `query {
					latestCommits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", fieldId: "1") {
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
				"cid":   "bafybeihfw5lufgs7ygv45to5rqvt3xkecjgikoccjyx6y2i7lnaclmrcjm",
				"links": []map[string]any{},
			},
		},
	}

	executeTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (Users should not be specifying field ids).
func TestQueryLatestCommitsWithDocIDAndCompositeFieldId(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple latest commits query with docID and composite field id",
		Request: `query {
					latestCommits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", fieldId: "C") {
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
				"cid": "bafybeiedu23doqe2nagdbmkvfyuouajnfxo7ezy57vbv34dqewhwbfg45u",
				"links": []map[string]any{
					{
						"cid":  "bafybeihfw5lufgs7ygv45to5rqvt3xkecjgikoccjyx6y2i7lnaclmrcjm",
						"name": "age",
					},
					{
						"cid":  "bafybeigmez6gtszsqx6aevzlanvpazhhezw5va4wizhqtqz5k4s2dqjb24",
						"name": "name",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
