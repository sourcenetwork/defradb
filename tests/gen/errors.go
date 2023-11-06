// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package gen

import "github.com/sourcenetwork/defradb/errors"

const (
	errInvalidConfiguration string = "invalid configuration"
	errFailedToParse        string = "failed to parse schema"
)

func NewErrInvalidConfiguration(reason string) error {
	return errors.New(errInvalidConfiguration, errors.NewKV("Reason", reason))
}

func NewErrFailedToParse(reason string) error {
	return errors.New(errFailedToParse, errors.NewKV("Reason", reason))
}
