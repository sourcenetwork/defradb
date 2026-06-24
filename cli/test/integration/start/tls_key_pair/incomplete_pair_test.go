// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package start

import (
	"testing"

	"github.com/sourcenetwork/defradb/cli"
	"github.com/sourcenetwork/defradb/cli/test/action"
	"github.com/sourcenetwork/defradb/cli/test/integration"
)

// Enabling TLS requires both the public (pubkeypath) and private (privkeypath)
// key paths. Providing only one must fail fast with ErrIncompleteTLSKeyPair.

func TestStart_WithOnlyTLSPublicKeyPath_Error(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			action.StartWithArgsE([]string{"--pubkeypath=server.crt"}, cli.ErrIncompleteTLSKeyPair),
		},
	}
	test.Execute(t)
}

func TestStart_WithOnlyTLSPrivateKeyPath_Error(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			action.StartWithArgsE([]string{"--privkeypath=server.key"}, cli.ErrIncompleteTLSKeyPair),
		},
	}
	test.Execute(t)
}
