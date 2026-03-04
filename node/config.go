// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node

const (
	// 1 MB, this matches the maximum badger-in-memory value size.
	//
	// Nearly at least, badger panics if this is set to it's max for reasons not yet
	// looked into.  Going one byte smaller does not have this issue.
	defaultChunkSize = (1 << 20) - 1
)
