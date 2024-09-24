// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

import (
	"github.com/fxamacker/cbor/v2"
	"google.golang.org/grpc/encoding"
)

const cborCodecName = "cbor"

// cborCodec is a gRPC Codec implementation with CBOR encoding.
type cborCodec struct{}

func (c *cborCodec) Marshal(v interface{}) ([]byte, error) {
	return cbor.Marshal(v)
}

func (c *cborCodec) Unmarshal(data []byte, v interface{}) error {
	if v == nil {
		return nil
	}
	return cbor.Unmarshal(data, v)
}

func (c *cborCodec) Name() string {
	return cborCodecName
}

func init() {
	encoding.RegisterCodec(&cborCodec{})
}
