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

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestView_SimpleWithTransform(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			&action.AddLens{
				Lens: model.Lens{
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
				},
			},
			&action.CreateView{
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView @materialized(if: false) {
						fullName: String
					}
				`,
				TransformCID: immutable.Some("{{.LensID0}}"),
			},
			&action.CreateDoc{
				// Set the `name` field only
				Doc: `{
					"name":	"John"
				}`,
			},
			&action.CreateDoc{
				// Set the `name` field only
				Doc: `{
					"name":	"Fred"
				}`,
			},
			&action.Request{
				Request: `
					query {
						UserView {
							fullName
						}
					}
				`,
				Results: map[string]any{
					"UserView": []map[string]any{
						{
							"fullName": "John",
						},
						{
							"fullName": "Fred",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_SimpleWithMultipleTransforms(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			&action.AddLens{
				Lens: model.Lens{
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
				},
			},
			&action.CreateView{
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView @materialized(if: false) {
						fullName: String
						age: Int
					}
				`,
				TransformCID: immutable.Some("{{.LensID0}}"),
			},
			&action.CreateDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name":	"Fred"
				}`,
			},
			&action.Request{
				Request: `
					query {
						UserView {
							fullName
							age
						}
					}
				`,
				Results: map[string]any{
					"UserView": []map[string]any{
						{
							"fullName": "John",
							"age":      23,
						},
						{
							"fullName": "Fred",
							"age":      23,
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_SimpleWithTransformReturningMoreDocsThanInput(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			&action.AddLens{
				Lens: model.Lens{
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
				},
			},
			&action.CreateView{
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
				TransformCID: immutable.Some("{{.LensID0}}"),
			},
			&action.CreateDoc{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_SimpleWithTransformReturningFewerDocsThanInput(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						valid: Boolean
					}
				`,
			},
			&action.AddLens{
				Lens: model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.FilterModulePath,
							Arguments: map[string]any{
								"src":   "valid",
								"value": true,
							},
						},
					},
				},
			},
			&action.CreateView{
				Query: `
					User {
						name
						valid
					}
				`,
				SDL: `
					type UserView @materialized(if: false) {
						name: String
					}
				`,
				TransformCID: immutable.Some("{{.LensID0}}"),
			},
			&action.CreateDoc{
				Doc: `{
					"name":	"John",
					"valid": true
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name":	"Fred",
					"valid": false
				}`,
			},
			&action.CreateDoc{
				Doc: `{
					"name":	"Shahzad",
					"valid": true
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
						{
							"name": "Shahzad",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
