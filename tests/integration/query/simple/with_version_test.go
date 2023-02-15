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
					users {
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
				"Age":  uint64(21),
				"_version": []map[string]any{
					{
						"cid": "bafybeigpevf7mynknumjabpd3vjohqfyh5fciob236almqlf6m6i5ldhqq",
						"links": []map[string]any{
							{
								"cid":  "bafybeihsywma4bokoar5ajqzv6gdqcy5zl2ckksn7rslwpqmlizf5z23xm",
								"name": "Age",
							},
							{
								"cid":  "bafybeigazxmseq3siyopmp6tqclx7jjzzqo5w5a7gtttwtjqmbzioqjfqq",
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
					users {
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
						"schemaVersionId": "bafkreicl25inntuxlx4fbbcwb2tzbja4iyvujjwhk25pen7ljjdzcpgup4",
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
					users {
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
				"Age":  uint64(21),
				"_version": []map[string]any{
					{
						"cid": "bafybeigpevf7mynknumjabpd3vjohqfyh5fciob236almqlf6m6i5ldhqq",
						"L1": []map[string]any{
							{
								"cid":  "bafybeihsywma4bokoar5ajqzv6gdqcy5zl2ckksn7rslwpqmlizf5z23xm",
								"name": "Age",
							},
							{
								"cid":  "bafybeigazxmseq3siyopmp6tqclx7jjzzqo5w5a7gtttwtjqmbzioqjfqq",
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
