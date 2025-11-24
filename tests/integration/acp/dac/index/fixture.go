// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_dac_index

const userPolicy = `
actor:
  name: actor
description: a test policy which marks a collection in a database as a resource
name: test
resources:
- name: users
  permissions:
  - expr: owner
    name: delete
  - expr: owner + reader
    name: read
  - expr: owner
    name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: owner
    types:
    - actor
  - name: reader
    types:
    - actor
`

const bookAuthorPolicy = `
actor:
  name: actor
description: a test policy which marks a collection in a database as a resource
name: test
resources:
- name: author
  permissions:
  - expr: owner
    name: delete
  - expr: owner + reader
    name: read
  - expr: owner
    name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: owner
    types:
    - actor
  - name: reader
    types:
    - actor
- name: book
  permissions:
  - expr: owner
    name: delete
  - expr: owner + reader
    name: read
  - expr: owner
    name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: owner
    types:
    - actor
  - name: reader
    types:
    - actor
`
