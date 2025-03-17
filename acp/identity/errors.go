// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package identity

import (
	"encoding/hex"

	"github.com/sourcenetwork/defradb/errors"
)

const (
	errDIDCreation        = "could not produce did for key"
	errUnsupportedKeyType = "unsupported key type"
)

var (
	ErrDIDCreation = errors.New(errDIDCreation)
	// ErrUnsupportedKeyType is returned when attempting to use an unsupported key type.
	ErrUnsupportedKeyType = errors.New(errUnsupportedKeyType)
)

func newErrDIDCreation(inner error, keytype string, pubKey []byte) error {
	return errors.Wrap(
		errDIDCreation,
		inner,
		errors.NewKV("KeyType", keytype),
		errors.NewKV("PubKey", hex.EncodeToString(pubKey)),
	)
}

func newErrUnsupportedKeyType(keyType string) error {
	return errors.New(errUnsupportedKeyType, errors.NewKV("KeyType", keyType))
}
