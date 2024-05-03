// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	acpUtils "github.com/sourcenetwork/defradb/tests/integration/acp"
)

func TestACP_QueryAverageWithoutIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query average without identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Request: `
					query {
						_avg(Employee: {field: salary})
					}
				`,
				Results: []map[string]any{
					{
						"_avg": int(15000),
					},
				},
			},

			testUtils.Request{
				Request: `
					query {
						_avg(Company: {field: capital})
					}
				`,
				Results: []map[string]any{
					{
						"_avg": int(100000),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryAverageWithIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query average with identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Identity: acpUtils.Actor1Identity,
				Request: `
					query {
						_avg(Employee: {field: salary})
					}
				`,
				Results: []map[string]any{
					{
						"_avg": int(25000),
					},
				},
			},

			testUtils.Request{
				Identity: acpUtils.Actor1Identity,
				Request: `
					query {
						_avg(Company: {field: capital})
					}
				`,
				Results: []map[string]any{
					{
						"_avg": int(150000),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryAverageWithWrongIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query average without identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Identity: acpUtils.Actor2Identity,
				Request: `
					query {
						_avg(Employee: {field: salary})
					}
				`,
				Results: []map[string]any{
					{
						"_avg": int(15000),
					},
				},
			},

			testUtils.Request{
				Identity: acpUtils.Actor2Identity,
				Request: `
					query {
						_avg(Company: {field: capital})
					}
				`,
				Results: []map[string]any{
					{
						"_avg": int(100000),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
