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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/cli/test/action"
	"github.com/sourcenetwork/defradb/cli/test/integration"
	"github.com/sourcenetwork/defradb/client"
)

func TestIndexCreate_WithSingleField_ShouldSucceed(t *testing.T) {
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
				Expected: immutable.Some(client.IndexDescription{
					Name: "UsersByName",
					Fields: []client.IndexedFieldDescription{
						{Name: "name", Descending: false},
					},
					Unique: false,
				}),
			},
		},
	}

	test.Execute(t)
}

func TestIndexCreate_WithMultipleFieldsAndOrders_ShouldSucceed(t *testing.T) {
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
				Name:       "UsersByNameAndAge",
				Fields:     []string{"name:ASC", "age:DESC"},
				Expected: immutable.Some(client.IndexDescription{
					Name: "UsersByNameAndAge",
					Fields: []client.IndexedFieldDescription{
						{Name: "name", Descending: false},
						{Name: "age", Descending: true},
					},
					Unique: false,
				}),
			},
		},
	}

	test.Execute(t)
}

func TestIndexCreate_WithUniqueFlag_ShouldCreateUniqueIndex(t *testing.T) {
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
				Name:       "UniqueEmail",
				Fields:     []string{"email"},
				Unique:     true,
				Expected: immutable.Some(client.IndexDescription{
					Name: "UniqueEmail",
					Fields: []client.IndexedFieldDescription{
						{Name: "email", Descending: false},
					},
					Unique: true,
				}),
			},
		},
	}

	test.Execute(t)
}

func TestIndexCreate_WithoutName_ShouldGenerateName(t *testing.T) {
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
				Fields:     []string{"age"},
				Expected: immutable.Some(client.IndexDescription{
					// Name will be auto-generated, so we don't check it
					Fields: []client.IndexedFieldDescription{
						{Name: "age", Descending: false},
					},
					Unique: false,
				}),
			},
		},
	}

	test.Execute(t)
}

func TestIndexCreate_WithUnknownCollection_ShouldReturnError(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.IndexCreate{
				Collection:  "NonExistentCollection",
				Name:        "TestIndex",
				Fields:      []string{"field1"},
				ExpectError: "collection not found",
			},
		},
	}

	test.Execute(t)
}

func TestIndexCreate_WithoutCollection_ShouldReturnError(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.IndexCreate{
				// Collection is empty
				Name:        "TestIndex",
				Fields:      []string{"field1"},
				ExpectError: "collection not found",
			},
		},
	}

	test.Execute(t)
}

func TestIndexCreate_WithoutFields_ShouldReturnError(t *testing.T) {
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
			&action.IndexCreate{
				Collection: "User",
				Name:       "EmptyIndex",
				// Fields is empty
				ExpectError: "index missing fields",
			},
		},
	}

	test.Execute(t)
}

func TestIndexCreate_WithInvalidFieldOrder_ShouldReturnError(t *testing.T) {
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
			&action.IndexCreate{
				Collection:  "User",
				Name:        "InvalidOrderIndex",
				Fields:      []string{"name:INVALID"},
				ExpectError: "invalid order: expected ASC or DESC",
			},
		},
	}

	test.Execute(t)
}

func TestIndexCreate_WithNonExistentField_ShouldReturnError(t *testing.T) {
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
			&action.IndexCreate{
				Collection:  "User",
				Name:        "InvalidFieldIndex",
				Fields:      []string{"nonexistent"},
				ExpectError: "creating an index on a non-existing property",
			},
		},
	}

	test.Execute(t)
}

func TestIndexCreate_WithDuplicateName_ShouldReturnError(t *testing.T) {
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
			&action.IndexCreate{
				Collection: "User",
				Name:       "DuplicateIndex",
				Fields:     []string{"name"},
			},
			&action.IndexCreate{
				Collection:  "User",
				Name:        "DuplicateIndex",
				Fields:      []string{"age"},
				ExpectError: "already exists",
			},
		},
	}

	test.Execute(t)
}
