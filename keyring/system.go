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

// systemKeyring is a keyring that utilizies the
// built in key management system of the OS.
type systemKeyring struct {
	// service is the service name to use when using the system keyring
	service string
}

func openSystemKeyring(service string) *systemKeyring {
	return &systemKeyring{
		service: service,
	}
}

func (s *systemKeyring) Set(name string, key []byte) error {
	enc := base64.StdEncoding.EncodeToString(key)
	return keyring.Set(s.service, name, enc)
}

func (s *systemKeyring) Get(name string) ([]byte, error) {
	enc, err := keyring.Get(s.service, name)
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

func (s *systemKeyring) Delete(user string) error {
	return keyring.Delete(s.service, user)
}
