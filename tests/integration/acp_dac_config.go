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

package tests

import (
	"os"

	"github.com/sourcenetwork/defradb/tests/state"
)

const (
	documentACPTypeEnvName = "DEFRA_DOCUMENT_ACP_TYPE"
	veraImageEnvName       = "DEFRA_VERA_IMAGE"
)

var (
	// documentACPType is the document ACP implementation under test.
	documentACPType state.DocumentACPType

	// veraImage is the container image used to run Vera.
	veraImage string
)

func init() {
	documentACPType = state.DocumentACPType(os.Getenv(documentACPTypeEnvName))
	if documentACPType == "" {
		documentACPType = state.LocalDocumentACPType
	}
	veraImage = os.Getenv(veraImageEnvName)
	if veraImage == "" {
		veraImage = "ghcr.io/sourcenetwork/vera:dev"
	}
}
