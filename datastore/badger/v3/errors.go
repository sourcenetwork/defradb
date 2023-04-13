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

	"github.com/sourcenetwork/defradb/errors"
)

const errOrderType string = "invalid order type"

func ErrOrderType(orderType dsq.Order) error {
	return errors.New(errOrderType, errors.NewKV("Order type", orderType))
}
