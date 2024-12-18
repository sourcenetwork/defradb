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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSystemKeyringListThrowsError(t *testing.T) {
	service := "test-service"
	systemKeyring := OpenSystemKeyring(service)

	keys, err := systemKeyring.List()

	assert.Nil(t, keys, "keys should be nil when List is called")
	assert.Error(t, err, "an error should be returned when List is called")
	assert.Equal(t, ErrSystemKeyringListInvoked, err, "the error should be ErrSystemKeyringListInvoked")
}
