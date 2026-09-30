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
	"bytes"
	"errors"
	"io"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/require"
)

type trackedStream struct {
	io.Reader
	closed bool
}

func (s *trackedStream) Close() error {
	s.closed = true
	return nil
}

func TestReceive_InvalidStreamReleasesResources(t *testing.T) {
	for name, reader := range map[string]io.Reader{
		"read error":        iotest.ErrReader(errors.New("read failed")),
		"invalid encoding":  bytes.NewReader([]byte{0xff}),
		"oversized message": bytes.NewReader(make([]byte, maxMessageSize+1)),
	} {
		t.Run(name, func(t *testing.T) {
			stream := &trackedStream{Reader: reader}
			err := Receive(stream, "some peer ID", nil, &MetaData{})
			require.Error(t, err)
			require.True(t, stream.closed, "rejected messages must release their stream resources")
		})
	}
}

func TestReceive_StreamLargerThanMax_ReturnsErrMessageTooLarge(t *testing.T) {
	stream := bytes.NewReader(make([]byte, maxMessageSize+1))
	err := Receive(stream, "some peer ID", nil, &MetaData{})
	require.ErrorIs(t, err, ErrMessageTooLarge)
}
