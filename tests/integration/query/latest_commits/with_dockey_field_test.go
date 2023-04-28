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
					latestCommits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", field: "age") {
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
func TestQueryLatestCommitsWithDocKeyAndFieldId(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple latest commits query with dockey and field id",
		Request: `query {
					latestCommits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", field: "1") {
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
				"cid":   "bafybeid4q6fhbbchwife54qqumb2rof6lui7d5njbkylkradmewqdibhjm",
				"links": []map[string]any{},
			},
		},
	}

	executeTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (Users should not be specifying field ids).
func TestQueryLatestCommitsWithDocKeyAndCompositeFieldId(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple latest commits query with dockey and composite field id",
		Request: `query {
					latestCommits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", field: "C") {
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
				"cid": "bafybeic267ibnl45al5ekxpqorsbwv2xghsuxm4dpdi47ojhl7yuvdonuy",
				"links": []map[string]any{
					{
						"cid":  "bafybeid4q6fhbbchwife54qqumb2rof6lui7d5njbkylkradmewqdibhjm",
						"name": "age",
					},
					{
						"cid":  "bafybeid435xjpnucmhshryyg3bfzf7be7hotq4m2kfw77yn7utd5yyimiq",
						"name": "name",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
