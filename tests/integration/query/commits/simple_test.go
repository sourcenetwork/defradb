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
						"cid": "bafybeic4x7hxoh7yhqmvo7c3mqoyv6j7lnnajkt2hzf2j3mjaf6wmwwl6u",
					},
					{
						"cid": "bafybeidd6rsya2q5gxaarx52da22ih5jdn5wgxsfehcuwquffgjvmdrh34",
					},
					{
						"cid": "bafybeiax37emgcmyjjsiae7kwqis675whyc73wth44amhcmsndfygfhl7m",
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
						"cid": "bafybeidvkbxozsr7qkhghljwth7g2sfipttevkyxbstqchoa26tqeopdjq",
					},
					{
						"cid": "bafybeifwd2xgjtokmauglml4jotv2sszmdux27ttiqtbc6vvfdzaabsw7y",
					},
					{
						"cid": "bafybeihbevhd4dcydunuhq4otjdpw4uyk2lfcqa5zfrrn7jyk3e4rcieom",
					},
					{
						"cid": "bafybeic4x7hxoh7yhqmvo7c3mqoyv6j7lnnajkt2hzf2j3mjaf6wmwwl6u",
					},
					{
						"cid": "bafybeidd6rsya2q5gxaarx52da22ih5jdn5wgxsfehcuwquffgjvmdrh34",
					},
					{
						"cid": "bafybeiax37emgcmyjjsiae7kwqis675whyc73wth44amhcmsndfygfhl7m",
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
						"cid":             "bafybeic4x7hxoh7yhqmvo7c3mqoyv6j7lnnajkt2hzf2j3mjaf6wmwwl6u",
						"schemaVersionId": "bafkreig2vbhp6udwzcn2grqjnr7b3qbzperexbdazyovgwvkotpq3ehmny",
					},
					{
						"cid":             "bafybeidd6rsya2q5gxaarx52da22ih5jdn5wgxsfehcuwquffgjvmdrh34",
						"schemaVersionId": "bafkreig2vbhp6udwzcn2grqjnr7b3qbzperexbdazyovgwvkotpq3ehmny",
					},
					{
						"cid":             "bafybeiax37emgcmyjjsiae7kwqis675whyc73wth44amhcmsndfygfhl7m",
						"schemaVersionId": "bafkreig2vbhp6udwzcn2grqjnr7b3qbzperexbdazyovgwvkotpq3ehmny",
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
