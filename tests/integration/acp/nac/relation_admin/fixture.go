// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

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
