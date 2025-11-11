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

func TestQuerySimpleWithEmbeddedLatestCommit(t *testing.T) {
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
								name
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
											"cid":  "bafyreih4kr6m7wil7xgvkwktbnfab4fs6hrhytf62wogov2i4bjzjddk2m",
											"name": "Name",
										},
										{
											"cid":  "bafyreih5pxyir6jxoeb2lptmoiwkvixz4p2fty6jpfztq5setgnf3f4mru",
											"name": "Age",
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
								name
							}
							L2: links {
								name
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
											"cid":  "bafyreih4kr6m7wil7xgvkwktbnfab4fs6hrhytf62wogov2i4bjzjddk2m",
											"name": "Name",
										},
										{
											"cid":  "bafyreih5pxyir6jxoeb2lptmoiwkvixz4p2fty6jpfztq5setgnf3f4mru",
											"name": "Age",
										},
									},
									"L2": []map[string]any{
										{
											"name": "Name",
										},
										{
											"name": "Age",
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
								name
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
											"cid":  "bafyreih4kr6m7wil7xgvkwktbnfab4fs6hrhytf62wogov2i4bjzjddk2m",
											"name": "Name",
										},
										{
											"cid":  "bafyreih5pxyir6jxoeb2lptmoiwkvixz4p2fty6jpfztq5setgnf3f4mru",
											"name": "Age",
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
								name
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
											"cid":  "bafyreieoljg2ynsazfcesosye5gc2zcl2bgyuefjintc4eu7hrbzfvbdli",
											"name": "_head",
										},
										{
											"cid":  "bafyreibvn3oanzbe4uxw2vocro6u7widukuriwi4fctt7jx3np425nxzqa",
											"name": "Age",
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
											"cid":  "bafyreih4kr6m7wil7xgvkwktbnfab4fs6hrhytf62wogov2i4bjzjddk2m",
											"name": "Name",
										},
										{
											"cid":  "bafyreih5pxyir6jxoeb2lptmoiwkvixz4p2fty6jpfztq5setgnf3f4mru",
											"name": "Age",
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
