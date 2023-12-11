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
							cid: "bafybeiebx4gp6ghqksxq56imktbyefze6bezrd42tco2vqcv2hrx2jstam",
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
							cid: "bafybeiebx4gp6ghqksxq56imktbyefze6bezrd42tco2vqcv2hrx2jstam",
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
							cid: "bafybeiawyt67laj4wzigasmbwvglrdlpkba5xiagsseealuacrixitizoa"
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
							cid: "bafybeibg345u3ea7fq7hupjtz34w5tq3ffb4j6owygv6j33hfmgnj4tx64",
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
							cid: "bafybeiebx4gp6ghqksxq56imktbyefze6bezrd42tco2vqcv2hrx2jstam",
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
						"schemaVersionId": "bafkreidvd63bawkelxe3wtf7a65klkq4x3dvenqafyasndyal6fvffkeam",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithUpdateAndCidAndDocKeyWithPNCounterInt_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with second last cid and dockey with pncounter int type",
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
						cid: "bafybeiemzwrx7jotpqjzajfmu2gi2o7h35yptgehihbdyoxkmzox4coetm",
						dockey: "bae-a688789e-d8a6-57a7-be09-22e005ab79e0"
					) {
						name
						points
					}
				}`,
				Results: []map[string]any{
					{
						"name":   "John",
						"points": int64(5),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimple_WithUpdateAndCidAndDocKeyWithPNCounterFloat_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with second last cid and dockey with pncounter and float type",
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
						cid: "bafybeidanfj4emzh4izmij5bwocgnecpvolx2b77il4mtwyrucvmgueyp4",
						dockey: "bae-fa6a97e9-e0e9-5826-8a8c-57775d35e07c"
					) {
						name
						points
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John",
						// Note the lack of precision of float types.
						"points": 4.8999999999999995,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
