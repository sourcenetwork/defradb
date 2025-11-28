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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithNestedLatestCommit(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.Request{
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
									"cid": "bafyreieoljg2ynsazfcesosye5gc2zcl2bgyuefjintc4eu7hrbzfvbdli",
									"links": []map[string]any{
										{
											"cid":       "bafyreih4kr6m7wil7xgvkwktbnfab4fs6hrhytf62wogov2i4bjzjddk2m",
											"fieldName": "Name",
										},
										{
											"cid":       "bafyreih5pxyir6jxoeb2lptmoiwkvixz4p2fty6jpfztq5setgnf3f4mru",
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

func TestQuery_CreateDocWithNestedLatestCommit(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.Request{
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
									"cid": "bafyreieoljg2ynsazfcesosye5gc2zcl2bgyuefjintc4eu7hrbzfvbdli",
									"links": []map[string]any{
										{
											"cid":       "bafyreih4kr6m7wil7xgvkwktbnfab4fs6hrhytf62wogov2i4bjzjddk2m",
											"fieldName": "Name",
											"height":    uint64(1),
										},
										{
											"cid":       "bafyreih5pxyir6jxoeb2lptmoiwkvixz4p2fty6jpfztq5setgnf3f4mru",
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
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"Age": 22
				}`,
			},
			testUtils.Request{
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
									"cid": "bafyreigftnzgbysanputrc65nys3feyebwprvh3x3hucjt5xrkpog2auay",
									"links": []map[string]any{
										{
											"cid":       "bafyreibvn3oanzbe4uxw2vocro6u7widukuriwi4fctt7jx3np425nxzqa",
											"fieldName": "Age",
											"height":    uint64(2),
										},
									},
									"heads": []map[string]any{
										{
											"cid":    "bafyreieoljg2ynsazfcesosye5gc2zcl2bgyuefjintc4eu7hrbzfvbdli",
											"height": uint64(1),
										},
									},
								},
								{
									"cid": "bafyreieoljg2ynsazfcesosye5gc2zcl2bgyuefjintc4eu7hrbzfvbdli",
									"links": []map[string]any{
										{
											"cid":       "bafyreih4kr6m7wil7xgvkwktbnfab4fs6hrhytf62wogov2i4bjzjddk2m",
											"fieldName": "Name",
											"height":    uint64(1),
										},
										{
											"cid":       "bafyreih5pxyir6jxoeb2lptmoiwkvixz4p2fty6jpfztq5setgnf3f4mru",
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

func TestQuerySimpleWithEmbeddedLatestCommitWithSchemaVersionID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						Name
						_version {
							schemaVersionId
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"_version": []map[string]any{
								{
									"schemaVersionId": "bafyreia4ba6igfuvhp225vxxqpkn46lecvkih74g3wxvglum5nnv26m66e",
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
	const docID = "bae-619ea0d2-35ba-5e8c-ac4d-2b769937213b"

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.Request{
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
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.Request{
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
									"cid": "bafyreieoljg2ynsazfcesosye5gc2zcl2bgyuefjintc4eu7hrbzfvbdli",
									"L1": []map[string]any{
										{
											"cid":       "bafyreih4kr6m7wil7xgvkwktbnfab4fs6hrhytf62wogov2i4bjzjddk2m",
											"fieldName": "Name",
										},
										{
											"cid":       "bafyreih5pxyir6jxoeb2lptmoiwkvixz4p2fty6jpfztq5setgnf3f4mru",
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
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"Age": 22
				}`,
			},
			testUtils.Request{
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
									"cid": "bafyreigftnzgbysanputrc65nys3feyebwprvh3x3hucjt5xrkpog2auay",
									"L1": []map[string]any{
										{
											"cid":       "bafyreibvn3oanzbe4uxw2vocro6u7widukuriwi4fctt7jx3np425nxzqa",
											"fieldName": "Age",
										},
									},
									"H1": []map[string]any{
										{
											"cid": "bafyreieoljg2ynsazfcesosye5gc2zcl2bgyuefjintc4eu7hrbzfvbdli",
										},
									},
									"L2": []map[string]any{
										{
											"fieldName": "Age",
										},
									},
									"H2": []map[string]any{
										{
											"cid":    "bafyreieoljg2ynsazfcesosye5gc2zcl2bgyuefjintc4eu7hrbzfvbdli",
											"height": uint64(1),
										},
									},
								},
								{
									"cid": "bafyreieoljg2ynsazfcesosye5gc2zcl2bgyuefjintc4eu7hrbzfvbdli",
									"L1": []map[string]any{
										{
											"cid":       "bafyreih4kr6m7wil7xgvkwktbnfab4fs6hrhytf62wogov2i4bjzjddk2m",
											"fieldName": "Name",
										},
										{
											"cid":       "bafyreih5pxyir6jxoeb2lptmoiwkvixz4p2fty6jpfztq5setgnf3f4mru",
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
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						Name
						Age
						_version {
							cid
							L1: links(fieldName: "Age") {
								cid
								fieldName
								height
							}
							L2: links(fieldName: "Name") {
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
									"cid": "bafyreieoljg2ynsazfcesosye5gc2zcl2bgyuefjintc4eu7hrbzfvbdli",
									"L1": []map[string]any{
										{
											"cid":       "bafyreih5pxyir6jxoeb2lptmoiwkvixz4p2fty6jpfztq5setgnf3f4mru",
											"fieldName": "Age",
											"height":    uint64(1),
										},
									},
									"L2": []map[string]any{
										{
											"fieldName": "Name",
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
	const docID = "bae-619ea0d2-35ba-5e8c-ac4d-2b769937213b"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userCollectionGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.Request{
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
							schemaVersionId
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
									"cid":       "bafyreieoljg2ynsazfcesosye5gc2zcl2bgyuefjintc4eu7hrbzfvbdli",
									"delta":     nil,
									"docID":     "bae-619ea0d2-35ba-5e8c-ac4d-2b769937213b",
									"fieldName": "_C",
									"height":    int64(1),
									"links": []map[string]any{
										{
											"cid":       "bafyreih4kr6m7wil7xgvkwktbnfab4fs6hrhytf62wogov2i4bjzjddk2m",
											"fieldName": "Name",
										},
										{
											"cid":       "bafyreih5pxyir6jxoeb2lptmoiwkvixz4p2fty6jpfztq5setgnf3f4mru",
											"fieldName": "Age",
										},
									},
									"schemaVersionId": "bafyreia4ba6igfuvhp225vxxqpkn46lecvkih74g3wxvglum5nnv26m66e",
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
	const docID = "bae-619ea0d2-35ba-5e8c-ac4d-2b769937213b"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userCollectionGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"Age": 22}`,
			},
			testUtils.Request{
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
							schemaVersionId
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
									"cid":       "bafyreigftnzgbysanputrc65nys3feyebwprvh3x3hucjt5xrkpog2auay",
									"delta":     nil,
									"docID":     docID,
									"fieldName": "_C",
									"height":    int64(2),
									"links": []map[string]any{
										{
											"cid":       "bafyreibvn3oanzbe4uxw2vocro6u7widukuriwi4fctt7jx3np425nxzqa",
											"fieldName": "Age",
										},
									},
									"heads": []map[string]any{
										{
											"cid": "bafyreieoljg2ynsazfcesosye5gc2zcl2bgyuefjintc4eu7hrbzfvbdli",
										},
									},
									"schemaVersionId": "bafyreia4ba6igfuvhp225vxxqpkn46lecvkih74g3wxvglum5nnv26m66e",
								},
								{
									"cid":       "bafyreieoljg2ynsazfcesosye5gc2zcl2bgyuefjintc4eu7hrbzfvbdli",
									"delta":     nil,
									"docID":     docID,
									"fieldName": "_C",
									"height":    int64(1),
									"links": []map[string]any{
										{
											"cid":       "bafyreih4kr6m7wil7xgvkwktbnfab4fs6hrhytf62wogov2i4bjzjddk2m",
											"fieldName": "Name",
										},
										{
											"cid":       "bafyreih5pxyir6jxoeb2lptmoiwkvixz4p2fty6jpfztq5setgnf3f4mru",
											"fieldName": "Age",
										},
									},
									"heads":           []map[string]any{},
									"schemaVersionId": "bafyreia4ba6igfuvhp225vxxqpkn46lecvkih74g3wxvglum5nnv26m66e",
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
