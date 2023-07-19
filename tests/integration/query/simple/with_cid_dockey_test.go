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

func TestQuerySimpleWithInvalidCidAndInvalidDocKey(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with invalid cid and invalid dockey",
		Request: `query {
					Users (
							cid: "any non-nil string value - this will be ignored",
							dockey: "invalid docKey"
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
func TestQuerySimpleWithUnknownCidAndInvalidDocKey(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with unknown cid and invalid dockey",
		Request: `query {
					Users (
							cid: "bafybeid57gpbwi4i6bg7g357vwwyzsmr4bjo22rmhoxrwqvdxlqxcgaqvu",
							dockey: "invalid docKey"
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

func TestQuerySimpleWithCidAndDocKey(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with cid and dockey",
		Request: `query {
					Users (
							cid: "bafybeicloiyf5zl5k54cjuhfg6rpsj7rnhxswvnuizpagd2kwq4px6aqn4",
							dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"
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

func TestQuerySimpleWithUpdateAndFirstCidAndDocKey(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with (first) cid and dockey",
		Request: `query {
					Users (
							cid: "bafybeicloiyf5zl5k54cjuhfg6rpsj7rnhxswvnuizpagd2kwq4px6aqn4",
							dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"
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
				"Age":  uint64(21),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithUpdateAndLastCidAndDocKey(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with (last) cid and dockey",
		Request: `query {
					Users (
							cid: "bafybeih4is4so2tw47gfjsty6nk7fvcggd3uyq5tgiw4yjnobtbdwnqdoi"
							dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"
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
				"Age":  uint64(23),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithUpdateAndMiddleCidAndDocKey(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with (middle) cid and dockey",
		Request: `query {
					Users (
							cid: "bafybeigvkbbe6e5ztwz7qtwu3xg2zj4stpiwkhyj6qbu6ir3qdmrd3bhem",
							dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"
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
				"Age":  uint64(22),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithUpdateAndFirstCidAndDocKeyAndSchemaVersion(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with (first) cid and dockey and yielded schema version",
		Request: `query {
					Users (
							cid: "bafybeicloiyf5zl5k54cjuhfg6rpsj7rnhxswvnuizpagd2kwq4px6aqn4",
							dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"
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
				"Age":  uint64(21),
				"_version": []map[string]any{
					{
						"schemaVersionId": "bafkreicl3pjcorfcaexxmqcrilkhx7xl37o6b34nxgtiauygtl7hrqbhoq",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
