// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryWithIndexOnOneToManyRelation_IfFilterOnIndexedRelation_ShouldFilter(t *testing.T) {
	req1 := `query {
		User(filter: {
			devices: {model: {_eq: "MacBook Pro"}}
		}) {
			name
		}
	}`
	req2 := `query {
		User(filter: {
			devices: {model: {_eq: "iPhone 10"}}
		}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Filter on indexed relation field in 1-N relation",
		Actions: []any{
			createSchemaWithDocs(`
				type User {
					name: String 
					age: Int
					devices: [Device] 
				} 

				type Device {
					model: String @index
					owner: User
				} 
			`),
			testUtils.Request{
				Request: req1,
				Results: []map[string]any{
					{"name": "Islam"},
					{"name": "Shahzad"},
					{"name": "Keenan"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithDocFetches(6).WithFieldFetches(9).WithIndexFetches(3),
			},
			testUtils.Request{
				Request: req2,
				Results: []map[string]any{
					{"name": "Addo"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithDocFetches(2).WithFieldFetches(3).WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToOnesSecondaryRelation_IfFilterOnIndexedRelation_ShouldFilter(t *testing.T) {
	req1 := `query {
		User(filter: {
			address: {city: {_eq: "Munich"}}
		}) {
			name
		}
	}`
	req2 := `query {
		User(filter: {
			address: {city: {_eq: "Montreal"}}
		}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Filter on indexed secondary relation field in 1-1 relation",
		Actions: []any{
			createSchemaWithDocs(`
				type User {
					name: String 
					age: Int
					address: Address
				} 

				type Address {
					user: User
					city: String @index
				} 
			`),
			testUtils.Request{
				Request: req1,
				Results: []map[string]any{
					{"name": "Islam"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithDocFetches(2).WithFieldFetches(3).WithIndexFetches(1),
			},
			testUtils.Request{
				Request: req2,
				Results: []map[string]any{
					{"name": "Shahzad"},
					{"name": "Fred"},
					{"name": "John"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithDocFetches(6).WithFieldFetches(9).WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToOnePrimaryRelation_IfFilterOnIndexedRelation_ShouldFilter(t *testing.T) {
	req1 := `query {
		User(filter: {
			address: {city: {_eq: "Munich"}}
		}) {
			name
		}
	}`
	req2 := `query {
		User(filter: {
			address: {city: {_eq: "Montreal"}}
		}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Filter on indexed primary relation field in 1-1 relation",
		Actions: []any{
			createSchemaWithDocs(`
				type User {
					name: String 
					age: Int
					address: Address @primary 
				} 

				type Address {
					user: User
					city: String @index
					street: String 
				} 
			`),
			testUtils.Request{
				Request: req1,
				Results: []map[string]any{
					{"name": "Islam"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithDocFetches(2).WithFieldFetches(3).WithIndexFetches(1),
			},
			testUtils.Request{
				Request: req2,
				Results: []map[string]any{
					{"name": "John"},
					{"name": "Fred"},
					{"name": "Shahzad"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithDocFetches(14).WithFieldFetches(17).WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToTwoRelation_IfFilterOnIndexedRelation_ShouldFilter(t *testing.T) {
	req1 := `query {
		User(filter: {
			address: {city: {_eq: "Munich"}}
		}) {
			name
			address {
				city
			}
		}
	}`
	req2 := `query {
		User(filter: {
			devices: {model: {_eq: "Walkman"}}
		}) {
			name
			devices {
				model
			}
		}
	}`
	test := testUtils.TestCase{
		Description: "Filter on indexed relation field in 1-1 and 1-N relations",
		Actions: []any{
			createSchemaWithDocs(`
				type User {
					name: String 
					age: Int
					address: Address
					devices: [Device] 
				} 

				type Device {
					model: String @index
					owner: User
				} 

				type Address {
					user: User
					city: String @index
				} 
			`),
			testUtils.Request{
				Request: req1,
				Results: []map[string]any{
					{
						"name": "Islam",
						"address": map[string]any{
							"city": "Munich",
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithDocFetches(2).WithFieldFetches(3).WithIndexFetches(1),
			},
			testUtils.Request{
				Request: req2,
				Results: []map[string]any{
					{
						"name": "Chris",
						"devices": map[string]any{
							"model": "Walkman",
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithDocFetches(2).WithFieldFetches(3).WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
