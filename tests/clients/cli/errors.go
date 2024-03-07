// Copyright 2024 Democratized Data Foundation
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

const (
	errMethodIsNotImplemented string = "the method is not implemented"
)

// Errors returnable from this package.
var (
	ErrMethodIsNotImplemented = errors.New(errMethodIsNotImplemented)
)
