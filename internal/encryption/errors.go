// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package encryption

import (
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errNoStorageProvided     string = "no storage provided"
	errContextHasNoEncryptor string = "context has no encryptor"
)

var (
	ErrNoStorageProvided     = errors.New(errNoStorageProvided)
	ErrContextHasNoEncryptor = errors.New(errContextHasNoEncryptor)
)
