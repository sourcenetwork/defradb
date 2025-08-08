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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaCreate_ContainsPNCounterTypeWithIntKind_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						points: Int @crdt(type: pncounter)
					}
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
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
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						points: Float @crdt(type: pncounter)
					}
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "points",
								Kind: client.FieldKind_NILLABLE_FLOAT64,
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
			&action.AddSchema{
				Schema: `
					type Users {
						points: String @crdt(type: pncounter)
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
			&action.AddSchema{
				Schema: `
					type Users {
						points: Int @crdt(type: "invalid")
					}
				`,
				ExpectedError: `Argument "type" has invalid value "invalid"`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaCreate_ContainsPCounterTypeWithIntKind_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						points: Int @crdt(type: pcounter)
					}
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
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
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						points: Float @crdt(type: pcounter)
					}
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "points",
								Kind: client.FieldKind_NILLABLE_FLOAT64,
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

func TestSchemaCreate_ContainsPCounterTypeWithFloat64Kind_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						points: Float64 @crdt(type: pcounter)
					}
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "points",
								Kind: client.FieldKind_NILLABLE_FLOAT64,
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

func TestSchemaCreate_ContainsPCounterTypeWithFloat32Kind_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						points: Float32 @crdt(type: pcounter)
					}
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "points",
								Kind: client.FieldKind_NILLABLE_FLOAT32,
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
			&action.AddSchema{
				Schema: `
					type Users {
						points: String @crdt(type: pcounter)
					}
				`,
				ExpectedError: "CRDT type pcounter can't be assigned to field kind String",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
