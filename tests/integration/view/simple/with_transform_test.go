// Copyright 2024 Democratized Data Foundation
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

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestView_SimpleWithTransform(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple view with transform",
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
						fullName: String
					}
				`,
				Transform: immutable.Some(model.Lens{
					// This transform will copy the value from `name` into the `fullName` field,
					// like an overly-complicated alias
					Lenses: []model.LensModule{
						{
							Path: lenses.CopyModulePath,
							Arguments: map[string]any{
								"src": "name",
								"dst": "fullName",
							},
						},
					},
				}),
			},
			testUtils.CreateDoc{
				// Set the `name` field only
				Doc: `{
					"name":	"John"
				}`,
			},
			testUtils.CreateDoc{
				// Set the `name` field only
				Doc: `{
					"name":	"Fred"
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						UserView {
							fullName
						}
					}
				`,
				Results: []map[string]any{
					{
						"fullName": "Fred",
					},
					{
						"fullName": "John",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_SimpleWithMultipleTransforms(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple view with multiple transforms",
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
						fullName: String
						age: Int
					}
				`,
				Transform: immutable.Some(model.Lens{
					// This transform will copy the value from `name` into the `fullName` field,
					// like an overly-complicated alias.  It will then set `age` to 23.
					//
					// It is important that this test tests the returning of more fields than it is
					// provided with, given the production code.
					Lenses: []model.LensModule{
						{
							Path: lenses.CopyModulePath,
							Arguments: map[string]any{
								"src": "name",
								"dst": "fullName",
							},
						},
						{
							Path: lenses.SetDefaultModulePath,
							Arguments: map[string]any{
								"dst":   "age",
								"value": 23,
							},
						},
					},
				}),
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
				Request: `
					query {
						UserView {
							fullName
							age
						}
					}
				`,
				Results: []map[string]any{
					{
						"fullName": "Fred",
						"age":      23,
					},
					{
						"fullName": "John",
						"age":      23,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_SimpleWithTransformReturningMoreDocsThanInput(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple view with transform returning more docs than input",
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
				Transform: immutable.Some(model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.PrependModulePath,
							Arguments: map[string]any{
								"values": []map[string]any{
									{
										"name": "Fred",
									},
									{
										"name": "Shahzad",
									},
								},
							},
						},
					},
				}),
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
							name
						}
					}
				`,
				Results: []map[string]any{
					{
						"name": "Fred",
					},
					{
						"name": "Shahzad",
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

func TestView_SimpleWithTransformReturningFewerDocsThanInput(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple view with transform returning fewer docs than input",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						valid: Boolean
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					User {
						name
						valid
					}
				`,
				SDL: `
					type UserView {
						name: String
					}
				`,
				Transform: immutable.Some(model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.FilterModulePath,
							Arguments: map[string]any{
								"src":   "valid",
								"value": true,
							},
						},
					},
				}),
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"John",
					"valid": true
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"Fred",
					"valid": false
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"Shahzad",
					"valid": true
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						UserView {
							name
						}
					}
				`,
				Results: []map[string]any{
					{
						"name": "Shahzad",
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
