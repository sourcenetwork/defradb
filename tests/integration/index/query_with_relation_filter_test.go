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
	// 3 users have a MacBook Pro: Islam, Shahzad, Keenan
	req1 := `query {
		User(filter: {
			devices: {model: {_eq: "MacBook Pro"}}
		}) {
			name
		}
	}`
	// 1 user has an iPhone 10: Addo
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
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int
						devices: [Device] 
					} 

					type Device {
						model: String @index
						owner: User
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "Islam"},
						{"name": "Keenan"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
			testUtils.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Addo"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToOnesSecondaryRelation_IfFilterOnIndexedRelation_ShouldFilter(t *testing.T) {
	// 1 user lives in Munich: Islam
	req1 := `query {
		User(filter: {
			address: {city: {_eq: "Munich"}}
		}) {
			name
		}
	}`
	// 3 users live in Montreal: Shahzad, Fred, John
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
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int
						address: Address
					} 

					type Address {
						user: User @primary
						city: String @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
			testUtils.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "John"},
						{"name": "Shahzad"},
						{"name": "Fred"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToOnePrimaryRelation_IfFilterOnIndexedFieldOfRelationAndRelation_ShouldFilter(t *testing.T) {
	// 1 user lives in London: Andy
	req1 := `query {
		User(filter: {
			address: {city: {_eq: "London"}}
		}) {
			name
		}
	}`
	// 3 users live in Montreal: Shahzad, Fred, John
	req2 := `query {
		User(filter: {
			address: {city: {_eq: "Montreal"}}
		}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Filter on indexed field of primary relation in 1-1 relation",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int
						address: Address @primary @index
					} 

					type Address {
						user: User
						city: String @index
						street: String 
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req1),
				// we make 2 index fetches: 1. to get the only address with city == "London"
				// and 2. to get the corresponding user
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
			testUtils.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "John"},
						{"name": "Fred"},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req2),
				// we make 3 index fetches to get the 3 address with city == "Montreal"
				// and 3 more index fetches to get the corresponding users
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(6),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToOnePrimaryRelation_IfFilterOnIndexedFieldOfRelation_ShouldFilter(t *testing.T) {
	// 1 user lives in London: Andy
	req1 := `query {
		User(filter: {
			address: {city: {_eq: "London"}}
		}) {
			name
		}
	}`
	// 3 users live in Montreal: Shahzad, Fred, John
	req2 := `query {
		User(filter: {
			address: {city: {_eq: "Montreal"}}
		}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Filter on indexed field of primary relation in 1-1 relation",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int
						address: Address @primary 
					} 

					type Address {
						user: User
						city: String @index
						street: String 
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req1),
				// we make 1 index fetch to get the only address with city == "London"
				// we fetch 2 fields for Address doc: "city" and "street"
				// then we scan all 10 users to find one with matching "address_id"
				// for each of User docs we fetch 3 fields: "name", "age" and "address_id"
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(32).WithIndexFetches(1),
			},
			testUtils.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "John"},
						{"name": "Fred"},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req2),
				// we make 3 index fetch to get the 3 address with city == "Montreal"
				// then we scan all 10 users to find one with matching "address_id" for each address
				// after this we fetch the name of each user
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToOnePrimaryRelation_IfFilterOnIndexedRelationWhileIndexedForeignField_ShouldFilter(t *testing.T) {
	// 1 user lives in London: Andy
	req := `query {
		User(filter: {
			address: {city: {_eq: "London"}}
		}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Filter on indexed field of primary relation while having indexed foreign field in 1-1 relation",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int
						address: Address @primary @index
					} 

					type Address {
						user: User
						city: String @index
						street: String 
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToMany_IfFilterOnIndexedPrimaryDoc_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Filter on indexed relation field in 1-N relations",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						devices: [Device]
					}

					type Device {
						model: String @index
						manufacturer: String
						owner: User
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Chris"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Addo"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "The Proclaimers",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Running Man",
					"manufacturer": "Braveworld Productions",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.Request{
				Request: `query {
					User(filter: {
						devices: {model: {_eq: "Walkman"}}
					}) {
						name
						devices {
							model
							manufacturer
						}
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Chris",
							"devices": []map[string]any{
								{
									"model":        "Walkman",
									"manufacturer": "Sony",
								},
								{
									"model":        "Walkman",
									"manufacturer": "The Proclaimers",
								},
								// The filter is on User, so all devices belonging to it will be returned
								{
									"model":        "Running Man",
									"manufacturer": "Braveworld Productions",
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

func TestQueryWithIndexOnOneToMany_IfFilterOnIndexedPrimaryDocAndSubFilter_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Filter on indexed relation field in 1-N relations",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						devices: [Device]
					}

					type Device {
						model: String @index
						manufacturer: String
						owner: User
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Chris"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Addo"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "The Proclaimers",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Running Man",
					"manufacturer": "Braveworld Productions",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.Request{
				Request: `query {
					User(filter: {
						devices: {model: {_eq: "Walkman"}}
					}) {
						name
						devices(filter: {manufacturer: {_ne: "Sony"}}) {
							model
							manufacturer
						}
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Chris",
							"devices": []map[string]any{
								{
									"model":        "Walkman",
									"manufacturer": "The Proclaimers",
								},
								{
									"model":        "Running Man",
									"manufacturer": "Braveworld Productions",
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

func TestQueryWithIndexOnOneToMany_IfFilterOnIndexedRelation_ShouldFilterWithExplain(t *testing.T) {
	req := `query {
		User(filter: {
			devices: {model: {_eq: "Walkman"}}
		}) {
			name
			devices {
				model
				manufacturer
			}
		}
	}`
	test := testUtils.TestCase{
		Description: "Filter on indexed relation field in 1-N relations",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						devices: [Device]
					} 

					type Device {
						model: String @index
						manufacturer: String
						owner: User
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Chris"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "The Proclaimers",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Running Man",
					"manufacturer": "Braveworld Productions",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Chris",
							"devices": []map[string]any{
								{
									"model":        "Walkman",
									"manufacturer": "Sony",
								},
								{
									"model":        "Walkman",
									"manufacturer": "The Proclaimers",
								},
								{
									"model":        "Running Man",
									"manufacturer": "Braveworld Productions",
								},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToOne_IfFilterOnIndexedRelation_ShouldFilter(t *testing.T) {
	// 1 user lives in Munich: Islam
	req := `query {
		User(filter: {
			address: {city: {_eq: "Munich"}}
		}) {
			name
			address {
				city
			}
		}
	}`
	test := testUtils.TestCase{
		Description: "Filter on indexed relation field in 1-1 relation",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						address: Address
					}

					type Address {
						user: User @primary
						city: String @index
					}
				`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Islam",
							"address": map[string]any{
								"city": "Munich",
							},
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnManyToOne_IfFilterOnIndexedField_ShouldFilterWithExplain(t *testing.T) {
	// This query will fetch first a matching device which is primary doc and therefore
	// has a reference to the secondary User doc.
	req := `query {
		Device(filter: {
			year: {_eq: 2021}
		}) {
			model
			owner {
				name
			}
		}
	}`
	test := testUtils.TestCase{
		Description: "With filter on indexed field of secondary relation (N-1) should fetch secondary and primary objects",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						devices: [Device]
					} 

					type Device {
						model: String 
						year: Int @index
						owner: User
					}
				`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"Device": []map[string]any{
						{
							"model": "Playstation 5",
							"owner": map[string]any{
								"name": "Addo",
							},
						},
						{
							"model": "iPhone 10",
							"owner": map[string]any{
								"name": "Addo",
							},
						},
						{
							"model": "Playstation 5",
							"owner": map[string]any{
								"name": "Islam",
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// we make 3 index fetches to get all 3 devices with year 2021
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnManyToOne_IfFilterOnIndexedRelation_ShouldFilterWithExplain(t *testing.T) {
	// This query will fetch first a matching user (owner) which is primary doc and therefore
	// has no direct reference to secondary Device docs.
	// Keenan has 3 devices.
	req := `query {
		Device(filter: {
			owner: {name: {_eq: "Keenan"}}
		}) {
			model
		}
	}`
	test := testUtils.TestCase{
		Description: "Upon querying secondary object with filter on indexed field of primary relation (in 1-N) should fetch all secondary objects of the same primary one",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String @index
						devices: [Device]
					}

					type Device {
						model: String
						owner: User @index
					}
				`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"Device": []map[string]any{
						{"model": "MacBook Pro"},
						{"model": "iPad Mini"},
						{"model": "iPhone 13"},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// we make 1 index fetch to get the owner by it's name
				// and 3 index fetches to get all 3 devices of the owner
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(4),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToMany_IfIndexedRelationIsNil_NeNilFilterShouldUseIndex(t *testing.T) {
	req := `query {
		Device(filter: {
			owner_id: {_ne: null}
		}) {
			model
		}
	}`
	test := testUtils.TestCase{
		Description: "Filter on indexed relation field in 1-N relations",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						devices: [Device]
					}

					type Device {
						model: String 
						manufacturer: String
						owner: User @index
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Chris"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "iPhone",
					"manufacturer": "Apple",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"model":	"Running Man",
					"manufacturer": "Braveworld Productions"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"model":	"PlayStation 5",
					"manufacturer": "Sony"
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"Device": []map[string]any{
						{"model": "Walkman"},
						{"model": "iPhone"},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// we make 4 index fetches to find 2 devices with owner_id != null
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(4),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToMany_IfIndexedRelationIsNil_EqNilFilterShouldUseIndex(t *testing.T) {
	req := `query {
		Device(filter: {
			owner_id: {_eq: null}
		}) {
			model
		}
	}`
	test := testUtils.TestCase{
		Description: "Filter on indexed relation field in 1-N relations",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						devices: [Device]
					}

					type Device {
						model: String 
						manufacturer: String
						owner: User @index
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Chris"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "iPhone",
					"manufacturer": "Apple",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"model":	"Running Man",
					"manufacturer": "Braveworld Productions"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"model":	"PlayStation 5",
					"manufacturer": "Sony"
				}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"Device": []map[string]any{
						{"model": "PlayStation 5"},
						{"model": "Running Man"},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// we make 2 index fetches to get all 2 devices with owner_id == null
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test was added during https://github.com/sourcenetwork/defradb/issues/2862
// multiple indexed fields on the second object are required for the failure.
func TestQueryWithIndexOnManyToOne_MultipleViaOneToMany(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						devices: [Device]
					}

					type Device {
						model: String
						owner: User @index
						manufacturer: Manufacturer @index
					}

					type Manufacturer {
						name: String
						devices: [Device]
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name": "Apple",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "MacBook Pro",
					"owner":        testUtils.NewDocIndex(0, 0),
					"manufacturer": testUtils.NewDocIndex(2, 0),
				},
			},
			testUtils.Request{
				Request: `query {
					User {
						devices {
							owner_id
							manufacturer_id
						}
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"devices": []map[string]any{
								{
									"owner_id":        testUtils.NewDocIndex(0, 0),
									"manufacturer_id": testUtils.NewDocIndex(2, 0),
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

func TestQueryWithUniqueIndex_WithFilterOnChildIndexedField_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String @index(unique: true)
						devices: [Device]
					}

					type Device {
						trusted: Boolean
						owner: User
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.Request{
				Request: `query {
					Device(filter: {owner: {name: {_eq: "John"}}}) {
						trusted
					}
				}`,
				Results: map[string]any{
					"Device": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
