// Copyright 2020 Source Inc.
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
	blocks "github.com/ipfs/go-block-format"
	cid "github.com/ipfs/go-cid"
)

func (db *DB) GetBlock(c cid.Cid) (blocks.Block, error) {
	return db.dagstore.Get(c)
}
