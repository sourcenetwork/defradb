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

// This covers a migration scenario the existing with_index_test.go matrix does not: a unique index
// whose migrated values overlap the pre-migration values. No other migration test exercises a
// unique index.

package query

import (
	"testing"

	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

// TestCollectionMigrationQuery_WithUniqueIndexOnMigratedField_RebuildsWithoutFalseCollision migrates
// a unique-indexed `age` by +10, so the first doc's migrated value (20) equals the second doc's
// pre-migration value (20). The rebuild must not treat that overlap as a uniqueness violation, and
// both docs must be queryable by their migrated values.
func TestCollectionMigrationQuery_WithUniqueIndexOnMigratedField_RebuildsWithoutFalseCollision(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int @index(unique: true)
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{"name": "Andy", "age": 10},
			},
			&action.AddDoc{
				DocMap: map[string]any{"name": "Beth", "age": 20},
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "{{.CollectionVersionID0}}",
					DestinationCollectionVersionID: "{{.CollectionVersionID1}}",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.IncrementModulePath,
								Arguments: map[string]any{
									"field": "age",
									"value": 10,
								},
							},
						},
					},
				},
			},
			// Andy's migrated age (20) equals Beth's pre-migration age (20); the rebuild must not
			// treat that as a uniqueness violation against the stale epoch.
			&action.Request{
				Request: `query {
					Users(filter: {age: {_eq: 20}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{{"name": "Andy"}},
				},
			},
			&action.Request{
				Request: `query {
					Users(filter: {age: {_eq: 30}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{{"name": "Beth"}},
				},
			},
			&action.Request{
				Request: `query @explain(type: execute) {
					Users(filter: {age: {_eq: 30}}) {
						name
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
