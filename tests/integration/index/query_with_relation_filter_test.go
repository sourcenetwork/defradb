// Copyright 2022 Democratized Data Foundation
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
			sendRequestAndExplain(`
				User(filter: {
					devices: {model: {_eq: "MacBook Pro"}}
				}) {
					name
				}`,
				[]map[string]any{
					{"name": "Islam"},
					{"name": "Shahzad"},
					{"name": "Keenan"},
				},
				testUtils.NewExplainAsserter().WithDocFetches(6).WithFieldFetches(9).WithIndexFetches(3),
			),
			sendRequestAndExplain(`
				User(filter: {
					devices: {model: {_eq: "iPhone 10"}}
				}) {
					name
				}`,
				[]map[string]any{
					{"name": "Addo"},
				},
				testUtils.NewExplainAsserter().WithDocFetches(2).WithFieldFetches(3).WithIndexFetches(1),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToOnesSecondaryRelation_IfFilterOnIndexedRelation_ShouldFilter(t *testing.T) {
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
			sendRequestAndExplain(`
				User(filter: {
					address: {city: {_eq: "Munich"}}
				}) {
					name
				}`,
				[]map[string]any{
					{"name": "Islam"},
				},
				testUtils.NewExplainAsserter().WithDocFetches(2).WithFieldFetches(3).WithIndexFetches(1),
			),
			sendRequestAndExplain(`
				User(filter: {
					address: {city: {_eq: "Montreal"}}
				}) {
					name
				}`,
				[]map[string]any{
					{"name": "Shahzad"},
					{"name": "Fred"},
					{"name": "John"},
				},
				testUtils.NewExplainAsserter().WithDocFetches(6).WithFieldFetches(9).WithIndexFetches(3),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToOnePrimaryRelation_IfFilterOnIndexedRelation_ShouldFilter(t *testing.T) {
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
			sendRequestAndExplain(`
				User(filter: {
					address: {city: {_eq: "Munich"}}
				}) {
					name
				}`,
				[]map[string]any{
					{"name": "Islam"},
				},
				testUtils.NewExplainAsserter().WithDocFetches(2).WithFieldFetches(3).WithIndexFetches(1),
			),
			sendRequestAndExplain(`
				User(filter: {
					address: {city: {_eq: "Montreal"}}
				}) {
					name
				}`,
				[]map[string]any{
					{"name": "John"},
					{"name": "Fred"},
					{"name": "Shahzad"},
				},
				testUtils.NewExplainAsserter().WithDocFetches(14).WithFieldFetches(17).WithIndexFetches(3),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToTwoRelation_IfFilterOnIndexedRelation_ShouldFilter(t *testing.T) {
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
			sendRequestAndExplain(`
				User(filter: {
					address: {city: {_eq: "Munich"}}
				}) {
					name
					address {
						city
					}
				}`,
				[]map[string]any{
					{
						"name": "Islam",
						"address": map[string]any{
							"city": "Munich",
						},
					},
				},
				testUtils.NewExplainAsserter().WithDocFetches(2).WithFieldFetches(3).WithIndexFetches(1),
			),
			sendRequestAndExplain(`
				User(filter: {
					devices: {model: {_eq: "Walkman"}}
				}) {
					name
					devices {
						model
					}
				}`,
				[]map[string]any{
					{
						"name": "Chris",
						"devices": map[string]any{
							"model": "Walkman",
						},
					},
				},
				testUtils.NewExplainAsserter().WithDocFetches(2).WithFieldFetches(3).WithIndexFetches(1),
			),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
