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

func TestACP_QueryCountWithoutIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query count without identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Request: `
					query {
						_count(Employee: {})
					}
				`,
				Results: []map[string]any{
					{
						"_count": int(2),
					},
				},
			},

			testUtils.Request{
				Request: `
					query {
						_count(Company: {})
					}
				`,
				Results: []map[string]any{
					{
						"_count": int(1),
					},
				},
			},

			testUtils.Request{
				Request: `
					query {
						Company {
							_count(employees: {})
						}
					}
				`,
				Results: []map[string]any{
					{
						"_count": int(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryCountWithIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query count with identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Identity: acpUtils.Actor1Identity,
				Request: `
					query {
						_count(Employee: {})
					}
				`,
				Results: []map[string]any{
					{
						"_count": int(4),
					},
				},
			},

			testUtils.Request{
				Identity: acpUtils.Actor1Identity,
				Request: `
					query {
						_count(Company: {})
					}
				`,
				Results: []map[string]any{
					{
						"_count": int(2),
					},
				},
			},

			testUtils.Request{
				Identity: acpUtils.Actor1Identity,
				Request: `
					query {
						Company {
							_count(employees: {})
						}
					}
				`,
				Results: []map[string]any{
					{
						"_count": int(2),
					},
					{
						"_count": int(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryCountWithWrongIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query count without identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Identity: acpUtils.Actor2Identity,
				Request: `
					query {
						_count(Employee: {})
					}
				`,
				Results: []map[string]any{
					{
						"_count": int(2),
					},
				},
			},

			testUtils.Request{
				Identity: acpUtils.Actor2Identity,
				Request: `
					query {
						_count(Company: {})
					}
				`,
				Results: []map[string]any{
					{
						"_count": int(1),
					},
				},
			},

			testUtils.Request{
				Identity: acpUtils.Actor2Identity,
				Request: `
					query {
						Company {
							_count(employees: {})
						}
					}
				`,
				Results: []map[string]any{
					{
						"_count": int(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
