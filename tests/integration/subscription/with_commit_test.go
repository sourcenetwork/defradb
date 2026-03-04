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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCommitSubscription_WithAddMutations_ReturnCommits(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.SubscriptionRequest{
				Request: `subscription {
					_commits {
						cid
					}
				}`,
				Results: []map[string]any{
					{
						"_commits": []map[string]any{
							{
								"cid": "bafyreid3ymo4wt3gdubzo2n247qqecsbazjaujprvuv62rc3rne5fx765m",
							},
						},
					},
					{
						"_commits": []map[string]any{
							{
								"cid": "bafyreib5dvg3wkm722kietpvx5gmfueilyvywyiz2tl44q6xnhv4bedcpq",
							},
						},
					},
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 27,
					"points": 42.1,
					"verified": true
				}`,
			},
			&action.AddDoc{
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

func TestCommitSubscription_WithCommitLinksAddMutations_ValidLinks(t *testing.T) {
	create1Links := testUtils.NewSameValue()
	create2Links := testUtils.NewSameValue()
	create1Heads := testUtils.NewSameValue()
	create2Heads := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			&action.SubscriptionRequest{
				Request: `subscription {
					_commits {
						cid
						links {
							cid
							fieldName
						}
						heads {
							cid
						}
					}
				}`,
				Results: []map[string]any{
					{
						"_commits": []map[string]any{
							{
								"cid":   "bafyreid3ymo4wt3gdubzo2n247qqecsbazjaujprvuv62rc3rne5fx765m",
								"links": create1Links,
								"heads": create1Heads,
							},
						},
					},
					{
						"_commits": []map[string]any{
							{
								"cid":   "bafyreib5dvg3wkm722kietpvx5gmfueilyvywyiz2tl44q6xnhv4bedcpq",
								"links": create2Links,
								"heads": create2Heads,
							},
						},
					},
				},
			},
			&action.Request{
				Request: `mutation {
					add_User(input: {name: "John", age: 27, points: 42.1, verified: true}) {
						name
						_version {
							links {
								cid
								fieldName
							}
							heads {
								cid
							}
						}
					}
				}`,
				Results: map[string]any{
					"add_User": []map[string]any{
						{
							"name": "John",
							"_version": []map[string]any{
								{
									"links": create1Links,
									"heads": create1Heads,
								},
							},
						},
					},
				},
			},
			&action.Request{
				Request: `mutation {
					add_User(input: {name: "Addo", age: 31, points: 42.1, verified: true}) {
						name
						_version {
							links {
								cid
								fieldName
							}
							heads {
								cid
							}
						}
					}
				}`,
				Results: map[string]any{
					"add_User": []map[string]any{
						{
							"name": "Addo",
							"_version": []map[string]any{
								{
									"links": create2Links,
									"heads": create2Heads,
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
	addoUpdateCid := testUtils.NewSameValue()
	addoDocID := testUtils.NewSameValue()
	addoCreateCid := testUtils.NewSameValue()
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Addo",
						"age":	31,
						"points": 42.1,
						"verified": true
					}`,
			},
			// subscription filtered on addo doc
			&action.SubscriptionRequest{
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
								"cid":   addoUpdateCid,
								"docID": addoDocID,
							},
						},
					},
				},
			},
			// this mutation will be ignored in the subscription (john doc)
			&action.Request{
				Request: `mutation {
					update_User(docID: "bae-77e2140d-fee0-5f32-b63a-854c9d4311f9", input: {verified: true}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"update_User": []map[string]any{
						{
							"_docID": "bae-77e2140d-fee0-5f32-b63a-854c9d4311f9",
						},
					},
				},
			},
			// this mutation will be included in the subscription (addo doc)
			&action.Request{
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
							"_docID": addoDocID,
							"_version": []map[string]any{
								{
									"cid": addoUpdateCid,
								},
								{
									"cid": addoCreateCid,
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
