// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/cli/test/action"
	"github.com/sourcenetwork/defradb/cli/test/integration"
	"github.com/sourcenetwork/defradb/client"
)

func TestIndexDrop_WithExistingIndex_ShouldSucceed(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.SchemaAdd{
				InlineSchema: `
					type User {
						name: String
						age: Int
						email: String
					}
				`,
			},
			&action.IndexCreate{
				Collection: "User",
				Name:       "UsersByName",
				Fields:     []string{"name"},
			},
			&action.IndexList{
				Collection: "User",
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "UsersByName",
						Fields: []client.IndexedFieldDescription{
							{Name: "name", Descending: false},
						},
						Unique: false,
					},
				},
			},
			&action.IndexDrop{
				Collection: "User",
				Name:       "UsersByName",
			},
			&action.IndexList{
				Collection:      "User",
				ExpectedIndexes: []client.IndexDescription{},
			},
		},
	}

	test.Execute(t)
}

func TestIndexDrop_WithUnknownCollection_ShouldReturnError(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.IndexDrop{
				Collection:  "NonExistentCollection",
				Name:        "SomeIndex",
				ExpectError: "collection not found",
			},
		},
	}

	test.Execute(t)
}

func TestIndexDrop_WithoutCollection_ShouldReturnError(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.IndexDrop{
				// Collection is empty
				Name:        "SomeIndex",
				ExpectError: "error expected",
			},
		},
	}

	test.Execute(t)
}

func TestIndexDrop_WithoutName_ShouldReturnError(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.SchemaAdd{
				InlineSchema: `
					type User {
						name: String
					}
				`,
			},
			&action.IndexDrop{
				Collection: "User",
				// Name is empty
				ExpectError: "error expected",
			},
		},
	}

	test.Execute(t)
}

func TestIndexDrop_WithNonExistentIndex_ShouldReturnError(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.SchemaAdd{
				InlineSchema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			&action.IndexDrop{
				Collection:  "User",
				Name:        "NonExistentIndex",
				ExpectError: "index not found",
			},
		},
	}

	test.Execute(t)
}

func TestIndexDrop_WithMultipleIndexes_ShouldDropOnlySpecified(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.SchemaAdd{
				InlineSchema: `
					type User {
						name: String
						age: Int
						email: String
					}
				`,
			},
			// Create multiple indexes
			&action.IndexCreate{
				Collection: "User",
				Name:       "UsersByName",
				Fields:     []string{"name"},
			},
			&action.IndexCreate{
				Collection: "User",
				Name:       "UsersByAge",
				Fields:     []string{"age"},
			},
			&action.IndexCreate{
				Collection: "User",
				Name:       "UsersByEmail",
				Fields:     []string{"email"},
				Unique:     true,
			},
			// Verify all indexes exist
			&action.IndexList{
				Collection: "User",
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "UsersByName",
						Fields: []client.IndexedFieldDescription{
							{Name: "name", Descending: false},
						},
						Unique: false,
					},
					{
						Name: "UsersByAge",
						Fields: []client.IndexedFieldDescription{
							{Name: "age", Descending: false},
						},
						Unique: false,
					},
					{
						Name: "UsersByEmail",
						Fields: []client.IndexedFieldDescription{
							{Name: "email", Descending: false},
						},
						Unique: true,
					},
				},
			},
			// Drop one index
			&action.IndexDrop{
				Collection: "User",
				Name:       "UsersByAge",
			},
			// Verify only two indexes remain
			&action.IndexList{
				Collection: "User",
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "UsersByName",
						Fields: []client.IndexedFieldDescription{
							{Name: "name", Descending: false},
						},
						Unique: false,
					},
					{
						Name: "UsersByEmail",
						Fields: []client.IndexedFieldDescription{
							{Name: "email", Descending: false},
						},
						Unique: true,
					},
				},
			},
		},
	}

	test.Execute(t)
}
