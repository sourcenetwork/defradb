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

func TestACP_QueryManyToOneRelationObjectsWithoutIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query employees with their companies without identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Request: `
					query {
						Employee {
							name
							company {
								name
							}
						}
					}
				`,
				Results: []map[string]any{
					{
						"name":    "PubEmp in PubCompany",
						"company": map[string]any{"name": "Public Company"},
					},
					{
						"name":    "PubEmp in PrivateCompany",
						"company": nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryOneToManyRelationObjectsWithoutIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query companies with their employees without identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Request: `
					query {
						Company {
							name
							employees {
								name
							}
						}
					}
				`,
				Results: []map[string]any{
					{
						"name": "Public Company",
						"employees": []map[string]any{
							{"name": "PubEmp in PubCompany"},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryManyToOneRelationObjectsWithIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query employees with their companies with identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Identity: immutable.Some(1),
				Request: `
					query {
						Employee {
							name
							company {
								name
							}
						}
					}
				`,
				Results: []map[string]any{
					{
						"name":    "PrivateEmp in PubCompany",
						"company": map[string]any{"name": "Public Company"},
					},
					{
						"name":    "PrivateEmp in PrivateCompany",
						"company": map[string]any{"name": "Private Company"},
					},
					{
						"name":    "PubEmp in PubCompany",
						"company": map[string]any{"name": "Public Company"},
					},
					{
						"name":    "PubEmp in PrivateCompany",
						"company": map[string]any{"name": "Private Company"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryOneToManyRelationObjectsWithIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query companies with their employees with identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Identity: immutable.Some(1),
				Request: `
					query {
						Company {
							name
							employees {
								name
							}
						}
					}
				`,
				Results: []map[string]any{
					{
						"name": "Public Company",
						"employees": []map[string]any{
							{"name": "PrivateEmp in PubCompany"},
							{"name": "PubEmp in PubCompany"},
						},
					},
					{
						"name": "Private Company",
						"employees": []map[string]any{
							{"name": "PrivateEmp in PrivateCompany"},
							{"name": "PubEmp in PrivateCompany"},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryManyToOneRelationObjectsWithWrongIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query employees with their companies with wrong identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Identity: immutable.Some(2),
				Request: `
					query {
						Employee {
							name
							company {
								name
							}
						}
					}
				`,
				Results: []map[string]any{
					{
						"name":    "PubEmp in PubCompany",
						"company": map[string]any{"name": "Public Company"},
					},
					{
						"name":    "PubEmp in PrivateCompany",
						"company": nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_QueryOneToManyRelationObjectsWithWrongIdentity(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, query companies with their employees with wrong identity",

		Actions: []any{
			getSetupEmployeeCompanyActions(),

			testUtils.Request{
				Identity: immutable.Some(2),
				Request: `
					query {
						Company {
							name
							employees {
								name
							}
						}
					}
				`,
				Results: []map[string]any{
					{
						"name": "Public Company",
						"employees": []map[string]any{
							{"name": "PubEmp in PubCompany"},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
