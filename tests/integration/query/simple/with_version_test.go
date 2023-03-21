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
						"cid": "bafybeidb2ddr5o374hkaf5m6ziihucwn4lhzhyzi3dddj3afgfpu7p7tb4",
						"links": []map[string]any{
							{
								"cid":  "bafybeihbnch3akfu22wlbq3es7vudxev5ymo2deq2inknxhbpoacd7x5aq",
								"name": "Age",
							},
							{
								"cid":  "bafybeib6ujnjsuox6fnzrwoepvyobyxk26hlczefwrv2myuefbenlglyzy",
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

func TestQuerySimpleWithEmbeddedLatestCommitWithDockey(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Embedded commits query within object query with schema version id",
		Request: `query {
					users {
						Name
						_version {
							dockey
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
						"dockey": "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
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
						"cid": "bafybeidb2ddr5o374hkaf5m6ziihucwn4lhzhyzi3dddj3afgfpu7p7tb4",
						"L1": []map[string]any{
							{
								"cid":  "bafybeihbnch3akfu22wlbq3es7vudxev5ymo2deq2inknxhbpoacd7x5aq",
								"name": "Age",
							},
							{
								"cid":  "bafybeib6ujnjsuox6fnzrwoepvyobyxk26hlczefwrv2myuefbenlglyzy",
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
