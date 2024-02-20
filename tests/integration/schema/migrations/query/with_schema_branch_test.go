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
	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
)

func TestSchemaMigrationQuery_WithBranchingSchema(t *testing.T) {
	schemaVersion1ID := "bafkreiht46o4lakri2py2zw57ed3pdeib6ud6ojlsomgjlrgwh53wl3q4a"

	test := testUtils.TestCase{
		Description: "Test schema update, with branching schema migrations",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				SetAsDefaultVersion: immutable.Some(true),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
				Lens: immutable.Some(model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.SetDefaultModulePath,
							Arguments: map[string]any{
								"dst":   "name",
								"value": "Fred",
							},
						},
					},
				}),
			},
			testUtils.CreateDoc{
				// Create a document on the second schema version, with an email field value
				Doc: `{
					"name": "John",
					"email": "john@source.hub"
				}`,
			},
			testUtils.SetActiveSchemaVersion{
				// Set the active schema version back to the first
				SchemaVersionID: schemaVersion1ID,
			},
			testUtils.SchemaPatch{
				// The third schema version will be set as the active version, going from version 1 to 3
				SetAsDefaultVersion: immutable.Some(true),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "phone", "Kind": 11} }
					]
				`,
				Lens: immutable.Some(model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.SetDefaultModulePath,
							Arguments: map[string]any{
								"dst":   "phone",
								"value": "1234567890",
							},
						},
					},
				}),
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							name
							phone
						}
					}
				`,
				Results: []map[string]any{
					{
						// name has been cleared by the inverse of the migration from version 1 to 2
						"name": nil,
						// phone has been set by the migration from version 1 to 3
						"phone": "1234567890",
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
