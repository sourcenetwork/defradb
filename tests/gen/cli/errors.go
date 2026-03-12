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

package cli

import "github.com/sourcenetwork/defradb/errors"

const (
	errInvalidDemandValue string = "invalid demand value"
)

func NewErrInvalidDemandValue(inner error) error {
	return errors.Wrap(errInvalidDemandValue, inner)
}
