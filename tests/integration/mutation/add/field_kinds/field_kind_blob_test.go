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

func TestMutationAddFieldKinds_WithBlob(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						data: Blob
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"data": "00FF",
				},
			},
			&action.Request{
				Request: `query {
					User {
						data
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
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

func TestMutationAddFieldKinds_WithNillableBlob_Nil(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						data: Blob
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"data": nil,
				},
			},
			&action.Request{
				Request: `query {
					User {
						data
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
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

func TestMutationAddFieldKinds_WithNonNillableBlob(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						data: Blob!
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"data": "00FF",
				},
			},
			&action.Request{
				Request: `query {
					User {
						data
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
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
