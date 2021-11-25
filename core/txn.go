// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package core

import (
	"github.com/sourcenetwork/defradb/datastores/iterable"
)

// Txn is a common interface to the db.Txn struct
type Txn interface {
	iterable.IterableTxn
	MultiStore
	Systemstore() DSReaderWriter
}
