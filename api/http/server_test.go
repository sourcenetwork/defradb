// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServerAndRunWithoutListener(t *testing.T) {
	s := NewServer(nil, WithAddress(":3131"))
	if ok := assert.NotNil(t, s); ok {
		assert.Equal(t, errNoListener, s.Run())
	}
}

func TestNewServerAndRunWithListenerAndInvalidPort(t *testing.T) {
	ctx := context.Background()
	s := NewServer(nil, WithAddress(":303000"))
	if ok := assert.NotNil(t, s); ok {
		assert.Error(t, s.Listen(ctx))
	}
}

func TestNewServerAndRunWithListenerAndValidPort(t *testing.T) {
	ctx := context.Background()
	serverRunning := make(chan struct{})
	serverDone := make(chan struct{})
	s := NewServer(nil, WithAddress(":3131"))
	go func() {
		close(serverRunning)
		err := s.Listen(ctx)
		assert.NoError(t, err)
		err = s.Run()
		assert.ErrorIs(t, http.ErrServerClosed, err)
		defer close(serverDone)
	}()

	<-serverRunning

	s.Shutdown(context.Background())

	<-serverDone
}

func TestNewServerWithoutOptions(t *testing.T) {
	s := NewServer(nil)
	assert.Equal(t, "localhost:9181", s.Addr)
	assert.Equal(t, []string(nil), s.options.allowedOrigins)
}

func TestNewServerWithAddress(t *testing.T) {
	s := NewServer(nil, WithAddress("localhost:9999"))
	assert.Equal(t, "localhost:9999", s.Addr)
}

func TestNewServerWithAllowedOrigins(t *testing.T) {
	s := NewServer(nil, WithAllowedOrigins("https://source.network", "https://app.source.network"))
	assert.Equal(t, []string{"https://source.network", "https://app.source.network"}, s.options.allowedOrigins)
}

func TestNewServerWithPeerID(t *testing.T) {
	s := NewServer(nil, WithPeerID("12D3KooWFpi6VTYKLtxUftJKEyfX8jDfKi8n15eaygH8ggfYFZbR"))
	assert.Equal(t, "12D3KooWFpi6VTYKLtxUftJKEyfX8jDfKi8n15eaygH8ggfYFZbR", s.options.peerID)
}
