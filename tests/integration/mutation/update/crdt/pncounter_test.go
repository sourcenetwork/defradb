// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package update

import (
	"fmt"
	"math"
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestPNCounterUpdate_IntKindWithPositiveIncrement_ShouldIncrement(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Positive increments of a PN Counter with Int type",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Int @crdt(type: "pncounter")
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"points": 0
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 0,
				Doc: `{
					"points": 10
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 0,
				Doc: `{
					"points": 10
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						points
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "John",
							"points": int64(20),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test documents what happens when an overflow occurs in a PN Counter with Int type.
func TestPNCounterUpdate_IntKindWithPositiveIncrementOverflow_RollsOverToMinInt64(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Positive increments of a PN Counter with Int type causing overflow behaviour",
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// GQL mutation will return a type error in this case
			// because we are testing the internal overflow behaviour with
			// a int64 but the GQL Int type is an int32.
			testUtils.CollectionNamedMutationType,
			testUtils.CollectionSaveMutationType,
		}),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Int @crdt(type: "pncounter")
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: fmt.Sprintf(`{
					"name": "John",
					"points": %d
				}`, math.MaxInt64),
			},
			testUtils.UpdateDoc{
				DocID: 0,
				Doc: `{
					"points": 1
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						points
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "John",
							"points": int64(math.MinInt64),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestPNCounterUpdate_FloatKindWithPositiveIncrement_ShouldIncrement(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Positive increments of a PN Counter with Float type. Note the lack of precision",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Float @crdt(type: "pncounter")
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"points": 0
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 0,
				Doc: `{
					"points": 10.1
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 0,
				Doc: `{
					"points": 10.2
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						points
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							// Note the lack of precision of float types.
							"points": 20.299999999999997,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test documents what happens when an overflow occurs in a PN Counter with Float type.
func TestPNCounterUpdate_FloatKindWithPositiveIncrementOverflow_PositiveInf(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Positive increments of a PN Counter with Float type and overflow",
		SupportedClientTypes: immutable.Some(
			[]testUtils.ClientType{
				// This test only supports the Go client at the moment due to
				// https://github.com/sourcenetwork/defradb/issues/2569
				testUtils.GoClientType,
			},
		),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Float @crdt(type: "pncounter")
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: fmt.Sprintf(`{
					"name": "John",
					"points": %g
				}`, math.MaxFloat64),
			},
			testUtils.UpdateDoc{
				DocID: 0,
				Doc: fmt.Sprintf(`{
					"points": %g
				}`, math.MaxFloat64/10),
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						points
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "John",
							"points": math.Inf(1),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test documents what happens when an overflow occurs in a PN Counter with Float type.
func TestPNCounterUpdate_FloatKindWithDecrementOverflow_NegativeInf(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Positive increments of a PN Counter with Float type and overflow",
		SupportedClientTypes: immutable.Some(
			[]testUtils.ClientType{
				// This test only supports the Go client at the moment due to
				// https://github.com/sourcenetwork/defradb/issues/2569
				testUtils.GoClientType,
			},
		),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Float @crdt(type: "pncounter")
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: fmt.Sprintf(`{
					"name": "John",
					"points": %g
				}`, -math.MaxFloat64),
			},
			testUtils.UpdateDoc{
				DocID: 0,
				Doc: fmt.Sprintf(`{
					"points": %g
				}`, -math.MaxFloat64/10),
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						points
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "John",
							"points": math.Inf(-1),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestPNCounterUpdate_FloatKindWithPositiveIncrementInsignificantValue_DoesNothing(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Positive increments of a PN Counter with Float type and an insignificant value",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Float @crdt(type: "pncounter")
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: fmt.Sprintf(`{
					"name": "John",
					"points": %g
				}`, math.MaxFloat64/10),
			},
			testUtils.UpdateDoc{
				// `1` is insignificant to a large float64 and adding it to the large value
				// should not result in a value change
				Doc: `{
					"points": 1
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						points
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "John",
							"points": math.MaxFloat64 / 10,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
