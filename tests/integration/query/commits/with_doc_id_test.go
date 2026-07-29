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

func TestQueryCommitsWithUnknownDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(docID: "unknown document ID") {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocID(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(docID: "{{.DocID0_0}}") {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{

						{
							"cid": uniqueCid,
						},
						{
							"cid": uniqueCid,
						},
						{
							"cid": uniqueCid,
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndLinks(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()
	ageCreateCid := testUtils.NewSameValue()
	nameCreateCid := testUtils.NewSameValue()
	createCompositeCid := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(docID: "{{.DocID0_0}}") {
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
				Results: map[string]any{
					"_commits": []map[string]any{
						{

							"cid":   gomega.And(ageCreateCid, uniqueCid),
							"links": []map[string]any{},
							"heads": []map[string]any{},
						},
						{

							"cid":   gomega.And(nameCreateCid, uniqueCid),
							"links": []map[string]any{},
							"heads": []map[string]any{},
						},
						{

							"cid": gomega.And(createCompositeCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid":       ageCreateCid,
									"fieldName": "age",
								},
								{
									"cid":       nameCreateCid,
									"fieldName": "name",
								},
							},
							"heads": []map[string]any{},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndUpdate(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			&action.Request{
				Request: `query {
						_commits(docID: "{{.DocID0_0}}") {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{

						{
							"cid":    uniqueCid,
							"height": int64(2),
						},
						{
							"cid":    uniqueCid,
							"height": int64(1),
						},
						{
							"cid":    uniqueCid,
							"height": int64(1),
						},
						{
							"cid":    uniqueCid,
							"height": int64(2),
						},
						{
							"cid":    uniqueCid,
							"height": int64(1),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (first results includes link._head, second
// includes link._Name).
func TestQueryCommitsWithDocIDAndUpdateAndLinks(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()
	ageCreateCid := testUtils.NewSameValue()
	ageUpdateCid := testUtils.NewSameValue()
	nameCreateCid := testUtils.NewSameValue()
	createCompositeCid := testUtils.NewSameValue()
	updateCompositeCid := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			&action.Request{
				Request: `query {
						_commits(docID: "{{.DocID0_0}}") {
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
				Results: map[string]any{
					"_commits": []map[string]any{
						{

							"cid":   gomega.And(ageUpdateCid, uniqueCid),
							"links": []map[string]any{},
							"heads": []map[string]any{
								{
									"cid": ageCreateCid,
								},
							},
						},
						{
							"cid":   gomega.And(ageCreateCid, uniqueCid),
							"links": []map[string]any{},
							"heads": []map[string]any{},
						},
						{

							"cid":   gomega.And(nameCreateCid, uniqueCid),
							"links": []map[string]any{},
							"heads": []map[string]any{},
						},
						{

							"cid": gomega.And(updateCompositeCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid":       ageUpdateCid,
									"fieldName": "age",
								},
							},
							"heads": []map[string]any{

								{
									"cid": createCompositeCid,
								},
							},
						},
						{
							"cid": gomega.And(createCompositeCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid":       ageCreateCid,
									"fieldName": "age",
								},
								{
									"cid":       nameCreateCid,
									"fieldName": "name",
								},
							},
							"heads": []map[string]any{},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_DocIDEmptyList(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Fred",
				},
			},
			&action.Request{
				Request: `query {
						_commits(docID: []) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_DocIDListOfOne(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Fred",
				},
			},
			&action.Request{
				Request: `query {
						_commits(docID: ["{{.DocID0_0}}"]) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": testUtils.NewUniqueValue(),
						},
						{
							"cid": testUtils.NewUniqueValue(),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_DocIDListOfMany(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Fred",
				},
			},
			&action.Request{
				Request: `query {
						_commits(docID: ["{{.DocID0_0}}", "{{.DocID0_1}}"]) {
							cid
						}
					}`,
				ExpectedError: "querying by multiple docIDs is not yet supported",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
