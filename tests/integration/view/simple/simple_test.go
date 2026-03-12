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

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestView_Simple(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddView{
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView @materialized(if: false) {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			&action.Request{
				Request: `query {
							UserView {
								name
							}
						}`,
				Results: map[string]any{
					"UserView": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_SimpleMultipleDocs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddView{
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView @materialized(if: false) {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"Fred"
				}`,
			},
			&action.Request{
				Request: `query {
							UserView {
								name
							}
						}`,
				Results: map[string]any{
					"UserView": []map[string]any{
						{
							"name": "John",
						},
						{
							"name": "Fred",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_SimpleWithFieldSubset_ErrorsSelectingExcludedField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.AddView{
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView @materialized(if: false) {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			&action.Request{
				Request: `
					query {
						UserView {
							age
						}
					}
				`,
				ExpectedError: `Cannot query field "age" on type "UserView"`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_SimpleWithExtraFieldInViewSDL(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.AddView{
				Query: `
					User {
						name
					}
				`,
				// `age` is present in SDL but not the query
				SDL: `
					type UserView @materialized(if: false) {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			&action.Request{
				Request: `query {
							UserView {
								name
							}
						}`,
				Results: map[string]any{
					"UserView": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_SimpleWithExtraFieldInViewQuery(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.AddView{
				// `age` is present in the query but not the SDL
				Query: `
					User {
						name
						age
					}
				`,
				SDL: `
					type UserView @materialized(if: false) {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			&action.Request{
				Request: `
					query {
						UserView {
							name
						}
					}
				`,
				Results: map[string]any{
					"UserView": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_SimpleViewOfView(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddView{
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView @materialized(if: false) {
						name: String
					}
				`,
			},
			&action.AddView{
				Query: `
					UserView {
						name
					}
				`,
				SDL: `
					type UserViewView @materialized(if: false) {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			&action.Request{
				Request: `
					query {
						UserViewView {
							name
						}
					}
				`,
				Results: map[string]any{
					"UserViewView": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
