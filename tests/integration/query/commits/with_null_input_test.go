// Copyright 2024 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsWithNullDepth(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(depth: null) {
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithNullCID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(cid: null) {
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithNullField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(filter: {fieldName: {_eq: null}}) {
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

func TestQueryCommitsWithNullOrder(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(order: null) {
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithNullOrderField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(docID: null) {
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithNullLimit(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(limit: null) {
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithNullOffset(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(offset: null) {
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithNullGroupBy(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(groupBy: null) {
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
