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

func TestCommitSubscription_WithCreateMutations_ReturnCommits(t *testing.T) {
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
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 27,
					"points": 42.1,
					"verified": true
				}`,
			},
			&action.CreateDoc{
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
					create_User(input: {name: "John", age: 27, points: 42.1, verified: true}) {
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
					"create_User": []map[string]any{
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
					create_User(input: {name: "Addo", age: 31, points: 42.1, verified: true}) {
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
					"create_User": []map[string]any{
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
	updateCid := testUtils.NewSameValue()

	docID := "bae-45e90427-d499-598b-902a-6a3c65d0b504"
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
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
								"cid":   updateCid,
								"docID": docID,
							},
						},
					},
				},
			},
			// this mutation must be ignored by the subscription
			&action.Request{
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
