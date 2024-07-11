// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package badger

import (
	dsq "github.com/ipfs/go-datastore/query"
	badger "github.com/sourcenetwork/badger/v4"

	datastoreErrors "github.com/sourcenetwork/defradb/datastore/errors"
	"github.com/sourcenetwork/defradb/errors"
)

var (
	ErrClosed       = datastoreErrors.ErrClosed
	ErrTxnConflict  = datastoreErrors.ErrTxnConflict
	ErrDiscardedTxn = datastoreErrors.ErrTxnDiscarded
)

const errOrderType string = "invalid order type"

func ErrOrderType(orderType dsq.Order) error {
	return errors.New(errOrderType, errors.NewKV("Order type", orderType))
}

// convertError converts badger specific errors into datastore errors.
func convertError(err error) error {
	// The errors we are matching against are never wrapped.
	//
	//nolint:errorlint
	switch err {
	case badger.ErrConflict:
		return datastoreErrors.ErrTxnConflict

	case badger.ErrReadOnlyTxn:
		return datastoreErrors.ErrReadOnlyTxn

	case badger.ErrDiscardedTxn:
		return datastoreErrors.ErrTxnDiscarded

	default:
		return err
	}
}
