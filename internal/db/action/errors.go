// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package action

import (
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errActionInProgress string = "multiple actions of the same kind and collection " +
		"cannot be submitted concurrently"
	errInvalidActionStatusEncoding string = "invalid action status encoding"
)

var (
	ErrActionInProgress            = errors.New(errActionInProgress)
	ErrInvalidActionStatusEncoding = errors.New(errInvalidActionStatusEncoding)
)

func NewErrActionInProgress(collectionID string, action client.Action) error {
	return errors.New(
		errActionInProgress,
		errors.NewKV("CollectionID", collectionID),
		errors.NewKV("Action", action),
	)
}
