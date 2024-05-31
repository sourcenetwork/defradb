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
			ExpectedPolicyID: "6f11799717723307077147736fddccd8a7b5e68d2ec22e2155f0186e0c43a2e2",
		},

		testUtils.SchemaUpdate{
			Schema: `
					type Employee @policy(
						id: "6f11799717723307077147736fddccd8a7b5e68d2ec22e2155f0186e0c43a2e2",
						resource: "employees"
					) {
						name: String
						salary: Int
						company: Company
					}

					type Company @policy(
						id: "6f11799717723307077147736fddccd8a7b5e68d2ec22e2155f0186e0c43a2e2",
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
						"company": "bae-1ab7ac86-3c68-5abb-b526-803858c9dccf"
					}
				`,
		},
		testUtils.CreateDoc{
			CollectionID: 0,
			Doc: `
					{
						"name": "PubEmp in PrivateCompany",
						"salary": 20000,
						"company": "bae-4aef4bd6-e2ee-5075-85a5-4d64bbf80bca"
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
						"company": "bae-1ab7ac86-3c68-5abb-b526-803858c9dccf"
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
						"company": "bae-4aef4bd6-e2ee-5075-85a5-4d64bbf80bca"
					}
				`,
		},
	}
}
