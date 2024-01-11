// Copyright 2023 Democratized Data Foundation
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

func TestView_Simple(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple view",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			testUtils.Request{
				Request: `query {
							UserView {
								name
							}
						}`,
				Results: []map[string]any{
					{
						"name": "John",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_SimpleMultipleDocs(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple view, multiple docs",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"Fred"
				}`,
			},
			testUtils.Request{
				Request: `query {
							UserView {
								name
							}
						}`,
				Results: []map[string]any{
					{
						"name": "Fred",
					},
					{
						"name": "John",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_SimpleWithFieldSubset_ErrorsSelectingExcludedField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple view with field subset errors selecting excluded field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			testUtils.Request{
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
		Description: "Simple view with extra field in SDL",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					User {
						name
					}
				`,
				// `age` is present in SDL but not the query
				SDL: `
					type UserView {
						name: String
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			testUtils.Request{
				Request: `query {
							UserView {
								name
							}
						}`,
				Results: []map[string]any{
					{
						"name": "John",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_SimpleWithExtraFieldInViewQuery(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple view with extra field in view query",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			testUtils.CreateView{
				// `age` is present in the query but not the SDL
				Query: `
					User {
						name
						age
					}
				`,
				SDL: `
					type UserView {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			testUtils.Request{
				Request: `query {
							UserView {
								name
							}
						}`,
				Results: []map[string]any{
					{
						"name": "John",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
