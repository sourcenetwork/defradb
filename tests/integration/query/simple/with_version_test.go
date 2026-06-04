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

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithNestedLatestCommit(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						Name
						Age
						_version {
							cid
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
					"Users": []map[string]any{
						{
							"Name": "John",
							"Age":  int64(21),
							"_version": []map[string]any{
								{
									"cid": "bafyreidzdrrjkjch3icknknh6tfmwitc352v54bds4h5ftblghwhsojgru",
									"links": []map[string]any{
										{
											"cid":       "bafyreiapiprlzrtsh7bkf4ru36q7g5nm2nz5unke2kkl4uyofd3scxmlye",
											"fieldName": "Name",
										},
										{
											"cid":       "bafyreie7qecbfvigblfgobuyt7hrejf7k2isscppzr75hwfkm5brntauva",
											"fieldName": "Age",
										},
									},
									"heads": []map[string]any{},
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuery_AddDocWithNestedLatestCommit(t *testing.T) {
	docCompositeCid := testUtils.NewUniqueValue()
	ageCreateCid := testUtils.NewUniqueValue()
	nameCreateCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						Name
						Age
						_version {
							cid
							links {
								cid
								fieldName
								height
							}
							heads {
								cid
								height
							}
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"Age":  int64(21),
							"_version": []map[string]any{
								{
									"cid": docCompositeCid,
									"links": []map[string]any{
										{
											"cid":       nameCreateCid,
											"fieldName": "Name",
											"height":    uint64(1),
										},
										{
											"cid":       ageCreateCid,
											"fieldName": "Age",
											"height":    uint64(1),
										},
									},
									"heads": []map[string]any{},
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuery_UpdateDocWithNestedLatestCommit(t *testing.T) {
	docUpdateCompositeCid := testUtils.NewUniqueValue()
	docCreateCompositeCid := testUtils.NewSameValue()
	ageUpdateCid := testUtils.NewUniqueValue()
	ageCreateCid := testUtils.NewUniqueValue()
	nameCreateCid := testUtils.NewUniqueValue()
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"Age": 22
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						Name
						Age
						_version {
							cid
							links {
								cid
								fieldName
								height
							}
							heads {
								cid
								height
							}
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"Age":  int64(22),
							"_version": []map[string]any{
								{
									"cid": docUpdateCompositeCid,
									"links": []map[string]any{
										{
											"cid":       ageUpdateCid,
											"fieldName": "Age",
											"height":    uint64(2),
										},
									},
									"heads": []map[string]any{
										{
											"cid":    docCreateCompositeCid,
											"height": uint64(1),
										},
									},
								},
								{
									"cid": docCreateCompositeCid,
									"links": []map[string]any{
										{
											"cid":       nameCreateCid,
											"fieldName": "Name",
											"height":    uint64(1),
										},
										{
											"cid":       ageCreateCid,
											"fieldName": "Age",
											"height":    uint64(1),
										},
									},
									"heads": []map[string]any{},
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithEmbeddedLatestCommitWithCollectionVersionID(t *testing.T) {
	collectionVersionID := testUtils.NewUniqueValue()
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						Name
						_version {
							collectionVersionId
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_version": []map[string]any{
								{
									"collectionVersionId": collectionVersionID,
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithEmbeddedLatestCommitWithDocID(t *testing.T) {
	docID := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						Name
						_docID
						_version {
							docID
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":   "John",
							"_docID": docID,
							"_version": []map[string]any{
								{
									"docID": docID,
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithMultipleAliasedEmbeddedLatestCommit(t *testing.T) {
	docCreateCompositeCid := testUtils.NewUniqueValue()
	ageCreateCid := testUtils.NewUniqueValue()
	nameCreateCid := testUtils.NewUniqueValue()
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						Name
						Age
						_version {
							cid
							L1: links {
								cid
								fieldName
							}
							L2: links {
								fieldName
							}
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"Age":  int64(21),
							"_version": []map[string]any{
								{
									"cid": docCreateCompositeCid,
									"L1": []map[string]any{
										{
											"cid":       nameCreateCid,
											"fieldName": "Name",
										},
										{
											"cid":       ageCreateCid,
											"fieldName": "Age",
										},
									},
									"L2": []map[string]any{
										{
											"fieldName": "Name",
										},
										{
											"fieldName": "Age",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithMultipleAliasedInterleavedNestedLatestCommit(t *testing.T) {
	docUpdateCompositeCid := testUtils.NewUniqueValue()
	docCreateCompositeCid := testUtils.NewSameValue()
	ageUpdateCid := testUtils.NewUniqueValue()
	ageCreateCid := testUtils.NewUniqueValue()
	nameCreateCid := testUtils.NewUniqueValue()
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"Age": 22
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						Name
						Age
						_version {
							cid
							L1: links {
								cid
								fieldName
							}
							H1: heads {
								cid
							}
							L2: links {
								fieldName
							}
							H2: heads {
								cid
								height
							}
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"Age":  int64(22),
							"_version": []map[string]any{
								{
									"cid": docUpdateCompositeCid,
									"L1": []map[string]any{
										{
											"cid":       ageUpdateCid,
											"fieldName": "Age",
										},
									},
									"H1": []map[string]any{
										{
											"cid": docCreateCompositeCid,
										},
									},
									"L2": []map[string]any{
										{
											"fieldName": "Age",
										},
									},
									"H2": []map[string]any{
										{
											"cid":    docCreateCompositeCid,
											"height": uint64(1),
										},
									},
								},
								{
									"cid": docCreateCompositeCid,
									"L1": []map[string]any{
										{
											"cid":       nameCreateCid,
											"fieldName": "Name",
										},
										{
											"cid":       ageCreateCid,
											"fieldName": "Age",
										},
									},
									"H1": []map[string]any{},
									"L2": []map[string]any{
										{
											"fieldName": "Name",
										},
										{
											"fieldName": "Age",
										},
									},
									"H2": []map[string]any{},
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuery_WithMultipleAliasedFilteredEmbeddedLatestCommit(t *testing.T) {
	docCreateCompositeCid := testUtils.NewUniqueValue()
	ageCreateCid := testUtils.NewUniqueValue()
	nameCreateCid := testUtils.NewUniqueValue()
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						Name
						Age
						_version {
							cid
							L1: links(filter: {fieldName: {_eq: "Age"}}) {
								cid
								fieldName
								height
							}
							L2: links(filter: {fieldName: {_eq: "Name"}}) {
								fieldName
								height
							}
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"Age":  int64(21),
							"_version": []map[string]any{
								{
									"cid": docCreateCompositeCid,
									"L1": []map[string]any{
										{
											"cid":       ageCreateCid,
											"fieldName": "Age",
											"height":    uint64(1),
										},
									},
									"L2": []map[string]any{
										{
											"fieldName": nameCreateCid,
											"height":    uint64(1),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuery_WithAllCommitFields_NoError(t *testing.T) {
	docID := testUtils.NewSameValue()
	collectionVersionID := testUtils.NewUniqueValue()
	docCreateCompositeCid := testUtils.NewUniqueValue()
	ageCreateCid := testUtils.NewUniqueValue()
	nameCreateCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: userCollectionGQLSchema,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						Name
						_docID
						_version {
							cid
							delta
							docID
							fieldName
							height
							links {
								cid
								fieldName
							}
							collectionVersionId
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":   "John",
							"_docID": docID,
							"_version": []map[string]any{
								{
									"cid":       docCreateCompositeCid,
									"delta":     nil,
									"docID":     docID,
									"fieldName": "_C",
									"height":    int64(1),
									"links": []map[string]any{
										{
											"cid":       nameCreateCid,
											"fieldName": "Name",
										},
										{
											"cid":       ageCreateCid,
											"fieldName": "Age",
										},
									},
									"collectionVersionId": collectionVersionID,
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

func TestQuery_WithAllCommitFieldsWithUpdate_NoError(t *testing.T) {
	docID := testUtils.NewSameValue()
	collectionVersionID := testUtils.NewSameValue()
	docUpdateCompositeCid := testUtils.NewUniqueValue()
	docCreateCompositeCid := testUtils.NewSameValue()
	ageUpdateCid := testUtils.NewUniqueValue()
	ageCreateCid := testUtils.NewUniqueValue()
	nameCreateCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: userCollectionGQLSchema,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"Age": 22}`,
			},
			&action.Request{
				Request: `query {
					Users {
						Name
						Age
						_docID
						_version {
							cid
							delta
							docID
							fieldName
							height
							links {
								cid
								fieldName
							}
							heads {
								cid
							}
							collectionVersionId
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":   "John",
							"Age":    int64(22),
							"_docID": docID,
							"_version": []map[string]any{
								{
									"cid":       docUpdateCompositeCid,
									"delta":     nil,
									"docID":     docID,
									"fieldName": "_C",
									"height":    int64(2),
									"links": []map[string]any{
										{
											"cid":       ageUpdateCid,
											"fieldName": "Age",
										},
									},
									"heads": []map[string]any{
										{
											"cid": docCreateCompositeCid,
										},
									},
									"collectionVersionId": collectionVersionID,
								},
								{
									"cid":       docCreateCompositeCid,
									"delta":     nil,
									"docID":     docID,
									"fieldName": "_C",
									"height":    int64(1),
									"links": []map[string]any{
										{
											"cid":       nameCreateCid,
											"fieldName": "Name",
										},
										{
											"cid":       ageCreateCid,
											"fieldName": "Age",
										},
									},
									"heads":               []map[string]any{},
									"collectionVersionId": collectionVersionID,
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
