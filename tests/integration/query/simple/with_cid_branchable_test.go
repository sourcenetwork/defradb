// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithCidOfBranchableCollection_FirstCid(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users @branchable {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name": "Freddddd"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
							cid: "bafyreiewwsnu2ld5qlntamdm77ayb7xtmxz3p5difvaaakaome7zbtpo4u"
						) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithCidOfBranchableCollection_MiddleCid(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users @branchable {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name": "Freddddd"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
							cid: "bafyreifpamlyhcbriztgbhds5ctgi5rm6w5wcar2py7246lo6j5v7iusxm"
						) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Freddddd",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithCidOfBranchableCollection_LastCid(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users @branchable {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name": "Freddddd"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
							cid: "bafyreigmt6ytph32jjxts2bij7fkne5ntionldsnklp35vcamvvl2x3a5i"
						) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Freddddd",
						},
						{
							"name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
