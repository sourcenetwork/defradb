// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package base

type DataEncoding uint32

const (
	// Indicates that the data is encoded using the CBOR encoding for values.
	DataEncoding_VALUE_CBOR DataEncoding = 0
)
