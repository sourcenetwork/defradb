// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package commits

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsWithUnknownDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						_commits(docID: "unknown document ID") {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738") {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": "bafyreiajq6jmyblg2b6vupjdapzkaodbt7kkwqp4fijekdvydnyxvr4y7q",
						},
						{
							"cid": "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm",
						},
						{
							"cid": "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndLinks(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738") {
							cid
							links {
								cid
								fieldName
							}
							heads {
								cid
							}
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":   "bafyreiajq6jmyblg2b6vupjdapzkaodbt7kkwqp4fijekdvydnyxvr4y7q",
							"links": []map[string]any{},
							"heads": []map[string]any{},
						},
						{
							"cid":   "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm",
							"links": []map[string]any{},
							"heads": []map[string]any{},
						},
						{
							"cid": "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
							"links": []map[string]any{
								{
									"cid":       "bafyreiajq6jmyblg2b6vupjdapzkaodbt7kkwqp4fijekdvydnyxvr4y7q",
									"fieldName": "age",
								},
								{
									"cid":       "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm",
									"fieldName": "name",
								},
							},
							"heads": []map[string]any{},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: `query {
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738") {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":    "bafyreihht6jz3vxk3fvr4sp3kqnvuplmva36hivbjtpdum7zydvb2yztwu",
							"height": int64(2),
						},
						{
							"cid":    "bafyreiajq6jmyblg2b6vupjdapzkaodbt7kkwqp4fijekdvydnyxvr4y7q",
							"height": int64(1),
						},
						{
							"cid":    "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm",
							"height": int64(1),
						},
						{
							"cid":    "bafyreia4x5ju33jenbimdqbtnuqc7pby4lydpa7efyk5iu4nl6urm6ofla",
							"height": int64(2),
						},
						{
							"cid":    "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
							"height": int64(1),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (first results includes link._head, second
// includes link._Name).
func TestQueryCommitsWithDocIDAndUpdateAndLinks(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: `query {
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738") {
							cid
							links {
								cid
								fieldName
							}
							heads {
								cid
							}
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":   "bafyreihht6jz3vxk3fvr4sp3kqnvuplmva36hivbjtpdum7zydvb2yztwu",
							"links": []map[string]any{},
							"heads": []map[string]any{
								{
									"cid": "bafyreiajq6jmyblg2b6vupjdapzkaodbt7kkwqp4fijekdvydnyxvr4y7q",
								},
							},
						},
						{
							"cid":   "bafyreiajq6jmyblg2b6vupjdapzkaodbt7kkwqp4fijekdvydnyxvr4y7q",
							"links": []map[string]any{},
							"heads": []map[string]any{},
						},
						{
							"cid":   "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm",
							"links": []map[string]any{},
							"heads": []map[string]any{},
						},
						{
							"cid": "bafyreia4x5ju33jenbimdqbtnuqc7pby4lydpa7efyk5iu4nl6urm6ofla",
							"links": []map[string]any{
								{
									"cid":       "bafyreihht6jz3vxk3fvr4sp3kqnvuplmva36hivbjtpdum7zydvb2yztwu",
									"fieldName": "age",
								},
							},
							"heads": []map[string]any{
								{
									"cid": "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
								},
							},
						},
						{
							"cid": "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
							"links": []map[string]any{
								{
									"cid":       "bafyreiajq6jmyblg2b6vupjdapzkaodbt7kkwqp4fijekdvydnyxvr4y7q",
									"fieldName": "age",
								},
								{
									"cid":       "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm",
									"fieldName": "name",
								},
							},
							"heads": []map[string]any{},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
