// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"

	"github.com/sourcenetwork/immutable"
)

func TestSchemaCreate_WithDefaultFieldValues(t *testing.T) {
	schemaVersionID := "bafkreidgt7jiy2abozwydf3frpvvvu6whuvw2saqm6erflnf3vqlnapxly"

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						active: Boolean @default(bool: true)
						name: String @default(string: "Bob")
						age: Int @default(int: 10)
						points: Float @default(float: 30)
					}
				`,
			},
			testUtils.GetSchema{
				VersionID: immutable.Some(schemaVersionID),
				ExpectedResults: []client.SchemaDescription{
					{
						Name:      "Users",
						VersionID: schemaVersionID,
						Root:      schemaVersionID,
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "active",
								Kind:         client.FieldKind_NILLABLE_BOOL,
								Typ:          client.LWW_REGISTER,
								DefaultValue: true,
							},
							{
								Name:         "age",
								Kind:         client.FieldKind_NILLABLE_INT,
								Typ:          client.LWW_REGISTER,
								DefaultValue: float64(10),
							},
							{
								Name:         "name",
								Kind:         client.FieldKind_NILLABLE_STRING,
								Typ:          client.LWW_REGISTER,
								DefaultValue: "Bob",
							},
							{
								Name:         "points",
								Kind:         client.FieldKind_NILLABLE_FLOAT,
								Typ:          client.LWW_REGISTER,
								DefaultValue: float64(30),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
