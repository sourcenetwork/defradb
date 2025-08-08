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
		Description: "Embedded latest commits query within object query",
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
									"cid": "bafyreifpahyxbugzl6viejumx5ykqbysgjbkqj322h3bpltxx4edu7lguq",
									"links": []map[string]any{
										{
											"cid":  "bafyreia5rwzyr4fjirpyhd7mhyzx3zvha3bgrzp567nmyatczdypr2mbue",
											"name": "Name",
										},
										{
											"cid":  "bafyreidjvwwjyxtle526qlmc3u5pib3qvavcpetgdwyigl4vv6scdflwse",
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
		Description: "Embedded commits query within object query with schema version id",
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
									"schemaVersionId": "bafyreib6m76pzu3y5h2zfbguioxrb4mfpsnyanxeyfltns4usaamrlewgm",
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
	const docID = "bae-75cb8b0a-00d7-57c8-8906-29687cbbb15c"

	test := testUtils.TestCase{
		Description: "Embedded commits query within object query with document ID",
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
		Description: "Embedded, aliased, latest commits query within object query",
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
									"cid": "bafyreifpahyxbugzl6viejumx5ykqbysgjbkqj322h3bpltxx4edu7lguq",
									"L1": []map[string]any{
										{
											"cid":  "bafyreia5rwzyr4fjirpyhd7mhyzx3zvha3bgrzp567nmyatczdypr2mbue",
											"name": "Name",
										},
										{
											"cid":  "bafyreidjvwwjyxtle526qlmc3u5pib3qvavcpetgdwyigl4vv6scdflwse",
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
	const docID = "bae-75cb8b0a-00d7-57c8-8906-29687cbbb15c"

	test := testUtils.TestCase{
		Description: "Embedded commits query within object query with document ID",
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
									"cid":       "bafyreifpahyxbugzl6viejumx5ykqbysgjbkqj322h3bpltxx4edu7lguq",
									"delta":     nil,
									"docID":     "bae-75cb8b0a-00d7-57c8-8906-29687cbbb15c",
									"fieldName": "_C",
									"height":    int64(1),
									"links": []map[string]any{
										{
											"cid":  "bafyreia5rwzyr4fjirpyhd7mhyzx3zvha3bgrzp567nmyatczdypr2mbue",
											"name": "Name",
										},
										{
											"cid":  "bafyreidjvwwjyxtle526qlmc3u5pib3qvavcpetgdwyigl4vv6scdflwse",
											"name": "Age",
										},
									},
									"schemaVersionId": "bafyreib6m76pzu3y5h2zfbguioxrb4mfpsnyanxeyfltns4usaamrlewgm",
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
	const docID = "bae-75cb8b0a-00d7-57c8-8906-29687cbbb15c"

	test := testUtils.TestCase{
		Description: "Embedded commits query within object query with document ID",
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
									"cid":       "bafyreidhmnxhossvuoanhgj3hoylwxpqslzajwc6nkag364t6myhboyp2y",
									"delta":     nil,
									"docID":     docID,
									"fieldName": "_C",
									"height":    int64(2),
									"links": []map[string]any{
										{
											"cid":  "bafyreifpahyxbugzl6viejumx5ykqbysgjbkqj322h3bpltxx4edu7lguq",
											"name": "_head",
										},
										{
											"cid":  "bafyreierm7op3gi6xdlbidavwycmb772ia4j7dz7rj7pw62bpggibeavem",
											"name": "Age",
										},
									},
									"schemaVersionId": "bafyreib6m76pzu3y5h2zfbguioxrb4mfpsnyanxeyfltns4usaamrlewgm",
								},
								{
									"cid":       "bafyreifpahyxbugzl6viejumx5ykqbysgjbkqj322h3bpltxx4edu7lguq",
									"delta":     nil,
									"docID":     docID,
									"fieldName": "_C",
									"height":    int64(1),
									"links": []map[string]any{
										{
											"cid":  "bafyreia5rwzyr4fjirpyhd7mhyzx3zvha3bgrzp567nmyatczdypr2mbue",
											"name": "Name",
										},
										{
											"cid":  "bafyreidjvwwjyxtle526qlmc3u5pib3qvavcpetgdwyigl4vv6scdflwse",
											"name": "Age",
										},
									},
									"schemaVersionId": "bafyreib6m76pzu3y5h2zfbguioxrb4mfpsnyanxeyfltns4usaamrlewgm",
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
