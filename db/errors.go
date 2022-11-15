// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"fmt"

	"github.com/sourcenetwork/defradb/errors"
)

var (
	ErrSubscriptionsNotAllowed = errors.New("server does not accept subscriptions")
	ErrUnexpectedType          = errors.New("unexpected type")
)

func NewErrUnexpectedType[T any](actual any) error {
	var expected T
	return errors.WithStack(
		ErrUnexpectedType,
		errors.NewKV("Expected", fmt.Sprintf("%T", expected)),
		errors.NewKV("Actual", fmt.Sprintf("%T", actual)),
	)
}
