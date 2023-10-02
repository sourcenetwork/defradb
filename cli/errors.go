// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"github.com/sourcenetwork/defradb/errors"
)

const errInvalidLensConfig = "invalid lens configuration"

var (
	ErrNoDocOrFile         = errors.New("document or file must be defined")
	ErrInvalidDocument     = errors.New("invalid document")
	ErrNoDocKeyOrFilter    = errors.New("document key or filter must be defined")
	ErrInvalidExportFormat = errors.New("invalid export format")
	ErrNoLensConfig        = errors.New("lens config cannot be empty")
	ErrInvalidLensConfig   = errors.New("invalid lens configuration")
)

func NewErrInvalidLensConfig(inner error) error {
	return errors.Wrap(errInvalidLensConfig, inner)
}
