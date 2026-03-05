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
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

const employeeCompanyPolicy = `
description: A Policy
name: test
resources:
- name: companies
  permissions:
  - name: delete
  - expr: reader
    name: read
  - name: update
  relations:
  - name: reader
    types:
    - actor
- name: employees
  permissions:
  - name: delete
  - expr: reader
    name: read
  - name: update
  relations:
  - name: reader
    types:
    - actor
`

func getSetupEmployeeCompanyActions() []any {
	return []any{
		testUtils.AddDACPolicy{
			Identity: testUtils.ClientIdentity(1),
			Policy:   employeeCompanyPolicy,
		},

		&action.AddCollection{
			SDL: `
					type Employee @policy(
						id: "{{.Policy0}}",
						resource: "employees"
					) {
						name: String
						salary: Int
						company: Company
					}

					type Company @policy(
						id: "{{.Policy0}}",
						resource: "companies"
					) {
						name: String
						capital: Int
						employees: [Employee]
					}
				`,
		},

		&action.AddDoc{
			CollectionID: 1,
			Doc: `
					{
						"name": "Public Company",
						"capital": 100000
					}
				`,
		},
		&action.AddDoc{
			CollectionID: 1,
			Identity:     testUtils.ClientIdentity(1),
			Doc: `
					{
						"name": "Private Company",
						"capital": 200000
					}
				`,
		},
		&action.AddDoc{
			CollectionID: 0,
			DocMap: map[string]any{
				"name":    "PubEmp in PubCompany",
				"salary":  10000,
				"company": testUtils.NewDocIndex(1, 0),
			},
		},
		&action.AddDoc{
			CollectionID: 0,
			DocMap: map[string]any{
				"name":    "PubEmp in PrivateCompany",
				"salary":  20000,
				"company": testUtils.NewDocIndex(1, 1),
			},
		},
		&action.AddDoc{
			CollectionID: 0,
			Identity:     testUtils.ClientIdentity(1),
			DocMap: map[string]any{
				"name":    "PrivateEmp in PubCompany",
				"salary":  30000,
				"company": testUtils.NewDocIndex(1, 0),
			},
		},
		&action.AddDoc{
			CollectionID: 0,
			Identity:     testUtils.ClientIdentity(1),
			DocMap: map[string]any{
				"name":    "PrivateEmp in PrivateCompany",
				"salary":  40000,
				"company": testUtils.NewDocIndex(1, 1),
			},
		},
	}
}
