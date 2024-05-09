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
						"cid": "bafyreidmbagmnhwb3qr5qctclsylkzgrwpbmiuxirtfbdf3fuzxbibljfi",
						"links": []map[string]any{
							{
								"cid":  "bafyreibhdfmodhqycxtw33ffdceh2wlxqlwcwbyowvs2lrlvimph7ekg2u",
								"name": "Name",
							},
							{
								"cid":  "bafyreigrxupxvzvjfx6wblmpc6fgekapr7nxlmokvi4gmz6ojmzmbrnapa",
								"name": "Age",
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
						"schemaVersionId": "bafkreigqmcqzkbg3elpe24vfza4rjle2r6cxu7ihzvg56aov57crhaebry",
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
						"cid": "bafyreidmbagmnhwb3qr5qctclsylkzgrwpbmiuxirtfbdf3fuzxbibljfi",
						"L1": []map[string]any{
							{
								"cid":  "bafyreibhdfmodhqycxtw33ffdceh2wlxqlwcwbyowvs2lrlvimph7ekg2u",
								"name": "Name",
							},
							{
								"cid":  "bafyreigrxupxvzvjfx6wblmpc6fgekapr7nxlmokvi4gmz6ojmzmbrnapa",
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
								"cid":          "bafyreidmbagmnhwb3qr5qctclsylkzgrwpbmiuxirtfbdf3fuzxbibljfi",
								"collectionID": int64(1),
								"delta":        nil,
								"docID":        "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
								"fieldId":      "C",
								"fieldName":    nil,
								"height":       int64(1),
								"links": []map[string]any{
									{
										"cid":  "bafyreibhdfmodhqycxtw33ffdceh2wlxqlwcwbyowvs2lrlvimph7ekg2u",
										"name": "Name",
									},
									{
										"cid":  "bafyreigrxupxvzvjfx6wblmpc6fgekapr7nxlmokvi4gmz6ojmzmbrnapa",
										"name": "Age",
									},
								},
								"schemaVersionId": "bafkreigqmcqzkbg3elpe24vfza4rjle2r6cxu7ihzvg56aov57crhaebry",
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
								"cid":          "bafyreibbn2vjovh65xe5v2bqxqxkb6sek5xkbnouhryya6enesbhzfplvm",
								"collectionID": int64(1),
								"delta":        nil,
								"docID":        "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
								"fieldId":      "C",
								"fieldName":    nil,
								"height":       int64(2),
								"links": []map[string]any{
									{
										"cid":  "bafyreidmbagmnhwb3qr5qctclsylkzgrwpbmiuxirtfbdf3fuzxbibljfi",
										"name": "_head",
									},
									{
										"cid":  "bafyreifycx5aqjhdlmzaf3bqb6ieomfxrzercas3hxnqcwz2jb25mkrzxi",
										"name": "Age",
									},
								},
								"schemaVersionId": "bafkreigqmcqzkbg3elpe24vfza4rjle2r6cxu7ihzvg56aov57crhaebry",
							},
							{
								"cid":          "bafyreidmbagmnhwb3qr5qctclsylkzgrwpbmiuxirtfbdf3fuzxbibljfi",
								"collectionID": int64(1),
								"delta":        nil,
								"docID":        "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
								"fieldId":      "C",
								"fieldName":    nil,
								"height":       int64(1),
								"links": []map[string]any{
									{
										"cid":  "bafyreibhdfmodhqycxtw33ffdceh2wlxqlwcwbyowvs2lrlvimph7ekg2u",
										"name": "Name",
									},
									{
										"cid":  "bafyreigrxupxvzvjfx6wblmpc6fgekapr7nxlmokvi4gmz6ojmzmbrnapa",
										"name": "Age",
									},
								},
								"schemaVersionId": "bafkreigqmcqzkbg3elpe24vfza4rjle2r6cxu7ihzvg56aov57crhaebry",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
