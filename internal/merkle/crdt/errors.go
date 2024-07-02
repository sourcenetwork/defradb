// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package merklecrdt

import (
	"fmt"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errUnexpectedValueType = "unexpected value type for merkle CRDT"
)

var (
	ErrUnexpectedValueType = errors.New(errUnexpectedValueType)
)

func NewErrUnexpectedValueType(cType client.CType, expected, actual any) error {
	return errors.New(
		errUnexpectedValueType,
		errors.NewKV("CRDT", cType.String()),
		errors.NewKV("expected", fmt.Sprintf("%T", expected)),
		errors.NewKV("actual", fmt.Sprintf("%T", actual)),
	)
}
