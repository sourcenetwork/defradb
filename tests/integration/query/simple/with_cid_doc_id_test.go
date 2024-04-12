// Copyright 2022 Democratized Data Foundation
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

func TestQuerySimpleWithInvalidCidAndInvalidDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with invalid cid and invalid docID",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.Request{
				Request: `query {
					Users (
							cid: "any non-nil string value - this will be ignored",
							docID: "invalid docID"
						) {
						name
					}
				}`,
				ExpectedError: "invalid cid: selected encoding not supported",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (should just return empty).
func TestQuerySimpleWithUnknownCidAndInvalidDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with unknown cid and invalid docID",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.Request{
				Request: `query {
					Users (
							cid: "bafybeid57gpbwi4i6bg7g357vwwyzsmr4bjo22rmhoxrwqvdxlqxcgaqvu",
							docID: "invalid docID"
						) {
						name
					}
				}`,
				ExpectedError: "failed to get block in blockstore: ipld: could not find",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithCidAndDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with cid and docID",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
							cid: "bafybeif757a4mdwimqwl24ujjnao6xlajiajz2hwuleopnptusuttri6zu",
							docID: "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"
						) {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithUpdateAndFirstCidAndDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with (first) cid and docID",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name": "Johnn"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
							cid: "bafybeif757a4mdwimqwl24ujjnao6xlajiajz2hwuleopnptusuttri6zu",
							docID: "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"
						) {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithUpdateAndLastCidAndDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with (last) cid and docID",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name": "Johnn"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
							cid: "bafybeibwxvtvppws6sjfoajazevrdh27g4qwn5wguslpabyl3kzxd2a6fm",
							docID: "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"
						) {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Johnn",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithUpdateAndMiddleCidAndDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with (middle) cid and docID",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name": "Johnn"
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name": "Johnnn"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
							cid: "bafybeibwxvtvppws6sjfoajazevrdh27g4qwn5wguslpabyl3kzxd2a6fm",
							docID: "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"
						) {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Johnn",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithUpdateAndFirstCidAndDocIDAndSchemaVersion(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with (first) cid and docID and yielded schema version",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name": "Johnn"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
							cid: "bafybeif757a4mdwimqwl24ujjnao6xlajiajz2hwuleopnptusuttri6zu",
							docID: "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad"
						) {
						name
						_version {
							schemaVersionId
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John",
						"_version": []map[string]any{
							{
								"schemaVersionId": "bafkreia3o3cetvcnnxyu5spucimoos77ifungfmacxdkva4zah2is3aooe",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Note: Only the first CID is reproducible given the added entropy to the Counter CRDT type.
func TestCidAndDocIDQuery_ContainsPNCounterWithIntKind_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with first cid and docID with pncounter int type",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Int @crdt(type: "pncounter")
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"points": 10
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"points": -5
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"points": 20
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
						cid: "bafybeicruxxfhxhyvefbxid7gukdbnfzkyad45phu4mnwzzqde24p32xnu",
						docID: "bae-a688789e-d8a6-57a7-be09-22e005ab79e0"
					) {
						name
						points
					}
				}`,
				Results: []map[string]any{
					{
						"name":   "John",
						"points": int64(10),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Note: Only the first CID is reproducible given the added entropy to the Counter CRDT type.
func TestCidAndDocIDQuery_ContainsPNCounterWithFloatKind_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with first cid and docID with pncounter and float type",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Float @crdt(type: "pncounter")
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"points": 10.2
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"points": -5.3
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"points": 20.6
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
						cid: "bafybeibeo7pmvzpkkanwd72q4qu3m4yxex3coufq7uogvcnjwgqzrlpco4",
						docID: "bae-fa6a97e9-e0e9-5826-8a8c-57775d35e07c"
					) {
						name
						points
					}
				}`,
				Results: []map[string]any{
					{
						"name":   "John",
						"points": 10.2,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Note: Only the first CID is reproducible given the added entropy to the Counter CRDT type.
func TestCidAndDocIDQuery_ContainsPCounterWithIntKind_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with first cid and docID with pcounter int type",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Int @crdt(type: "pcounter")
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"points": 10
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"points": 20
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
						cid: "bafybeidtjhrssohan2f5nt7ml3nh4bovpaqhqjvijlpacfednyx77iw5y4",
						docID: "bae-a688789e-d8a6-57a7-be09-22e005ab79e0"
					) {
						name
						points
					}
				}`,
				Results: []map[string]any{
					{
						"name":   "John",
						"points": int64(10),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Note: Only the first CID is reproducible given the added entropy to the Counter CRDT type.
func TestCidAndDocIDQuery_ContainsPCounterWithFloatKind_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with first cid and docID with pcounter and float type",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Float @crdt(type: "pcounter")
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"points": 10.2
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"points": 20.6
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
						cid: "bafybeieimeijjl4hdvqkt5gkn62j54nlnaetm4te7w4z2mdljlyphfsyji",
						docID: "bae-fa6a97e9-e0e9-5826-8a8c-57775d35e07c"
					) {
						name
						points
					}
				}`,
				Results: []map[string]any{
					{
						"name":   "John",
						"points": 10.2,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
