// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package auth

import (
	"testing"

	"github.com/sourcenetwork/defradb/cli/test/action"
	"github.com/sourcenetwork/defradb/cli/test/integration"
	"github.com/sourcenetwork/defradb/client"
)

// testIdentityKeyHex is a valid secp256k1 private key used to mint the auth token.
const testIdentityKeyHex = "e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac"

// When an identity is provided, the CLI mints an auth token whose audience defaults
// to the host of --url. That host is exactly what the server validates the token
// against, so the request must be accepted.
func TestAuth_AudienceDerivedFromURL_IsAccepted(t *testing.T) {
	addCollection := &action.AddCollection{
		InlineSDL: `
			type User {}
		`,
	}
	addCollection.AddArgs("--identity", testIdentityKeyHex)

	describeCollection := &action.DescribeCollection{
		Expected: []client.CollectionVersion{
			{
				Name:           "User",
				IsMaterialized: true,
			},
		},
	}
	describeCollection.AddArgs("--identity", testIdentityKeyHex)

	test := &integration.Test{
		Actions: []action.Action{
			addCollection,
			describeCollection,
		},
	}

	test.Execute(t)
}

// Overriding the audience with a host the server is not reached at produces a token
// whose audience cannot match the request host, so the server must reject it.
func TestAuth_AudienceMismatch_IsForbidden(t *testing.T) {
	describeCollection := &action.DescribeCollection{
		ExpectError: "forbidden",
	}
	describeCollection.AddArgs(
		"--identity", testIdentityKeyHex,
		"--audience", "wrong-host:9999",
	)

	test := &integration.Test{
		Actions: []action.Action{
			describeCollection,
		},
	}

	test.Execute(t)
}
