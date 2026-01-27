// Copyright 2025 Democratized Data Foundation
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

	"github.com/onsi/gomega"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommits_WithSingleCreateNestedLinks_Succeed(t *testing.T) {
	ageCreateCid := testUtils.NewSameValue()
	nameCreateCid := testUtils.NewSameValue()
	createCompositeCid := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `
					query {
						_commits {
							cid
							height
							fieldName
							links {
								cid
								height
								fieldName
							}
							heads {
								cid
								height
							}	
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":       ageCreateCid,
							"height":    uint64(1),
							"fieldName": "age",
							"links":     []map[string]any{},
							"heads":     []map[string]any{},
						},
						{
							"cid":       nameCreateCid,
							"height":    uint64(1),
							"fieldName": "name",
							"links":     []map[string]any{},
							"heads":     []map[string]any{},
						},
						{
							"cid":       createCompositeCid,
							"height":    uint64(1),
							"fieldName": "_C",
							"links": []map[string]any{
								{
									"cid":       ageCreateCid,
									"height":    uint64(1),
									"fieldName": "age",
								},
								{
									"cid":       nameCreateCid,
									"height":    uint64(1),
									"fieldName": "name",
								},
							},
							"heads": []map[string]any{},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithSingleCreateNestedLinksCompositeFilter_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `
					query {
						_commits(filter: {fieldName: {_eq: "_C"}}) {
							height
							fieldName
							links {
								height
								fieldName
							}
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"height":    uint64(1),
							"fieldName": "_C",
							"links": []map[string]any{
								{
									"height":    uint64(1),
									"fieldName": "age",
								},
								{
									"height":    uint64(1),
									"fieldName": "name",
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithSingleCreateNestedLinksNestedFilter_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `
					query {
						_commits(filter: {fieldName: {_eq: "_C"}}) {
							height
							fieldName
							links(filter: {fieldName: {_eq: "age"}}) {
								height
								fieldName
							}
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"height":    uint64(1),
							"fieldName": "_C",
							"links": []map[string]any{
								{
									"height":    uint64(1),
									"fieldName": "age",
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithSingleUpdateDoubleNestedLinks_Succeeds(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	ageUpdateCid := testUtils.NewSameValue()
	ageCreateCid := testUtils.NewSameValue()
	nameCreateCid := testUtils.NewSameValue()
	updateCompositeCid := testUtils.NewSameValue()
	createCompositeCid := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
						"age":	22
					}`,
			},
			&action.Request{
				Request: `
					query {
						_commits {
							cid
							delta
							docID
							fieldName
							height
							links {
								fieldName
								cid
								height
								docID
								heads {
									fieldName
									cid
								}
							}
							heads {
								cid
								height
								docID
								links {
									fieldName
									cid
								}
							}
							
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":       gomega.And(ageUpdateCid, uniqueCid),
							"delta":     testUtils.CBORValue(22),
							"docID":     "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
							"fieldName": "age",
							"height":    int64(2),
							"links":     []map[string]any{},
							"heads": []map[string]any{
								{
									"cid":    ageCreateCid,
									"height": int64(1),
									"docID":  "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
									"links":  []map[string]any{},
								},
							},
						},
						{
							"cid":       gomega.And(ageCreateCid, uniqueCid),
							"delta":     testUtils.CBORValue(21),
							"docID":     "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
							"fieldName": "age",
							"height":    int64(1),
							"links":     []map[string]any{},
							"heads":     []map[string]any{},
						},
						{
							"cid":       gomega.And(nameCreateCid, uniqueCid),
							"delta":     testUtils.CBORValue("John"),
							"docID":     "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
							"fieldName": "name",
							"height":    int64(1),
							"links":     []map[string]any{},
							"heads":     []map[string]any{},
						},
						{
							"cid":       gomega.And(updateCompositeCid, uniqueCid),
							"delta":     nil,
							"docID":     "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
							"fieldName": "_C",
							"height":    int64(2),
							"links": []map[string]any{
								{
									"cid":       ageUpdateCid,
									"fieldName": "age",
									"height":    int64(2),
									"docID":     "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
									"heads": []map[string]any{
										{
											"fieldName": "age",
											"cid":       ageCreateCid,
										},
									},
								},
							},
							"heads": []map[string]any{
								{
									"cid":    createCompositeCid,
									"height": int64(1),
									"docID":  "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
									"links": []map[string]any{
										{
											"fieldName": "age",
											"cid":       ageCreateCid,
										},
										{
											"fieldName": "name",
											"cid":       nameCreateCid,
										},
									},
								},
							},
						},
						{
							"cid":       gomega.And(createCompositeCid, uniqueCid),
							"delta":     nil,
							"docID":     "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
							"fieldName": "_C",
							"height":    int64(1),
							"links": []map[string]any{
								{
									"cid":       ageCreateCid,
									"fieldName": "age",
									"height":    uint64(1),
									"docID":     "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
									"heads":     []map[string]any{},
								},
								{
									"cid":       nameCreateCid,
									"fieldName": "name",
									"height":    uint64(1),
									"docID":     "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
									"heads":     []map[string]any{},
								},
							},
							"heads": []map[string]any{},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
