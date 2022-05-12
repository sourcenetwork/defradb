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

func TestNewServerAndListen(t *testing.T) {
	s := NewServer(nil)
	if ok := assert.NotNil(t, s); ok {
		assert.Error(t, s.Listen(":303000"))
	}

	serverRunning := make(chan struct{})
	serverDone := make(chan struct{})
	go func() {
		close(serverRunning)
		err := s.Listen(":3131")
		assert.ErrorIs(t, http.ErrServerClosed, err)
		defer close(serverDone)
	}()

	<-serverRunning

	s.Shutdown(context.Background())

	<-serverDone
}
