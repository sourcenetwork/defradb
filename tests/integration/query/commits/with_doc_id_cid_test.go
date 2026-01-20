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

func TestQueryCommitsWithDocIDAndCidForDifferentDoc(t *testing.T) {
	test := testUtils.TestCase{
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
				Request: ` {
						_commits(
							docID: "bae-not-this-doc",
							cid: "bafybeica4js2abwqjjrz7dcialbortbz32uxp7ufxu7yljbwvmhjqqxzny"
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
				ExpectedError: "cid either does not exist or belong to document",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndCidForDifferentDocWithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: ` {
						_commits(
							docID: "bae-not-this-doc",
							cid: "bafyreia4x5ju33jenbimdqbtnuqc7pby4lydpa7efyk5iu4nl6urm6ofla"
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
				ExpectedError: "cid either does not exist or belong to document",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithDocIDAndCidWithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: ` {
						_commits(
							docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
							cid: "bafyreia4x5ju33jenbimdqbtnuqc7pby4lydpa7efyk5iu4nl6urm6ofla"
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": "bafyreia4x5ju33jenbimdqbtnuqc7pby4lydpa7efyk5iu4nl6urm6ofla",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndCidWithUpdateAndDepth(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			// depth is pretty arbitrary here, as long as its big enough to cover the updates
			// from the target cid (ie >=2)
			testUtils.Request{
				Request: ` {
						_commits(
							docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
							cid: "bafyreia4x5ju33jenbimdqbtnuqc7pby4lydpa7efyk5iu4nl6urm6ofla",
							depth: 5
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": "bafyreia4x5ju33jenbimdqbtnuqc7pby4lydpa7efyk5iu4nl6urm6ofla",
						},
						{
							"cid": "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
