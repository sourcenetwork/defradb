// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package protocol

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
)

type trackedStream struct {
	io.Reader
	closed bool
}

func (s *trackedStream) Close() error {
	s.closed = true
	return nil
}

// handlerHost records the stream handlers a protocol registers.
type handlerHost struct {
	client.Host
	handlers map[string]client.StreamHandler
}

func (h *handlerHost) SetStreamHandler(protocolID string, handler client.StreamHandler) {
	h.handlers[protocolID] = handler
}

type noopProcessor struct{}

func (noopProcessor) ProcessRequest(context.Context, PushLogRequest) (PushLogReply, error) {
	return PushLogReply{}, nil
}

func TestStreamHandlers_InvalidStreamReleasesResources(t *testing.T) {
	host := &handlerHost{handlers: map[string]client.StreamHandler{}}
	NewCommChannel[PushLogRequest, PushLogReply](host, "test", noopProcessor{})
	NewIdentityProtocol(host, func(context.Context, immutable.Option[string]) ([]byte, error) {
		return nil, nil
	})
	require.Len(t, host.handlers, 4)

	for protocolID, handler := range host.handlers {
		for name, reader := range map[string]io.Reader{
			"read error":        iotest.ErrReader(errors.New("read failed")),
			"invalid encoding":  bytes.NewReader([]byte{0xff}),
			"oversized message": bytes.NewReader(make([]byte, 16*1024*1024+1)),
		} {
			t.Run(protocolID+"/"+name, func(t *testing.T) {
				stream := &trackedStream{Reader: reader}
				handler(stream, "some peer ID")
				require.True(t, stream.closed, "rejected messages must release their stream resources")
			})
		}
	}
}
