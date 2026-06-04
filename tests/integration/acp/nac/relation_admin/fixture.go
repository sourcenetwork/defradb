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

package test_acp_nac_relation_admin

const examplePolicy = `
description: A Policy
name: Test Policy
resources:
- name: users
  permissions:
  - expr: deleter
    name: delete
  - expr: reader + updater + deleter
    name: read
  - expr: updater
    name: update
  relations:
  - name: deleter
    types:
    - actor
  - manages:
    - reader
    - updater
    - deleter
    name: manager
    types:
    - actor
  - name: reader
    types:
    - actor
  - name: updater
    types:
    - actor
`
