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

func TestSchemaCreate_ContainsPNCounterTypeWithIntKind_NoError(t *testing.T) {
	schemaVersionID := "bafkreig54q5pw7elljueepsyux4qgdspm3ozct5dqocr5b2kufpjwb2mae"

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
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "points",
								ID:   1,
								Kind: client.FieldKind_INT,
								Typ:  client.PN_COUNTER,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaCreate_ContainsPNCounterTypeWithFloatKind_NoError(t *testing.T) {
	schemaVersionID := "bafkreibaeypr2i2eg3kozq3mlfsibgtolqlrcozo5ufqfb725dfq3hx43e"

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						points: Float @crdt(type: "pncounter")
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
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "points",
								ID:   1,
								Kind: client.FieldKind_FLOAT,
								Typ:  client.PN_COUNTER,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaCreate_ContainsPNCounterTypeWithWrongKind_Error(t *testing.T) {
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

func TestSchemaCreate_ContainsPNCounterWithInvalidType_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						points: Int @crdt(type: "invalid")
					}
				`,
				ExpectedError: "CRDT type not supported. Name: points, CRDTType: invalid",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
