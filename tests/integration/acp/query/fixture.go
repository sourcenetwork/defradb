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
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	acpUtils "github.com/sourcenetwork/defradb/tests/integration/acp"
)

const employeeCompanyPolicy = `
description: A Valid DefraDB Policy Interface (DPI)

actor:
  name: actor

resources:
  employees:
    permissions:
      read:
        expr: owner + reader
      write:
        expr: owner

    relations:
      owner:
        types:
          - actor
      reader:
        types:
          - actor

  companies:
    permissions:
      read:
        expr: owner + reader
      write:
        expr: owner

    relations:
      owner:
        types:
          - actor
      reader:
        types:
          - actor
`

func getSetupEmployeeCompanyActions() []any {
	return []any{
		testUtils.AddPolicy{
			Identity:         acpUtils.Actor1Identity,
			Policy:           employeeCompanyPolicy,
			ExpectedPolicyID: "67607eb2a2a873f4a69eb6876323cee7601d8a4d4fedcc18154aaee65cf38e7f",
		},

		testUtils.SchemaUpdate{
			Schema: `
					type Employee @policy(
						id: "67607eb2a2a873f4a69eb6876323cee7601d8a4d4fedcc18154aaee65cf38e7f",
						resource: "employees"
					) {
						name: String
						salary: Int
						company: Company
					}

					type Company @policy(
						id: "67607eb2a2a873f4a69eb6876323cee7601d8a4d4fedcc18154aaee65cf38e7f",
						resource: "companies"
					) {
						name: String
						capital: Int
						employees: [Employee]
					}
				`,
		},

		testUtils.CreateDoc{
			CollectionID: 1,
			Doc: `
					{
						"name": "Public Company",
						"capital": 100000
					}
				`,
		},
		testUtils.CreateDoc{
			CollectionID: 1,
			Identity:     acpUtils.Actor1Identity,
			Doc: `
					{
						"name": "Private Company",
						"capital": 200000
					}
				`,
		},
		testUtils.CreateDoc{
			CollectionID: 0,
			Doc: `
					{
						"name": "PubEmp in PubCompany",
						"salary": 10000,
						"company": "bae-f4beaae4-ca13-5349-bccf-74a47bf1309e"
					}
				`,
		},
		testUtils.CreateDoc{
			CollectionID: 0,
			Doc: `
					{
						"name": "PubEmp in PrivateCompany",
						"salary": 20000,
						"company": "bae-74d9f5d4-fad7-5b87-8e8b-4d1e8cfbf83c"
					}
				`,
		},
		testUtils.CreateDoc{
			CollectionID: 0,
			Identity:     acpUtils.Actor1Identity,
			Doc: `
					{
						"name": "PrivateEmp in PubCompany",
						"salary": 30000,
						"company": "bae-f4beaae4-ca13-5349-bccf-74a47bf1309e"
					}
				`,
		},
		testUtils.CreateDoc{
			CollectionID: 0,
			Identity:     acpUtils.Actor1Identity,
			Doc: `
					{
						"name": "PrivateEmp in PrivateCompany",
						"salary": 40000,
						"company": "bae-74d9f5d4-fad7-5b87-8e8b-4d1e8cfbf83c"
					}
				`,
		},
	}
}
