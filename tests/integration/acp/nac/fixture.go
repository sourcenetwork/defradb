// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_nac

const examplePolicy = `
    name: Test Policy
    description: A Policy
    actor:
      name: actor
    resources:
      users:
        permissions:
          read:
            expr: owner + reader + updater + deleter
          update:
            expr: owner + updater
          delete:
            expr: owner + deleter
        relations:
          owner:
            types:
              - actor
          manager:
            types:
              - actor
            manages:
              - reader
              - updater
              - deleter
          reader:
            types:
              - actor
          updater:
            types:
              - actor
          deleter:
            types:
              - actor
`
