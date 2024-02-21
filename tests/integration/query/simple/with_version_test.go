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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithEmbeddedLatestCommit(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Embedded latest commits query within object query",
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
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
				"Age":  int64(21),
				"_version": []map[string]any{
					{
						"cid": "bafybeicojqe66grk564b2hns3zi6rhquqvugxj6wi4s6xk4e2gg65dzx5e",
						"links": []map[string]any{
							{
								"cid":  "bafybeic45t5rj54wx47fhaqm6dubwt2cf5fkqzwm2nea7ypam3f6s2zbk4",
								"name": "Age",
							},
							{
								"cid":  "bafybeifkcrogypyaq5iw7krgi5jd26s7jlfsy5u232e7e7y7dqe3wm2hcu",
								"name": "Name",
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithEmbeddedLatestCommitWithSchemaVersionId(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Embedded commits query within object query with schema version id",
		Request: `query {
					Users {
						Name
						_version {
							schemaVersionId
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
				"_version": []map[string]any{
					{
						"schemaVersionId": "bafkreiekkppcdl573ru624wh3kwkmy2nhqzjsvqpu6jv5dgq2kidpnon4u",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithEmbeddedLatestCommitWithDocID(t *testing.T) {
	const docID = "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"

	test := testUtils.RequestTestCase{
		Description: "Embedded commits query within object query with document ID",
		Request: `query {
					Users {
						Name
						_docID
						_version {
							docID
						}
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Results: []map[string]any{
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
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithMultipleAliasedEmbeddedLatestCommit(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Embedded, aliased, latest commits query within object query",
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
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Name": "John",
				"Age":  int64(21),
				"_version": []map[string]any{
					{
						"cid": "bafybeicojqe66grk564b2hns3zi6rhquqvugxj6wi4s6xk4e2gg65dzx5e",
						"L1": []map[string]any{
							{
								"cid":  "bafybeic45t5rj54wx47fhaqm6dubwt2cf5fkqzwm2nea7ypam3f6s2zbk4",
								"name": "Age",
							},
							{
								"cid":  "bafybeifkcrogypyaq5iw7krgi5jd26s7jlfsy5u232e7e7y7dqe3wm2hcu",
								"name": "Name",
							},
						},
						"L2": []map[string]any{
							{
								"name": "Age",
							},
							{
								"name": "Name",
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
	const docID = "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"

	test := testUtils.TestCase{
		Description: "Embedded commits query within object query with document ID",
		Actions: []any{
			testUtils.SchemaUpdate{
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
							collectionID
							delta
							docID
							fieldId
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
				Results: []map[string]any{
					{
						"Name":   "John",
						"_docID": docID,
						"_version": []map[string]any{
							{
								"cid":          "bafybeicojqe66grk564b2hns3zi6rhquqvugxj6wi4s6xk4e2gg65dzx5e",
								"collectionID": int64(1),
								"delta":        nil,
								"docID":        "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
								"fieldId":      "C",
								"fieldName":    nil,
								"height":       int64(1),
								"links": []map[string]any{
									{
										"cid":  "bafybeic45t5rj54wx47fhaqm6dubwt2cf5fkqzwm2nea7ypam3f6s2zbk4",
										"name": "Age",
									},
									{
										"cid":  "bafybeifkcrogypyaq5iw7krgi5jd26s7jlfsy5u232e7e7y7dqe3wm2hcu",
										"name": "Name",
									},
								},
								"schemaVersionId": "bafkreiekkppcdl573ru624wh3kwkmy2nhqzjsvqpu6jv5dgq2kidpnon4u",
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
	const docID = "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"

	test := testUtils.TestCase{
		Description: "Embedded commits query within object query with document ID",
		Actions: []any{
			testUtils.SchemaUpdate{
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
							collectionID
							delta
							docID
							fieldId
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
				Results: []map[string]any{
					{
						"Name":   "John",
						"Age":    int64(22),
						"_docID": docID,
						"_version": []map[string]any{
							{
								"cid":          "bafybeigcjabzlkuj4j35boczgcl4jmars7gz5a7dfvpq3m344bzth7ebqq",
								"collectionID": int64(1),
								"delta":        nil,
								"docID":        "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
								"fieldId":      "C",
								"fieldName":    nil,
								"height":       int64(2),
								"links": []map[string]any{
									{
										"cid":  "bafybeihzra5nmcai4omdv2hkplrpexjsau62eaa2ndrf2b7ksxvl7hx3qm",
										"name": "Age",
									},
									{
										"cid":  "bafybeicojqe66grk564b2hns3zi6rhquqvugxj6wi4s6xk4e2gg65dzx5e",
										"name": "_head",
									},
								},
								"schemaVersionId": "bafkreiekkppcdl573ru624wh3kwkmy2nhqzjsvqpu6jv5dgq2kidpnon4u",
							},
							{
								"cid":          "bafybeicojqe66grk564b2hns3zi6rhquqvugxj6wi4s6xk4e2gg65dzx5e",
								"collectionID": int64(1),
								"delta":        nil,
								"docID":        "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
								"fieldId":      "C",
								"fieldName":    nil,
								"height":       int64(1),
								"links": []map[string]any{
									{
										"cid":  "bafybeic45t5rj54wx47fhaqm6dubwt2cf5fkqzwm2nea7ypam3f6s2zbk4",
										"name": "Age",
									},
									{
										"cid":  "bafybeifkcrogypyaq5iw7krgi5jd26s7jlfsy5u232e7e7y7dqe3wm2hcu",
										"name": "Name",
									},
								},
								"schemaVersionId": "bafkreiekkppcdl573ru624wh3kwkmy2nhqzjsvqpu6jv5dgq2kidpnon4u",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
