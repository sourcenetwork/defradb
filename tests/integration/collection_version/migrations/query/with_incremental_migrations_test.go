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

package query

import (
	"testing"

	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

// These tests document the incremental-registration bug tracked in
// https://github.com/sourcenetwork/defradb/issues/4736. They go beyond
// `TestCollectionMigrationQuery_WithMigrationsAcrossMultipleVersions_AppliesAllMigrations` in `simple_test.go`
// (which registers exactly two incremental migrations) by covering sparse
// version chains, three or more incremental registrations, out-of-order
// intermediate registration, and persistence of stale stamps across a
// restart. Cases that require the proper fix (stamp invalidation on chain
// change) are excluded under the secondary-index multiplier.

// TestCollectionMigrationQuery_SparseChainWithIndexAndForwardMigration_ShouldMigrateAndReindexCorrectly
// covers a five-version chain where only the middle link (v3→v4) has a
// migration. An index on a migrated field must be built from the migrated
// values at v4 and remain correct at the active v5. This is distinct from
// `TestCollectionMigrationQuery_ApplyingMigrationBetweenOldVersions_ShouldReindex`
// in that it uses the secondary-index multiplier compatible shape and asserts
// both the migrated value and the index is actually used.
func TestCollectionMigrationQuery_SparseChainWithIndexAndForwardMigration_ShouldMigrateAndReindexCorrectly(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			setupDistantVersions(),
			addMigrationBetweenV3AndV4(),
			&action.Request{
				Request: `query {
					Users(filter: {age: {_eq: 30}}) {
						name
						age
					}
				}`,
				// Fred was 25, +5 = 30
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
							"age":  int64(30),
						},
					},
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

// TestCollectionMigrationQuery_ThreeIncrementalMigrations_ShouldMigrateAcrossAllThree
// extends the two-migration scenario in simple_test.go to three incremental
// registrations. Each registration triggers a reindex; intermediate registrations
// stamp the doc at the active version even though later transforms have not been
// applied, so subsequent registrations see a cached stamp and skip re-migration.
// https://github.com/sourcenetwork/defradb/issues/4736
func TestCollectionMigrationQuery_ThreeIncrementalMigrations_ShouldMigrateAcrossAllThree(t *testing.T) {
	test := testUtils.TestCase{
		// Same root cause as #4736 — skipped under secondary-index multiplier.
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }
					]
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} }
					]
				`,
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
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "verified",
									"value": true,
								},
							},
						},
					},
				},
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "{{.CollectionVersionID1}}",
					DestinationCollectionVersionID: "{{.CollectionVersionID2}}",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "email",
									"value": "ilovewasm@source.com",
								},
							},
						},
					},
				},
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "{{.CollectionVersionID2}}",
					DestinationCollectionVersionID: "{{.CollectionVersionID3}}",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "level",
									"value": 42,
								},
							},
						},
					},
				},
			},
			&action.Request{
				Request: `query {
					Users {
						name
						verified
						email
						level
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":     "John",
							"verified": true,
							"email":    "ilovewasm@source.com",
							"level":    int64(42),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestCollectionMigrationQuery_IntermediateMigrationRegisteredLast_ShouldProduceFullyMigratedResults
// covers the case where the outer migrations are registered first and an
// intermediate migration closes the chain last. Under an index, each
// registration triggers a reindex that stamps the doc past the still-nil
// intermediate edge; when that edge is later filled, the stamp already equals
// the active version and the fetcher short-circuits.
// https://github.com/sourcenetwork/defradb/issues/4736
func TestCollectionMigrationQuery_IntermediateMigrationRegisteredLast_ShouldProduceFullyMigratedResults(t *testing.T) {
	test := testUtils.TestCase{
		// Same root cause as #4736 — skipped under secondary-index multiplier.
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }
					]
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} }
					]
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "level", "Kind": "Int"} }
					]
				`,
			},
			// Register outer migrations first (v0->v1 and v2->v3).
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "{{.CollectionVersionID0}}",
					DestinationCollectionVersionID: "{{.CollectionVersionID1}}",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "verified",
									"value": true,
								},
							},
						},
					},
				},
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "{{.CollectionVersionID2}}",
					DestinationCollectionVersionID: "{{.CollectionVersionID3}}",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "level",
									"value": 42,
								},
							},
						},
					},
				},
			},
			// Finally register the intermediate migration (v1->v2) that closes the chain.
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "{{.CollectionVersionID1}}",
					DestinationCollectionVersionID: "{{.CollectionVersionID2}}",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "email",
									"value": "ilovewasm@source.com",
								},
							},
						},
					},
				},
			},
			&action.Request{
				Request: `query {
					Users {
						name
						verified
						email
						level
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":     "John",
							"verified": true,
							"email":    "ilovewasm@source.com",
							"level":    int64(42),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestCollectionMigrationQuery_IntermediateStampSurvivesRestart_ShouldRemigrateAfterRestart
// exercises the same incremental-registration pattern as the three-migration
// test but across a restart. The stale doc-version stamp is persisted to disk
// and survives the restart, so the second registration still sees a cache hit.
// https://github.com/sourcenetwork/defradb/issues/4736
func TestCollectionMigrationQuery_IntermediateStampSurvivesRestart_ShouldRemigrateAfterRestart(t *testing.T) {
	test := testUtils.TestCase{
		// Same root cause as #4736 — skipped under secondary-index multiplier.
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }
					]
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} }
					]
				`,
			},
			// Register only v0->v1. The v1->v2 chain will be incomplete, so the doc
			// is stamped at the intermediate version (v1) not at active v2.
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "{{.CollectionVersionID0}}",
					DestinationCollectionVersionID: "{{.CollectionVersionID1}}",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "verified",
									"value": true,
								},
							},
						},
					},
				},
			},
			testUtils.Restart{},
			// After restart, register the missing migration. If the intermediate
			// stamp survived, docs appear at v1 and get migrated v1->v2 here.
			// If restart lost the stamp and left the doc at v2, the v1->v2
			// migration would be skipped and `email` would stay nil.
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceCollectionVersionID:      "{{.CollectionVersionID1}}",
					DestinationCollectionVersionID: "{{.CollectionVersionID2}}",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "email",
									"value": "ilovewasm@source.com",
								},
							},
						},
					},
				},
			},
			&action.Request{
				Request: `query {
					Users {
						name
						verified
						email
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":     "John",
							"verified": true,
							"email":    "ilovewasm@source.com",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
