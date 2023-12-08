// Copyright 2022 Democratized Data Foundation
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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaCreate_PNCounterType_NoError(t *testing.T) {
	schemaVersionID := "bafkreia444xgvvpyyvxn2m56mgsyovhtrbbx6zpmn4ocnkqbbjnytlfvrm"

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						points: Int @crdt(type: "pncounter")
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
						Fields: []client.FieldDescription{
							{
								Name: "_key",
								Kind: client.FieldKind_DocKey,
							},
							{
								Name: "points",
								ID:   1,
								Kind: client.FieldKind_INT,
								Typ:  client.PN_COUNTER_REGISTER,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaCreate_PNCounterTypeWithWrongKind_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						points: String @crdt(type: "pncounter")
					}
				`,
				ExpectedError: "CRDT type pncounter can't be assigned to field kind String",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaCreate_InvalidType_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						points: Int @crdt(type: "pn_counter")
					}
				`,
				ExpectedError: "CRDT type not supported. Name: points, CRDTType: pn_counter",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
