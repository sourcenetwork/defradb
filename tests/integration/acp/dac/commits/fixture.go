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
const userDocID = "bae-cad49a1d-299c-5c34-9dab-a23f233f1a2f"

// userDocCompositeCid is the deterministic composite-block cid for the userDoc.
// It is only valid when the SignedDocs multiplier is excluded.
const userDocCompositeCid = "bafyreifiehbtwpqssac2tk33na4agof7a23ymr5vd5xx6zty3zclftdogu"

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

// secondUserDoc is a private document created by identity 2 used in multi-doc tests.
const secondUserDoc = `{"name": "John", "age": 30}`

// secondUserDocCompositeCid is the deterministic composite-block cid for secondUserDoc.
// Only valid when the SignedDocs and EncryptedDocs multipliers are excluded.
const secondUserDocCompositeCid = "bafyreidkaczhkludh2nk7u6zorcx7ltg5hi4lxyadvyqkerezdrj42py44"

