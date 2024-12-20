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
	"github.com/zalando/go-keyring"

	"github.com/sourcenetwork/defradb/errors"
)

var (
	// ErrNotFound is returned when a keyring item is not found.
	ErrNotFound                 = keyring.ErrNotFound
	ErrSystemKeyringListInvoked = errors.New("listing keys is not supported by OS keyring")
)
