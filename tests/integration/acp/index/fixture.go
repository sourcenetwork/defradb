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

// policy id: "94eb195c0e459aa79e02a1986c7e731c5015721c18a373f2b2a0ed140a04b454"
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

// policy id: "f6927e8861f91122a5e3e333249297e4315b672298b5cb93ee3f49facc1e0d11"
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
