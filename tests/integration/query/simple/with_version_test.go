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
						"cid": "bafybeib26cyuzbnf7uq3js5mykfveplsn4imo2fmf2jnnib6rrtnllv4pe",
						"links": []map[string]any{
							{
								"cid":  "bafybeihkhgtdogxwqe2lkjqord5bzthfwwthyo3gu6iljfm5l7n7fkhpsq",
								"name": "Age",
							},
							{
								"cid":  "bafybeico2g2tdkpo4i64ph6b5vgngn5zbxus4jxwav3bi2joieqicplfxi",
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
						"schemaVersionId": "bafkreihuvcb7e7vy6ua3yrwbwnul3djqrtbhyuv3c4dqe4y3i2ssudzveu",
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
						"cid": "bafybeib26cyuzbnf7uq3js5mykfveplsn4imo2fmf2jnnib6rrtnllv4pe",
						"L1": []map[string]any{
							{
								"cid":  "bafybeihkhgtdogxwqe2lkjqord5bzthfwwthyo3gu6iljfm5l7n7fkhpsq",
								"name": "Age",
							},
							{
								"cid":  "bafybeico2g2tdkpo4i64ph6b5vgngn5zbxus4jxwav3bi2joieqicplfxi",
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
								"cid":          "bafybeib26cyuzbnf7uq3js5mykfveplsn4imo2fmf2jnnib6rrtnllv4pe",
								"collectionID": int64(1),
								"delta":        nil,
								"docID":        "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
								"fieldId":      "C",
								"fieldName":    nil,
								"height":       int64(1),
								"links": []map[string]any{
									{
										"cid":  "bafybeihkhgtdogxwqe2lkjqord5bzthfwwthyo3gu6iljfm5l7n7fkhpsq",
										"name": "Age",
									},
									{
										"cid":  "bafybeico2g2tdkpo4i64ph6b5vgngn5zbxus4jxwav3bi2joieqicplfxi",
										"name": "Name",
									},
								},
								"schemaVersionId": "bafkreihuvcb7e7vy6ua3yrwbwnul3djqrtbhyuv3c4dqe4y3i2ssudzveu",
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
								"cid":          "bafybeie23a5xsx4qyoffa3riij3kei5to54bb6gq7m4lftfjujaohkabwu",
								"collectionID": int64(1),
								"delta":        nil,
								"docID":        "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
								"fieldId":      "C",
								"fieldName":    nil,
								"height":       int64(2),
								"links": []map[string]any{
									{
										"cid":  "bafybeicixwhd4prbj2jfnkkc3a7wr2f5twppyhivg3kajoe7jal5cvrdza",
										"name": "Age",
									},
									{
										"cid":  "bafybeib26cyuzbnf7uq3js5mykfveplsn4imo2fmf2jnnib6rrtnllv4pe",
										"name": "_head",
									},
								},
								"schemaVersionId": "bafkreihuvcb7e7vy6ua3yrwbwnul3djqrtbhyuv3c4dqe4y3i2ssudzveu",
							},
							{
								"cid":          "bafybeib26cyuzbnf7uq3js5mykfveplsn4imo2fmf2jnnib6rrtnllv4pe",
								"collectionID": int64(1),
								"delta":        nil,
								"docID":        "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
								"fieldId":      "C",
								"fieldName":    nil,
								"height":       int64(1),
								"links": []map[string]any{
									{
										"cid":  "bafybeihkhgtdogxwqe2lkjqord5bzthfwwthyo3gu6iljfm5l7n7fkhpsq",
										"name": "Age",
									},
									{
										"cid":  "bafybeico2g2tdkpo4i64ph6b5vgngn5zbxus4jxwav3bi2joieqicplfxi",
										"name": "Name",
									},
								},
								"schemaVersionId": "bafkreihuvcb7e7vy6ua3yrwbwnul3djqrtbhyuv3c4dqe4y3i2ssudzveu",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
