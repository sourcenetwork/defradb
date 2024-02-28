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
	test := testUtils.RequestTestCase{
		Description: "Simple query with invalid cid and invalid docID",
		Request: `query {
					Users (
							cid: "any non-nil string value - this will be ignored",
							docID: "invalid docID"
						) {
						Name
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
		ExpectedError: "invalid cid: selected encoding not supported",
	}

	executeTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (should just return empty).
func TestQuerySimpleWithUnknownCidAndInvalidDocID(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with unknown cid and invalid docID",
		Request: `query {
					Users (
							cid: "bafybeid57gpbwi4i6bg7g357vwwyzsmr4bjo22rmhoxrwqvdxlqxcgaqvu",
							docID: "invalid docID"
						) {
						Name
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
		ExpectedError: "failed to get block in blockstore: ipld: could not find",
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithCidAndDocID(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with cid and docID",
		Request: `query {
					Users (
							cid: "bafybeicojqe66grk564b2hns3zi6rhquqvugxj6wi4s6xk4e2gg65dzx5e",
							docID: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"
						) {
						Name
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
				"Name": "John",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithUpdateAndFirstCidAndDocID(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with (first) cid and docID",
		Request: `query {
					Users (
							cid: "bafybeicojqe66grk564b2hns3zi6rhquqvugxj6wi4s6xk4e2gg65dzx5e",
							docID: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"
						) {
						Name
						Age
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
		Updates: map[int]map[int][]string{
			0: {
				0: {
					// update to change age to 22 on document 0
					`{"Age": 22}`,
					// then update it again to change age to 23 on document 0
					`{"Age": 23}`,
				},
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
				"Age":  int64(21),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithUpdateAndLastCidAndDocID(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with (last) cid and docID",
		Request: `query {
					Users (
							cid: "bafybeifxz2k3qudz2fau37xu3unw5l4ihenha66tlb37gctq5mtdriq3ly"
							docID: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"
						) {
						Name
						Age
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
		Updates: map[int]map[int][]string{
			0: {
				0: {
					// update to change age to 22 on document 0
					`{"Age": 22}`,
					// then update it again to change age to 23 on document 0
					`{"Age": 23}`,
				},
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
				"Age":  int64(23),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithUpdateAndMiddleCidAndDocID(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with (middle) cid and docID",
		Request: `query {
					Users (
							cid: "bafybeigcjabzlkuj4j35boczgcl4jmars7gz5a7dfvpq3m344bzth7ebqq",
							docID: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"
						) {
						Name
						Age
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
		Updates: map[int]map[int][]string{
			0: {
				0: {
					// update to change age to 22 on document 0
					`{"Age": 22}`,
					// then update it again to change age to 23 on document 0
					`{"Age": 23}`,
				},
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
				"Age":  int64(22),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithUpdateAndFirstCidAndDocIDAndSchemaVersion(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with (first) cid and docID and yielded schema version",
		Request: `query {
					Users (					
							cid: "bafybeicojqe66grk564b2hns3zi6rhquqvugxj6wi4s6xk4e2gg65dzx5e",
							docID: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"
						) {
						Name
						Age
						_version {
							schemaVersionId
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
		Updates: map[int]map[int][]string{
			0: {
				0: {
					// update to change age to 22 on document 0
					`{"Age": 22}`,
					// then update it again to change age to 23 on document 0
					`{"Age": 23}`,
				},
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
				"Age":  int64(21),
				"_version": []map[string]any{
					{
						"schemaVersionId": "bafkreiekkppcdl573ru624wh3kwkmy2nhqzjsvqpu6jv5dgq2kidpnon4u",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
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
						cid: "bafybeiebqzqml6nn3laarr7yekakrsdnkn4nbgrl4xc5rshljp3in6au2m",
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
						cid: "bafybeifzuh74aq47vjngkwipjne4r2gi3v2clewgsruspqirihnps4vcmu",
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
