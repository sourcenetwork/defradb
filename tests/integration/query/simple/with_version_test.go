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
						"cid": "bafybeigwxfw2nfcwelqxzgjsmm5okrt7dctzvzml4tm7i7q7fsdit3ihz4",
						"links": []map[string]any{
							{
								"cid":  "bafybeigcmjyt2ux4mzfckbsz5snkoqrr42vfkesgk7rdw6xzblrowrzfg4",
								"name": "Age",
							},
							{
								"cid":  "bafybeihkekm4kfn2ttx3wb33l2ps7aductuzd7hrmu6n7zloaicrj5n75u",
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
						"schemaVersionId": "bafkreidvd63bawkelxe3wtf7a65klkq4x3dvenqafyasndyal6fvffkeam",
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
						"cid": "bafybeigwxfw2nfcwelqxzgjsmm5okrt7dctzvzml4tm7i7q7fsdit3ihz4",
						"L1": []map[string]any{
							{
								"cid":  "bafybeigcmjyt2ux4mzfckbsz5snkoqrr42vfkesgk7rdw6xzblrowrzfg4",
								"name": "Age",
							},
							{
								"cid":  "bafybeihkekm4kfn2ttx3wb33l2ps7aductuzd7hrmu6n7zloaicrj5n75u",
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
