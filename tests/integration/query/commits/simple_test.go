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
						"cid": "bafybeifh5fpsnusu6tedflmcglsm4rxsc2k3h35iqsbrcevfish6xrxftq",
					},
					{
						"cid": "bafybeid32koipavfiznapnb7cbhbwqfi7vmbob6xj6tzj2w66utir4qv5y",
					},
					{
						"cid": "bafybeickezedvwqujfvdih25kcwpkxtekdy3ykaofxejtxmhqybw2pfrnq",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
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
						"cid": "bafybeiculvxijb7chmmmb2cad3k4e2zfug5tpehp4gqwayarh4yzlnft4m",
					},
					{
						"cid": "bafybeiafn54vuzrmvqy2pi7wpqafb4rz4komweoqekx3ayrqjykitdfh4e",
					},
					{
						"cid": "bafybeicjgjnffau6xe36p4xdauuzujpurvi2ueizhjvxleq5pjnkyp7bfq",
					},
					{
						"cid": "bafybeifh5fpsnusu6tedflmcglsm4rxsc2k3h35iqsbrcevfish6xrxftq",
					},
					{
						"cid": "bafybeid32koipavfiznapnb7cbhbwqfi7vmbob6xj6tzj2w66utir4qv5y",
					},
					{
						"cid": "bafybeickezedvwqujfvdih25kcwpkxtekdy3ykaofxejtxmhqybw2pfrnq",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
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
						"cid":             "bafybeifh5fpsnusu6tedflmcglsm4rxsc2k3h35iqsbrcevfish6xrxftq",
						"schemaVersionId": "bafkreiaub6dogmiftv3h5jlsn6wyifpy2ur44shayu6ahkbrxurovtawlm",
					},
					{
						"cid":             "bafybeid32koipavfiznapnb7cbhbwqfi7vmbob6xj6tzj2w66utir4qv5y",
						"schemaVersionId": "bafkreiaub6dogmiftv3h5jlsn6wyifpy2ur44shayu6ahkbrxurovtawlm",
					},
					{
						"cid":             "bafybeickezedvwqujfvdih25kcwpkxtekdy3ykaofxejtxmhqybw2pfrnq",
						"schemaVersionId": "bafkreiaub6dogmiftv3h5jlsn6wyifpy2ur44shayu6ahkbrxurovtawlm",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
