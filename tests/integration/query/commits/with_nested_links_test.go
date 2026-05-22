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

package commits

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func commitLink(fieldName string, height uint64) gomega.OmegaMatcher {
	return gomega.SatisfyAll(
		gomega.HaveKeyWithValue("cid", testUtils.ValidCID()),
		gomega.HaveKeyWithValue("fieldName", fieldName),
		gomega.HaveKeyWithValue("height", gomega.BeNumerically("==", height)),
	)
}

func commitHeadLink(fieldName string) gomega.OmegaMatcher {
	return gomega.SatisfyAll(
		gomega.HaveKeyWithValue("cid", testUtils.ValidCID()),
		gomega.HaveKeyWithValue("fieldName", fieldName),
	)
}

func TestQueryCommits_WithSingleAddNestedLinks_Succeed(t *testing.T) {
	ageCreateCid := testUtils.NewSameValue()
	nameCreateCid := testUtils.NewSameValue()
	createCompositeCid := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
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
							"links": gomega.ConsistOf(
								commitLink("age", 1),
								commitLink("name", 1),
							),
							"heads": []map[string]any{},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithSingleAddNestedLinksCompositeFilter_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
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
							"links": gomega.ConsistOf(
								gomega.SatisfyAll(
									gomega.HaveKeyWithValue("height", gomega.BeNumerically("==", 1)),
									gomega.HaveKeyWithValue("fieldName", "age"),
								),
								gomega.SatisfyAll(
									gomega.HaveKeyWithValue("height", gomega.BeNumerically("==", 1)),
									gomega.HaveKeyWithValue("fieldName", "name"),
								),
							),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithSingleAddNestedLinksNestedFilter_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
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
			&action.AddDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.UpdateDoc{
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
							"docID":     "{{.DocID0_0}}",
							"fieldName": "age",
							"height":    int64(2),
							"links":     []map[string]any{},
							"heads": []map[string]any{
								{
									"cid":    ageCreateCid,
									"height": int64(1),
									"docID":  "{{.DocID0_0}}",
									"links":  []map[string]any{},
								},
							},
						},
						{
							"cid":       gomega.And(ageCreateCid, uniqueCid),
							"delta":     testUtils.CBORValue(21),
							"docID":     "{{.DocID0_0}}",
							"fieldName": "age",
							"height":    int64(1),
							"links":     []map[string]any{},
							"heads":     []map[string]any{},
						},
						{
							"cid":       gomega.And(nameCreateCid, uniqueCid),
							"delta":     testUtils.CBORValue("John"),
							"docID":     "{{.DocID0_0}}",
							"fieldName": "name",
							"height":    int64(1),
							"links":     []map[string]any{},
							"heads":     []map[string]any{},
						},
						{
							"cid":       gomega.And(updateCompositeCid, uniqueCid),
							"delta":     nil,
							"docID":     "{{.DocID0_0}}",
							"fieldName": "_C",
							"height":    int64(2),
							"links": []map[string]any{
								{
									"cid":       ageUpdateCid,
									"fieldName": "age",
									"height":    int64(2),
									"docID":     "{{.DocID0_0}}",
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
									"docID":  "{{.DocID0_0}}",
									"links": gomega.ConsistOf(
										commitHeadLink("age"),
										commitHeadLink("name"),
									),
								},
							},
						},
						{
							"cid":       gomega.And(createCompositeCid, uniqueCid),
							"delta":     nil,
							"docID":     "{{.DocID0_0}}",
							"fieldName": "_C",
							"height":    int64(1),
							"links": gomega.ConsistOf(
								commitLink("age", 1),
								commitLink("name", 1),
							),
							"heads": []map[string]any{},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
