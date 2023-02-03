// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package commits

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommits(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query",
		Request: `query {
					commits {
						cid
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
				"cid": "bafybeigju7dgicfq3fxvtlxtjao7won4xc7kusykkvumngjfx5i2c7ibny",
			},
			{
				"cid": "bafybeiaqarrcayyoly2gdiam6mhh72ls4azwa7brozxxc3q2srnggkkqkq",
			},
			{
				"cid": "bafybeid5l577igkgcn6wjqjeqxlta4dcc3a3iykwkborf4fklaenjuctoq",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsMultipleDocs(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple all commits query, multiple docs",
		Request: `query {
					commits {
						cid
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
				`{
					"Name": "Shahzad",
					"Age": 28
				}`,
			},
		},
		Results: []map[string]any{
			{
				"cid": "bafybeibmprk2bxsv2nj2sf5ofmu7yuqe7dz2dze546nxkzwwylxyzpruoy",
			},
			{
				"cid": "bafybeidbb4dv2smuzmeodcrbt2dk6loqj7i3a6fofl32ejbx2gtinxguye",
			},
			{
				"cid": "bafybeifmifbksnwuwxhkwjqdojbddw2274f7wzd4jamllaoud3llunm5xu",
			},
			{
				"cid": "bafybeigju7dgicfq3fxvtlxtjao7won4xc7kusykkvumngjfx5i2c7ibny",
			},
			{
				"cid": "bafybeiaqarrcayyoly2gdiam6mhh72ls4azwa7brozxxc3q2srnggkkqkq",
			},
			{
				"cid": "bafybeid5l577igkgcn6wjqjeqxlta4dcc3a3iykwkborf4fklaenjuctoq",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryCommitsWithSchemaVersionIdField(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple commits query yielding schemaVersionId",
		Request: `query {
					commits {
						cid
						schemaVersionId
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
				"cid":             "bafybeigju7dgicfq3fxvtlxtjao7won4xc7kusykkvumngjfx5i2c7ibny",
				"schemaVersionId": "bafkreihaqmvbjvm2q4iwkjnuafavvsakiaztlqnridiybxystfm27uwlde",
			},
			{
				"cid":             "bafybeiaqarrcayyoly2gdiam6mhh72ls4azwa7brozxxc3q2srnggkkqkq",
				"schemaVersionId": "bafkreihaqmvbjvm2q4iwkjnuafavvsakiaztlqnridiybxystfm27uwlde",
			},
			{
				"cid":             "bafybeid5l577igkgcn6wjqjeqxlta4dcc3a3iykwkborf4fklaenjuctoq",
				"schemaVersionId": "bafkreihaqmvbjvm2q4iwkjnuafavvsakiaztlqnridiybxystfm27uwlde",
			},
		},
	}

	executeTestCase(t, test)
}
