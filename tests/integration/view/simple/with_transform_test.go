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

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestView_SimpleWithTransform(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.AddView{
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
			&action.AddDoc{
				// Set the `name` field only
				Doc: `{
					"name":	"John"
				}`,
			},
			&action.AddDoc{
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
			&action.AddCollection{
				SDL: `
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
			&action.AddView{
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
			&action.AddCollection{
				SDL: `
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
				TransformCID: immutable.Some("{{.LensID0}}"),
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
			&action.AddCollection{
				SDL: `
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
			&action.AddView{
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
			&action.AddDoc{
				Doc: `{
					"name":	"John",
					"valid": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"Fred",
					"valid": false
				}`,
			},
			&action.AddDoc{
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
