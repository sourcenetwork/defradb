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
						"name":	"John",
						"age":	21
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
						"cid": "bafybeiaayywpuodsyqnp6vbnftzpzoke4iwofeyuta3wyfv6xpjenf7le4",
					},
					{
						"cid": "bafybeiam53sbmu6zqvullm53sq2hlrdc4icifbcznmltg5olijyrdnjhtm",
					},
					{
						"cid": "bafybeib5dcp5ccsm3r3sey6td23r7qgd2xjqnx4gfxss6hd77zhjl44gbq",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsMultipleDocs(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, multiple docs",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Shahzad",
						"age":	28
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
						"cid": "bafybeieeqyzf55hciob763jydnx57vxd76gk2jtrsdk5ebgt5wxsbcywba",
					},
					{
						"cid": "bafybeigbnfgjnek5bp4ju7yj2qbalfmjkvwlndccjltrbrbaxhvpcif2dy",
					},
					{
						"cid": "bafybeifmpp7azrslaowu5vissk5rellfoq6vbjplyd6ovrdsz6vwe5tm6e",
					},
					{
						"cid": "bafybeiaayywpuodsyqnp6vbnftzpzoke4iwofeyuta3wyfv6xpjenf7le4",
					},
					{
						"cid": "bafybeiam53sbmu6zqvullm53sq2hlrdc4icifbcznmltg5olijyrdnjhtm",
					},
					{
						"cid": "bafybeib5dcp5ccsm3r3sey6td23r7qgd2xjqnx4gfxss6hd77zhjl44gbq",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithSchemaVersionIdField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple commits query yielding schemaVersionId",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
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
						"cid":             "bafybeiaayywpuodsyqnp6vbnftzpzoke4iwofeyuta3wyfv6xpjenf7le4",
						"schemaVersionId": "bafkreiayhdsgzhmrz6t5d3x2cgqqbdjt7aqgldtlkmxn5eibg542j3n6ea",
					},
					{
						"cid":             "bafybeiam53sbmu6zqvullm53sq2hlrdc4icifbcznmltg5olijyrdnjhtm",
						"schemaVersionId": "bafkreiayhdsgzhmrz6t5d3x2cgqqbdjt7aqgldtlkmxn5eibg542j3n6ea",
					},
					{
						"cid":             "bafybeib5dcp5ccsm3r3sey6td23r7qgd2xjqnx4gfxss6hd77zhjl44gbq",
						"schemaVersionId": "bafkreiayhdsgzhmrz6t5d3x2cgqqbdjt7aqgldtlkmxn5eibg542j3n6ea",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithFieldNameField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple commits query yielding fieldName",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							fieldName
						}
					}
				`,
				Results: []map[string]any{
					{
						"fieldName": "age",
					},
					{
						"fieldName": "name",
					},
					{
						"fieldName": nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithFieldNameFieldAndUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple commits query yielding fieldName",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							fieldName
						}
					}
				`,
				Results: []map[string]any{
					{
						"fieldName": "age",
					},
					{
						"fieldName": "age",
					},
					{
						"fieldName": "name",
					},
					{
						"fieldName": nil,
					},
					{
						"fieldName": nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithFieldIDField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple commits query yielding fieldId",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							fieldId
						}
					}
				`,
				Results: []map[string]any{
					{
						"fieldId": "1",
					},
					{
						"fieldId": "2",
					},
					{
						"fieldId": "C",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithFieldIDFieldWithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple commits query yielding fieldId",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: `
					query {


						commits {
							fieldId
						}
					}
				`,
				Results: []map[string]any{
					{
						"fieldId": "1",
					},
					{
						"fieldId": "1",
					},
					{
						"fieldId": "2",
					},
					{
						"fieldId": "C",
					},
					{
						"fieldId": "C",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
