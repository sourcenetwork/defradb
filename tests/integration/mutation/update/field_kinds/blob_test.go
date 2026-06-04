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

package field_kinds

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_WithBlobField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						data: Blob
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"data": "00FE"
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"data": "00FF"
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							data
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"data": "00FF",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_IfBlobFieldSetToNull_ShouldBeNil(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						data: Blob
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"data": "00FE"
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"data": null
				}`,
			},
			&action.Request{
				Request: `
					query {
						Users {
							data
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"data": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
