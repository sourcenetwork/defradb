// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package message

import "github.com/sourcenetwork/defradb/errors"

var (
	ErrResponseTimeout      = errors.New("timeout waiting for response")
	ErrPubkeyPeerIDMismatch = errors.New("pubkey mismatch peerID")
	ErrInvalidSignature     = errors.New("invalid signature")
)
