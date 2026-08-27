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
	sourcehubImageEnvName  = "DEFRA_SOURCEHUB_IMAGE"
)

var (
	// documentACPType is the document ACP implementation under test.
	documentACPType state.DocumentACPType

	// sourcehubImage is the container image used to run SourceHub.
	sourcehubImage string
)

func init() {
	documentACPType = state.DocumentACPType(os.Getenv(documentACPTypeEnvName))
	if documentACPType == "" {
		documentACPType = state.LocalDocumentACPType
	}
	sourcehubImage = os.Getenv(sourcehubImageEnvName)
	if sourcehubImage == "" {
		sourcehubImage = "ghcr.io/sourcenetwork/vera:dev"
	}
}
