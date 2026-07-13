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

package test_acp_dac_commits

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

const userDoc = `
{
	"name": "Shahzad",
	"age": 28
}
`

// userDocID is the docID for the userDoc.
const userDocID = "{{.DocID0_0}}"

// userDocCompositeCid is the composite-block cid for the userDoc.
const userDocCompositeCid = "{{.CID0_0_0}}"

const usersAndPostsPolicy = `
description: a test policy with two resources
name: test
resources:
- name: users
  permissions:
  - name: delete
  - name: read
    expr: reader
  - name: update
  relations:
  - name: reader
    types:
    - actor
- name: posts
  permissions:
  - name: delete
  - name: read
    expr: reader
  - name: update
  relations:
  - name: reader
    types:
    - actor
`
