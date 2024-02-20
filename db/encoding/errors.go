// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package encoding

import (
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errCanNotDecode           = "can not decode"
	errCanNotDecodeFieldValue = "can not decode field value"
)

var (
	ErrCanNotDecode           = errors.New(errCanNotDecode)
	ErrCanNotDecodeFieldValue = errors.New(errCanNotDecodeFieldValue)
)

// NewErrCanNotDecode returns a new error indicating that the encoded
// bytes could not be decoded.
func NewErrCanNotDecode(b []byte) error {
	return errors.New(errCanNotDecode, errors.NewKV("Bytes", b))
}

// NewErrCanNotDecodeFieldValue returns a new error indicating that the encoded
// bytes could not be decoded into a client.FieldValue of a certain kind.
func NewErrCanNotDecodeFieldValue(b []byte, kind client.FieldKind, innerErr ...error) error {
	kvs := []errors.KV{errors.NewKV("Bytes", b), errors.NewKV("Kind", kind)}
	if len(innerErr) > 0 {
		kvs = append(kvs, errors.NewKV("InnerErr", innerErr[0]))
	}
	return errors.New(errCanNotDecodeFieldValue, kvs...)
}
