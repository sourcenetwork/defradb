// Copyright 2022 Democratized Data Foundation
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
	test := testUtils.TestCase{
		Description: "Simple all commits query",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						commits {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
					},
					{
						"cid": "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
					},
					{
						"cid": "bafybeicjgxd46coc73ybx7cvdxk5c6pdbp5nbick4drrlbevgg7a7mhkzy",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsMultipleDocs(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, multiple docs",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"Shahzad",
						"Age":	28
					}`,
			},
			testUtils.Request{
				Request: `query {
						commits {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeibkvya7a3qg2x5rqx5qz3r6vqbfddppwqkmxkygwpgzavjtr2dbsm",
					},
					{
						"cid": "bafybeicb2o2enown46mt4rlzshoswuuxddf2anbc74eeydthrpyoloyziu",
					},
					{
						"cid": "bafybeic52mkubp66yur3zsohpk4ccwpapx4snca5jvnrmlzhefpnn7x4am",
					},
					{
						"cid": "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
					},
					{
						"cid": "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
					},
					{
						"cid": "bafybeicjgxd46coc73ybx7cvdxk5c6pdbp5nbick4drrlbevgg7a7mhkzy",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithSchemaVersionIdField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple commits query yielding schemaVersionId",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						commits {
							cid
							schemaVersionId
						}
					}`,
				Results: []map[string]any{
					{
						"cid":             "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
						"schemaVersionId": "bafkreibwyhaiseplil6tayn7spazp3qmc7nkoxdjb7uoe5zvcac4pgbwhy",
					},
					{
						"cid":             "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
						"schemaVersionId": "bafkreibwyhaiseplil6tayn7spazp3qmc7nkoxdjb7uoe5zvcac4pgbwhy",
					},
					{
						"cid":             "bafybeicjgxd46coc73ybx7cvdxk5c6pdbp5nbick4drrlbevgg7a7mhkzy",
						"schemaVersionId": "bafkreibwyhaiseplil6tayn7spazp3qmc7nkoxdjb7uoe5zvcac4pgbwhy",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}
