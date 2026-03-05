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

func TestView_WithTransformCID_CanReuseExistingLens(t *testing.T) {
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
				Doc: `{
					"name": "John"
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
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_WithInvalidTransformCID_ReturnsError(t *testing.T) {
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
				TransformCID:  immutable.Some("bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi"),
				ExpectedError: "lens CID not found",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
