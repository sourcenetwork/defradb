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

package add_collection

import (
	"testing"

	"github.com/sourcenetwork/defradb/cli/test/action"
	"github.com/sourcenetwork/defradb/cli/test/integration"
	"github.com/sourcenetwork/defradb/client"
)

// TestFieldKindStringRepresentation verifies that collection describe outputs
// Kind as human-readable string representations (e.g., "String", "[String]")
// instead of numeric values (e.g., 11, 12). This tests the JSON round-trip
// through the CLI - if Kind was not properly serialized/deserialized as a
// string, the collection creation or describe would fail.
//
// This integration test verifies the round-trip works for multiple Kind types.
// The actual string values are verified by unit tests in client/field_kind_test.go
// (TestFieldKindMarshalJSON, TestFieldKindRoundTrip).
func TestFieldKindStringRepresentation(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.AddCollection{
				InlineSDL: `
					type Book {
						title: String
						tags: [String]
						rating: Float
					}
				`,
			},
			&action.DescribeCollection{
				Expected: []client.CollectionVersion{
					{
						Name:           "Book",
						IsMaterialized: true,
					},
				},
			},
		},
	}

	test.Execute(t)
}
