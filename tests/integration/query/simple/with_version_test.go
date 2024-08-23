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
									"cid": "bafyreiby7drdzfsg4wwo7f6vkdqhurbe74s4lhayn3k3226zvkgwjd2fbu",
									"links": []map[string]any{
										{
											"cid":  "bafyreid4sasigytiflrh3rupyevo6wy43b6mlfi4jwkjjwvohgjcd3oscu",
											"name": "Age",
										},
										{
											"cid":  "bafyreieg3p2kpyxwiowskvb3pp35nedzawmapjuib7glrvszcgmv6z37fm",
											"name": "Name",
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
									"schemaVersionId": "bafkreigqmcqzkbg3elpe24vfza4rjle2r6cxu7ihzvg56aov57crhaebry",
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
	const docID = "bae-d4303725-7db9-53d2-b324-f3ee44020e52"

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
									"cid": "bafyreiby7drdzfsg4wwo7f6vkdqhurbe74s4lhayn3k3226zvkgwjd2fbu",
									"L1": []map[string]any{
										{
											"cid":  "bafyreid4sasigytiflrh3rupyevo6wy43b6mlfi4jwkjjwvohgjcd3oscu",
											"name": "Age",
										},
										{
											"cid":  "bafyreieg3p2kpyxwiowskvb3pp35nedzawmapjuib7glrvszcgmv6z37fm",
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
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuery_WithAllCommitFields_NoError(t *testing.T) {
	const docID = "bae-d4303725-7db9-53d2-b324-f3ee44020e52"

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
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":   "John",
							"_docID": docID,
							"_version": []map[string]any{
								{
									"cid":          "bafyreiby7drdzfsg4wwo7f6vkdqhurbe74s4lhayn3k3226zvkgwjd2fbu",
									"collectionID": int64(1),
									"delta":        nil,
									"docID":        "bae-d4303725-7db9-53d2-b324-f3ee44020e52",
									"fieldId":      "C",
									"fieldName":    nil,
									"height":       int64(1),
									"links": []map[string]any{
										{
											"cid":  "bafyreid4sasigytiflrh3rupyevo6wy43b6mlfi4jwkjjwvohgjcd3oscu",
											"name": "Age",
										},
										{
											"cid":  "bafyreieg3p2kpyxwiowskvb3pp35nedzawmapjuib7glrvszcgmv6z37fm",
											"name": "Name",
										},
									},
									"schemaVersionId": "bafkreigqmcqzkbg3elpe24vfza4rjle2r6cxu7ihzvg56aov57crhaebry",
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
	const docID = "bae-d4303725-7db9-53d2-b324-f3ee44020e52"

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
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":   "John",
							"Age":    int64(22),
							"_docID": docID,
							"_version": []map[string]any{
								{
									"cid":          "bafyreigfstknvmsl77pg443lqqf2g64y7hr575tts5c4nxuzk3dynffkem",
									"collectionID": int64(1),
									"delta":        nil,
									"docID":        docID,
									"fieldId":      "C",
									"fieldName":    nil,
									"height":       int64(2),
									"links": []map[string]any{
										{
											"cid":  "bafyreiby7drdzfsg4wwo7f6vkdqhurbe74s4lhayn3k3226zvkgwjd2fbu",
											"name": "_head",
										},
										{
											"cid":  "bafyreiapjg22e47sanhjtqgu453mvmxcfcl4ksrcoctyfl6nfsh3xwfcvm",
											"name": "Age",
										},
									},
									"schemaVersionId": "bafkreigqmcqzkbg3elpe24vfza4rjle2r6cxu7ihzvg56aov57crhaebry",
								},
								{
									"cid":          "bafyreiby7drdzfsg4wwo7f6vkdqhurbe74s4lhayn3k3226zvkgwjd2fbu",
									"collectionID": int64(1),
									"delta":        nil,
									"docID":        docID,
									"fieldId":      "C",
									"fieldName":    nil,
									"height":       int64(1),
									"links": []map[string]any{
										{
											"cid":  "bafyreid4sasigytiflrh3rupyevo6wy43b6mlfi4jwkjjwvohgjcd3oscu",
											"name": "Age",
										},
										{
											"cid":  "bafyreieg3p2kpyxwiowskvb3pp35nedzawmapjuib7glrvszcgmv6z37fm",
											"name": "Name",
										},
									},
									"schemaVersionId": "bafkreigqmcqzkbg3elpe24vfza4rjle2r6cxu7ihzvg56aov57crhaebry",
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
