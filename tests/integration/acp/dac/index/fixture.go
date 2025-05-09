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
name: test
description: a test policy which marks a collection in a database as a resource

actor:
  name: actor

resources:
  users:
    permissions:
      read:
        expr: owner + reader
      update:
        expr: owner
      delete:
        expr: owner

    relations:
      owner:
        types:
          - actor
      reader:
        types:
          - actor
      admin:
        manages:
          - reader
        types:
          - actor
`

const bookAuthorPolicy = `
name: test
description: a test policy which marks a collection in a database as a resource

actor:
  name: actor

resources:
  author:
    permissions:
      read:
        expr: owner + reader
      update:
        expr: owner
      delete:
        expr: owner

    relations:
      owner:
        types:
          - actor
      reader:
        types:
          - actor
      admin:
        manages:
          - reader
        types:
          - actor

  book:
    permissions:
      read:
        expr: owner + reader
      update:
        expr: owner
      delete:
        expr: owner

    relations:
      owner:
        types:
          - actor
      reader:
        types:
          - actor
      admin:
        manages:
          - reader
        types:
          - actor
`
