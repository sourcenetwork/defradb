// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package query

import (
	"testing"

	"github.com/lens-vm/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestSchemaMigrationQueryInversesAcrossMultipleVersions(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema migration, inverses across multiple migrated versions",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						age: Int
						height: Int
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "verified", "Kind": "Boolean"} }
					]
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} }
					]
				`,
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "bafkreicdkt3m6mgwuoix7qyijvwxwtj3dlre4a4c6mdnqbucbndwuxjsvi",
					DestinationSchemaVersionID: "bafkreigijxrkfpadmnkpagokjdy6zpwtryad32m6nkgsqrd452kjlfp46e",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "age",
									"value": 30,
								},
							},
						},
					},
				},
			},
			testUtils.ConfigureMigration{
				LensConfig: client.LensConfig{
					SourceSchemaVersionID:      "bafkreigijxrkfpadmnkpagokjdy6zpwtryad32m6nkgsqrd452kjlfp46e",
					DestinationSchemaVersionID: "bafkreibtmdbc3nbdt74xdwvfrez53fxwyz6nh4b6ppwsrxiqpj5zpwgole",
					Lens: model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.SetDefaultModulePath,
								Arguments: map[string]any{
									"dst":   "height",
									"value": 190,
								},
							},
						},
					},
				},
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 33,
					"height": 185
				}`,
			},
			testUtils.SetActiveSchemaVersion{
				SchemaVersionID: "bafkreicdkt3m6mgwuoix7qyijvwxwtj3dlre4a4c6mdnqbucbndwuxjsvi",
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						age
						height
					}
				}`,
				Results: []map[string]any{
					{
						"name":   "John",
						"age":    nil,
						"height": nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
