// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package sequence

import "github.com/sourcenetwork/defradb/errors"

const (
	errGetSequenceValue    string = "failed to get sequence value"
	errUpdateSequenceValue string = "failed to update sequence value"
)

func NewErrGetSequenceValue(inner error, key string) error {
	return errors.Wrap(errGetSequenceValue, inner, errors.NewKV("Key", key))
}

func NewErrUpdateSequenceValue(inner error, key string) error {
	return errors.Wrap(errUpdateSequenceValue, inner, errors.NewKV("Key", key))
}
