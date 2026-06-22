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

// These tests cover the epoch-namespaced index rebuild of issue #4949 along dimensions the
// existing with_index_test.go matrix does not: a unique index whose migrated values would
// false-collide with the superseded epoch under a shared keyspace, a collection with more than one
// index so the rebuild loop runs per index, and recovery of a rebuild across a process restart.

package query

import (
	"testing"

	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

// TestCollectionMigrationQuery_WithUniqueIndexOnMigratedField_RebuildsWithoutFalseCollision is the
// unique-index-across-rebuild regression. The migration increments a unique-indexed `age` by 10, so
// the first doc's migrated value (20) equals the second doc's pre-migration value (20). Under a
// shared keyspace the rebuild would check the new value against the old entry and false-collide;
// with epoch namespacing the new epoch's uniqueness is independent of the superseded epoch, so the
// rebuild succeeds and both docs are queryable by their migrated values.
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

// TestCollectionMigrationQuery_WithMultipleIndexes_RebuildsAll covers the per-index loop in
// reindexNewActiveVersion: a collection with two indexes is switched to a migrated version, and both
// indexes must be rebuilt into their new epochs and usable. The existing rebuild tests all use a
// single index, so the loop body only ever ran once.
func TestCollectionMigrationQuery_WithMultipleIndexes_RebuildsAll(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String @index
						age: Int @index
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{"name": "Andy", "age": 20},
			},
			&action.AddDoc{
				DocMap: map[string]any{"name": "Beth", "age": 30},
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
									"value": 5,
								},
							},
						},
					},
				},
			},
			// The name index (not migrated) is rebuilt into its new epoch and still resolves.
			&action.Request{
				Request: `query {
					Users(filter: {name: {_eq: "Andy"}}) {
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{{"age": int64(25)}},
				},
			},
			&action.Request{
				Request: `query @explain(type: execute) {
					Users(filter: {name: {_eq: "Andy"}}) {
						age
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
			// The age index (migrated) is rebuilt into its new epoch and resolves migrated values.
			&action.Request{
				Request: `query {
					Users(filter: {age: {_eq: 35}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{{"name": "Beth"}},
				},
			},
			&action.Request{
				Request: `query @explain(type: execute) {
					Users(filter: {age: {_eq: 35}}) {
						name
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestCollectionMigrationQuery_WithIndexRebuildAndRestart_RecoversAndQueries covers crash-recovery
// of a rebuild across a restart. After the migration triggers a rebuild, the node is restarted;
// startup recovery must leave the index complete on its new epoch (resuming any unfinished build and
// collecting the superseded epoch), and the post-restart query must use the index with migrated
// values.
func TestCollectionMigrationQuery_WithIndexRebuildAndRestart_RecoversAndQueries(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int @index
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{"name": "Andy", "age": 20},
			},
			&action.AddDoc{
				DocMap: map[string]any{"name": "Beth", "age": 30},
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
									"value": 5,
								},
							},
						},
					},
				},
			},
			testUtils.Restart{},
			&action.Request{
				Request: `query {
					Users(filter: {age: {_eq: 35}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{{"name": "Beth"}},
				},
			},
			&action.Request{
				Request: `query @explain(type: execute) {
					Users(filter: {age: {_eq: 35}}) {
						name
					}
				}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
