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

package predefined

import "github.com/sourcenetwork/defradb/errors"

const (
	errFailedToGenerateDoc string = "failed to generate doc"
)

func NewErrFailedToGenerateDoc(inner error) error {
	return errors.Wrap(errFailedToGenerateDoc, inner)
}
