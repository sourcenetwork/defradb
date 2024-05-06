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
	schemaVersionID := "bafkreigsnu67poxm3663e7vl5cncl6pxdzndcc7jf66cnnvxzw5uko5iuu"

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
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "points",
								Kind: client.FieldKind_NILLABLE_INT,
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
	schemaVersionID := "bafkreieflo3tkhsywsqcyzoj6nqgxc6ovv5m5lc7bfbum6yqls5rxlwkye"

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
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "points",
								Kind: client.FieldKind_NILLABLE_FLOAT,
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

func TestSchemaCreate_ContainsPCounterTypeWithIntKind_NoError(t *testing.T) {
	schemaVersionID := "bafkreigbmy67fjsys3li5rbs64k3vezvdtbfryc67pxiju4nis7lrbanea"

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						points: Int @crdt(type: "pcounter")
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
								Name: "points",
								Kind: client.FieldKind_NILLABLE_INT,
								Typ:  client.P_COUNTER,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaCreate_ContainsPCounterTypeWithFloatKind_NoError(t *testing.T) {
	schemaVersionID := "bafkreifcyba45ov5zqi6dbhlu72rmf4wp3crjynjvvpq6iuauns2ofbvzi"

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						points: Float @crdt(type: "pcounter")
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
								Name: "points",
								Kind: client.FieldKind_NILLABLE_FLOAT,
								Typ:  client.P_COUNTER,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaCreate_ContainsPCounterTypeWithWrongKind_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						points: String @crdt(type: "pcounter")
					}
				`,
				ExpectedError: "CRDT type pcounter can't be assigned to field kind String",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
