// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			testUtils.AddPredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "Keenan"},
						{"name": "Islam"},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(3),
			},
			&action.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Addo"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(1),
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			testUtils.AddPredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(1),
			},
			&action.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "John"},
						{"name": "Fred"},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(3),
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
						address: Address @primary @index(unique: true)
					}

					type Address {
						user: User
						city: String @index
						street: String
					}`,
			},
			testUtils.AddPredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req1),
				// 1 index fetch on subType (Address city index) + 1 index fetch on root (User _addressID index)
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(1).
					WithLevel("root").WithIndexFetches(1),
			},
			&action.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "John"},
						{"name": "Fred"},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req2),
				// 3 index fetches on subType (Address city index) + 3 index fetches on root (User _addressID index)
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(3).
					WithLevel("root").WithIndexFetches(3),
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			testUtils.AddPredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req1),
				// subType: 2 fieldFetches, 1 indexFetch (Address city index)
				// root: 3 fieldFetches, 1 indexFetch (User _addressID index)
				Asserter: testUtils.NewExplainAsserter("subType").WithFieldFetches(2).WithIndexFetches(1).
					WithLevel("root").WithFieldFetches(3).WithIndexFetches(1),
			},
			&action.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "John"},
						{"name": "Fred"},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req2),
				// subType: 6 fieldFetches, 3 indexFetches (Address city index)
				// root: 9 fieldFetches, 3 indexFetches (User _addressID index)
				Asserter: testUtils.NewExplainAsserter("subType").WithFieldFetches(6).WithIndexFetches(3).
					WithLevel("root").WithFieldFetches(9).WithIndexFetches(3),
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
						address: Address @primary @index(unique: true)
					}

					type Address {
						user: User
						city: String @index
						street: String
					}`,
			},
			testUtils.AddPredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// subType: 1 indexFetch (Address city index), root: 1 indexFetch (User _addressID index)
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(1).
					WithLevel("root").WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToMany_IfFilterOnIndexedPrimaryDoc_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Chris"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Addo"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "The Proclaimers",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Running Man",
					"manufacturer": "Braveworld Productions",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
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
								// The filter is on User, so all devices belonging to it will be returned
								{
									"model":        "Running Man",
									"manufacturer": "Braveworld Productions",
								},
								{
									"model":        "Walkman",
									"manufacturer": "The Proclaimers",
								},
								{
									"model":        "Walkman",
									"manufacturer": "Sony",
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToMany_IfFilterOnIndexedPrimaryDocAndSubFilter_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Chris"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Addo"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "The Proclaimers",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Running Man",
					"manufacturer": "Braveworld Productions",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: `query {
					User(filter: {
						devices: {model: {_eq: "Walkman"}}
					}) {
						name
						devices(filter: {manufacturer: {_neq: "Sony"}}) {
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
									"model":        "Running Man",
									"manufacturer": "Braveworld Productions",
								},
								{
									"model":        "Walkman",
									"manufacturer": "The Proclaimers",
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Chris"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "The Proclaimers",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Running Man",
					"manufacturer": "Braveworld Productions",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Chris",
							"devices": []map[string]any{
								{
									"model":        "Running Man",
									"manufacturer": "Braveworld Productions",
								},
								{
									"model":        "Walkman",
									"manufacturer": "The Proclaimers",
								},
								{
									"model":        "Walkman",
									"manufacturer": "Sony",
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(2),
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			testUtils.AddPredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
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
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(1),
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			testUtils.AddPredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"Device": []map[string]any{
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
						{
							"model": "Playstation 5",
							"owner": map[string]any{
								"name": "Addo",
							},
						},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// 3 index fetches on root (Device year index) to get devices with year 2021
				Asserter: testUtils.NewExplainAsserter("root").WithIndexFetches(3),
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			testUtils.AddPredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"Device": []map[string]any{
						{"model": "MacBook Pro"},
						{"model": "iPad Mini"},
						{"model": "iPhone 13"},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// subType: 1 indexFetch (get owner by name)
				// root: 3 indexFetches (get devices by _ownerID)
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(1).
					WithLevel("root").WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToMany_IfIndexedRelationIsNil_NeNilFilterShouldUseIndex(t *testing.T) {
	req := `query {
		Device(filter: {
			_ownerID: {_neq: null}
		}) {
			model
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Chris"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "iPhone",
					"manufacturer": "Apple",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"model":	"Running Man",
					"manufacturer": "Braveworld Productions"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"model":	"PlayStation 5",
					"manufacturer": "Sony"
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"Device": []map[string]any{
						{"model": "Walkman"},
						{"model": "iPhone"},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// 4 index fetches on root (Device _ownerID index) to find devices with _ownerID != null
				Asserter: testUtils.NewExplainAsserter("root").WithIndexFetches(4),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndexOnOneToMany_IfIndexedRelationIsNil_EqNilFilterShouldUseIndex(t *testing.T) {
	req := `query {
		Device(filter: {
			_ownerID: {_eq: null}
		}) {
			model
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Chris"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Walkman",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "iPhone",
					"manufacturer": "Apple",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"model":	"Running Man",
					"manufacturer": "Braveworld Productions"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"model":	"PlayStation 5",
					"manufacturer": "Sony"
				}`,
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"Device": []map[string]any{
						{"model": "Running Man"},
						{"model": "PlayStation 5"},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// 2 index fetches on root (Device _ownerID index) to get devices with _ownerID == null
				Asserter: testUtils.NewExplainAsserter("root").WithIndexFetches(2),
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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name": "Apple",
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "MacBook Pro",
					"owner":        testUtils.NewDocIndex(0, 0),
					"manufacturer": testUtils.NewDocIndex(2, 0),
				},
			},
			&action.Request{
				Request: `query {
					User {
						devices {
							_ownerID
							_manufacturerID
						}
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"devices": []map[string]any{
								{
									"_ownerID":        testUtils.NewDocIndex(0, 0),
									"_manufacturerID": testUtils.NewDocIndex(2, 0),
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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Request{
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

func TestQueryWithIndex_WithScalarAndRelationFilterAtTopLevel_ShouldApplyBothAsAnd(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Author {
						name: String @index
						published: [Book]
					}

					type Book {
						title: String
						rating: Float
						author: Author
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Cornelia Funke"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Painted House",
					"rating": 4.9,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "A Time to Kill",
					"rating": 4.0,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Theif Lord",
					"rating": 4.8,
					"author": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Inkheart",
					"rating": 4.4,
					"author": testUtils.NewDocIndex(0, 1),
				},
			},
			// Implicit AND: scalar + relation conditions at top level
			&action.Request{
				Request: `query {
					Book(filter: {rating: {_gt: 4.5}, author: {name: {_eq: "John Grisham"}}}) {
						title
						rating
						author {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"title":  "Painted House",
							"rating": 4.9,
							"author": map[string]any{
								"name": "John Grisham",
							},
						},
					},
				},
				NonOrderedResults: true,
			},
			// Explicit _and: same conditions, should produce identical results
			&action.Request{
				Request: `query {
					Book(filter: {_and: [{rating: {_gt: 4.5}}, {author: {name: {_eq: "John Grisham"}}}]}) {
						title
						rating
						author {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"title":  "Painted House",
							"rating": 4.9,
							"author": map[string]any{
								"name": "John Grisham",
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_WithMultipleScalarsAndRelationFilter_ShouldApplyAllAsAnd(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Author {
						name: String @index
						published: [Book]
					}

					type Book {
						title: String
						rating: Float
						genre: String
						author: Author
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Cornelia Funke"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Painted House",
					"rating": 4.9,
					"genre":  "drama",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "A Time to Kill",
					"rating": 4.0,
					"genre":  "thriller",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "The Firm",
					"rating": 4.5,
					"genre":  "thriller",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Theif Lord",
					"rating": 4.8,
					"genre":  "fantasy",
					"author": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.Request{
				Request: `query {
					Book(filter: {genre: {_eq: "thriller"}, rating: {_gt: 4.0}, author: {name: {_eq: "John Grisham"}}}) {
						title
						rating
						genre
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"title":  "The Firm",
							"rating": 4.5,
							"genre":  "thriller",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
