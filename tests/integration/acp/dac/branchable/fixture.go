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

package test_acp_dac_branchable

// usersPolicy is a policy whose "users" resource grants read via the reader relation, and lets an
// admin manage readers. It is used both for documents and for the branchable collection object.
const usersPolicy = `
description: a test policy
name: test
resources:
- name: users
  permissions:
  - name: delete
  - name: read
    expr: reader
  - name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: reader
    types:
    - actor
`

// branchablePermissionedSDL is a branchable collection linked to the test policy.
const branchablePermissionedSDL = `
	type Users @branchable @policy(
		id: "{{.Policy0}}",
		resource: "users"
	) {
		name: String
		age: Int
	}
`

const userDoc = `
{
	"name": "Shahzad",
	"age": 28
}
`
