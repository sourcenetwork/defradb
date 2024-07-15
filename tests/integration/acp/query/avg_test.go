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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
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
						// 2 public employees, 1 with salary 10k, 1 with salary 20k
						"_avg": int(15000),
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
				Identity: immutable.Some(1),
				Request: `
					query {
						_avg(Employee: {field: salary})
					}
				`,
				Results: []map[string]any{
					{
						// 4 employees with salaries 10k, 20k, 30k, 40k
						"_avg": int(25000),
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
				Identity: immutable.Some(2),
				Request: `
					query {
						_avg(Employee: {field: salary})
					}
				`,
				Results: []map[string]any{
					{
						// 2 public employees, 1 with salary 10k, 1 with salary 20k
						"_avg": int(15000),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
