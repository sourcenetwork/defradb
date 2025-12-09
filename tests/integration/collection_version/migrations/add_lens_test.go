// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package migrations

import (
	"testing"

	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestAddLens_WithSimpleLens_ReturnsLensID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddLens{
				Lens: model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.SetDefaultModulePath,
							Arguments: map[string]any{
								"dst":   "name",
								"value": "Fred",
							},
						},
					},
				},
			},
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/Users/PreviousVersion/Transform",
							"value": "{{.LensID0}}"
						}
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAddLens_WithMultipleLenses_ReturnsUniqueLensIDs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddLens{
				Lens: model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.SetDefaultModulePath,
							Arguments: map[string]any{
								"dst":   "name",
								"value": "John",
							},
						},
					},
				},
			},
			&action.AddLens{
				Lens: model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.SetDefaultModulePath,
							Arguments: map[string]any{
								"dst":   "name",
								"value": "Andy",
							},
						},
					},
				},
			},
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Islam"
				}`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/Users/PreviousVersion/Transform",
							"value": "{{.LensID1}}"
						}
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Andy",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAddLens_WithIdenticalLenses_ReturnsSameCID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddLens{
				Lens: model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.SetDefaultModulePath,
							Arguments: map[string]any{
								"dst":   "name",
								"value": "Fred",
							},
						},
					},
				},
			},
			&action.AddLens{
				Lens: model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.SetDefaultModulePath,
							Arguments: map[string]any{
								"dst":   "name",
								"value": "Fred",
							},
						},
					},
				},
			},
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/Users/PreviousVersion/Transform",
							"value": "{{.LensID1}}"
						}
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
