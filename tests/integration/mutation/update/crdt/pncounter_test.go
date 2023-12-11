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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_PNCounterInt_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation of a PN Counter with Int type",
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
			testUtils.Request{
				Request: `query {
					Users {
						name
						points
					}
				}`,
				Results: []map[string]any{
					{
						"name":   "John",
						"points": int64(10),
					},
				},
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
				Results: []map[string]any{
					{
						"name":   "John",
						"points": int64(20),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test documents what happens when an overflow occurs in a PN Counter with Int type.
// In this case the value rolls over to the minimum int64 value.
func TestMutationUpdate_PNCounterIntWithOverflow_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation of a PN Counter with Int type and overflow",
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
				Results: []map[string]any{
					{
						"name":   "John",
						"points": math.MinInt64,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_PNCounterFloat_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation of a PN Counter with Float type",
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
			testUtils.Request{
				Request: `query {
					Users {
						name
						points
					}
				}`,
				Results: []map[string]any{
					{
						"name":   "John",
						"points": 10.1,
					},
				},
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
				Results: []map[string]any{
					{
						"name": "John",
						// Note the lack of precision of float types.
						"points": 20.299999999999997,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test documents what happens when an overflow occurs in a PN Counter with Float type.
// In this case it is the same as a no-op.
func TestMutationUpdate_PNCounterFloatWithOverflow_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple update mutation of a PN Counter with Float type and overflow",
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
					"points": %f
				}`, math.MaxFloat64),
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
				Results: []map[string]any{
					{
						"name":   "John",
						"points": math.MaxFloat64,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
