// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package test_acp_dac

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_QueryAverageWithoutIdentity(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			&action.Request{
				Request: `
					query {
						AVG(Employee: {field: salary})
					}
				`,
				Results: map[string]any{
					// 2 public employees, 1 with salary 10k, 1 with salary 20k
					"AVG": int(15000),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryAverageWithIdentity(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request: `
					query {
						AVG(Employee: {field: salary})
					}
				`,
				Results: map[string]any{
					// 4 employees with salaries 10k, 20k, 30k, 40k
					"AVG": int(25000),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryAverageWithWrongIdentity(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						AVG(Employee: {field: salary})
					}
				`,
				Results: map[string]any{
					// 2 public employees, 1 with salary 10k, 1 with salary 20k
					"AVG": int(15000),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
