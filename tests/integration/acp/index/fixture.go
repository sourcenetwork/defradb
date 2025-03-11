// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_index

// policy id: "abe378ae8dac56f43238b56126a5a5ff1d1021e6bf8027d477b5a366e6238fc2"
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
      write:
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

// policy id: "35645d025f17eb228814dd15b61c7bc04ff008e4829bc5473211930376f11ff8"
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
      write:
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
      write:
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
