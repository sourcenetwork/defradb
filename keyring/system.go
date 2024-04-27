// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keyring

import (
	"encoding/base64"

	"github.com/zalando/go-keyring"
)

var _ Keyring = (*systemKeyring)(nil)

type systemKeyring struct{}

func newSystemKeyring() *systemKeyring {
	return &systemKeyring{}
}

func (systemKeyring) Set(name string, key []byte) error {
	enc := base64.StdEncoding.EncodeToString(key)
	return keyring.Set(service, name, enc)
}

func (systemKeyring) Get(name string) ([]byte, error) {
	enc, err := keyring.Get(service, name)
	if err != nil {
		return nil, err
	}
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(enc)))
	n, err := base64.StdEncoding.Decode(dst, []byte(enc))
	if err != nil {
		return nil, err
	}
	return dst[:n], nil
}

func (systemKeyring) Delete(user string) error {
	return keyring.Delete(service, user)
}
