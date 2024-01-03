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
		 					cid: "bafybeiealfslrqsbiwotlducidmesjaemiq2hb7y2bxkcwc7bppuceujui",
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
		 					cid: "bafybeiealfslrqsbiwotlducidmesjaemiq2hb7y2bxkcwc7bppuceujui",
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
							cid: "bafybeibnj6yitgmynodaxnvtl22rhzclhsrc5asmocwyccsbsamobibpsy",
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
							cid: "bafybeify36bauenmsov4rijdmency367boy234mjezpvg4dj6r47ay3jwq",
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
		 					cid: "bafybeiealfslrqsbiwotlducidmesjaemiq2hb7y2bxkcwc7bppuceujui",
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
