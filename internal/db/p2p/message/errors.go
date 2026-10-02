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

import (
	"fmt"

	"github.com/sourcenetwork/defradb/errors"
)

const errPeerRejected = "request rejected by peer"

var (
	ErrPeerRejected         = errors.New(errPeerRejected)
	ErrResponseTimeout      = errors.New("timeout waiting for response")
	ErrPubkeyPeerIDMismatch = errors.New("pubkey mismatch peerID")
	ErrInvalidSignature     = errors.New("invalid signature")
	ErrResponseType         = errors.New("unexpected response type")
	ErrMessageTooLarge      = errors.New("message exceeds maximum size")
)

func NewErrResponseType(expected, actual any) error {
	return errors.WithStack(
		ErrResponseType,
		errors.NewKV("Expected", fmt.Sprintf("%T", expected)),
		errors.NewKV("Actual", fmt.Sprintf("%T", actual)),
	)
}

// NewErrPeerRejected returns an error for a request that reached the peer, but that the peer
// failed to process.
func NewErrPeerRejected(peerErrMessage string) error {
	return errors.Wrap(errPeerRejected, errors.New(peerErrMessage))
}
