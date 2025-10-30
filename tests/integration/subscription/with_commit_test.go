// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package subscription

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCommitSubscription_WithCreateMutations_ReturnCommits(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SubscriptionRequest{
				Request: `subscription {
					_commits {
						cid
					}
				}`,
				Results: []map[string]any{
					{
						"_commits": []map[string]any{
							{
								"cid": "bafyreiaxbbq4vafq22ptdverb7v22eaubqb5luxul7eooble7nqlqgg5ii",
							},
						},
					},
					{
						"_commits": []map[string]any{
							{
								"cid": "bafyreialxrvwrz4rhgomch7kr7scx6t7m6xspbjecvzneirkgskh2tjele",
							},
						},
					},
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 27,
					"points": 42.1,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Addo",
					"age": 31,
					"points": 42.1,
					"verified": true
				}`,
			},
		},
	}

	execute(t, test)
}

func TestCommitSubscription_WithCommitLinksCreateMutations_ValidLinks(t *testing.T) {
	create1Links := testUtils.NewSameValue()
	create2Links := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SubscriptionRequest{
				Request: `subscription {
					_commits {
						cid
						links {
							cid
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"_commits": []map[string]any{
							{
								"cid":   "bafyreiaxbbq4vafq22ptdverb7v22eaubqb5luxul7eooble7nqlqgg5ii",
								"links": create1Links,
							},
						},
					},
					{
						"_commits": []map[string]any{
							{
								"cid":   "bafyreialxrvwrz4rhgomch7kr7scx6t7m6xspbjecvzneirkgskh2tjele",
								"links": create2Links,
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
					create_User(input: {name: "John", age: 27, points: 42.1, verified: true}) {
						name
						_version {
							links {
								cid
								name
							}
						}
					}
				}`,
				Results: map[string]any{
					"create_User": []map[string]any{
						{
							"name": "John",
							"_version": []map[string]any{
								{
									"links": create1Links,
								},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
					create_User(input: {name: "Addo", age: 31, points: 42.1, verified: true}) {
						name
						_version {
							links {
								cid
								name
							}
						}
					}
				}`,
				Results: map[string]any{
					"create_User": []map[string]any{
						{
							"name": "Addo",
							"_version": []map[string]any{
								{
									"links": create2Links,
								},
							},
						},
					},
				},
			},
		},
	}

	execute(t, test)
}

func TestCommitSubscription_WithDocFilterAndMultipleMutations_FilteredDoc(t *testing.T) {
	updateCid := testUtils.NewSameValue()

	docID := "bae-45e90427-d499-598b-902a-6a3c65d0b504"
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.SubscriptionRequest{
				Request: `subscription {
					_commits(docID: "bae-45e90427-d499-598b-902a-6a3c65d0b504") {
						cid		
						docID
					}
				}`,
				Results: []map[string]any{
					{
						"_commits": []map[string]any{
							{
								"cid":   updateCid,
								"docID": docID,
							},
						},
					},
				},
			},
			// this mutation must be ignored by the subscription
			testUtils.Request{
				Request: `mutation {
					create_User(input: {name: "Addo", age: 31, points: 42.1, verified: true}) {
						name
					}
				}`,
				Results: map[string]any{
					"create_User": []map[string]any{
						{
							"name": "Addo",
						},
					},
				},
			},
			// this mutation will be included in the subscription
			testUtils.Request{
				Request: `mutation {
					update_User(docID: "bae-45e90427-d499-598b-902a-6a3c65d0b504", input: {verified: false}) {
						_docID
						_version {
							cid
						}
					}
				}`,
				Results: map[string]any{
					"update_User": []map[string]any{
						{
							"_docID": docID,
							"_version": []map[string]any{
								{
									"cid": updateCid,
								},
							},
						},
					},
				},
			},
		},
	}

	execute(t, test)
}
