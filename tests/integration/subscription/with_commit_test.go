// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

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
								"cid": "{{.CID0_0_0}}",
							},
						},
					},
					{
						"_commits": []map[string]any{
							{
								"cid": "{{.CID0_1_0}}",
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
								"cid":   testUtils.ValidCID(),
								"links": create1Links,
								"heads": create1Heads,
							},
						},
					},
					{
						"_commits": []map[string]any{
							{
								"cid":   testUtils.ValidCID(),
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
					_commits(docID: "{{.DocID0_1}}") {
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
					update_User(docID: "{{.DocID0_0}}", input: {verified: true}) {
						_docID
					}
				}`,
				Results: map[string]any{
					"update_User": []map[string]any{
						{
							"_docID": "{{.DocID0_0}}",
						},
					},
				},
			},
			// this mutation will be included in the subscription (addo doc)
			&action.Request{
				Request: `mutation {
					update_User(docID: "{{.DocID0_1}}", input: {verified: false}) {
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
