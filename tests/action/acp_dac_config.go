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

package action

import (
	"os"

	"github.com/sourcenetwork/defradb/tests/state"
)

const documentACPTypeEnvName = "DEFRA_DOCUMENT_ACP_TYPE"

// DocumentACPType is the document ACP implementation under test.
//
// Node setup and the test harness both read this, so it is resolved once here
// rather than copied into each package.
var DocumentACPType state.DocumentACPType

func init() {
	DocumentACPType = state.DocumentACPType(os.Getenv(documentACPTypeEnvName))
	if DocumentACPType == "" {
		DocumentACPType = state.LocalDocumentACPType
	}
}
