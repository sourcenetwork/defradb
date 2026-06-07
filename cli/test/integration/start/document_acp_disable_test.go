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
	"errors"
	"testing"

	"github.com/sourcenetwork/defradb/cli/test/action"
	"github.com/sourcenetwork/defradb/cli/test/integration"
)

// Document ACP can not be disabled: the legacy `--document-acp-type=none` value is no
// longer a supported engine, so startup fails with an unsupported-type error rather
// than starting the node with DAC turned off.
func TestStart_WithDocumentACPTypeNone_ReturnsUnsupportedTypeError(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			action.StartWithArgsE(
				[]string{"--document-acp-type=none"},
				errors.New("the selected acp type is not supported by this build"),
			),
		},
	}
	test.Execute(t)
}
