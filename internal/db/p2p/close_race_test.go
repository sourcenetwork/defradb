// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// The pubsub dispatcher keeps delivering until libp2p tears the subscription down, which
// Close does not do, so the handler runs after Close has begun. Enqueueing has to stop
// without the queue being closed underneath it.
func TestEnqueueDuringCloseDoesNotPanic(t *testing.T) {
	p := &P2P{
		ctx:           context.Background(), // never cancelled, as in the host
		msgQueue:      make(chan queuedMessage, 4),
		stopAccepting: make(chan struct{}),
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		deadline := time.After(200 * time.Millisecond)
		for {
			select {
			case <-deadline:
				return
			default:
			}
			// Mirrors the handler's enqueue arm. A closed queue would panic here even
			// though the select has a default case.
			select {
			case p.msgQueue <- queuedMessage{size: 1}:
			case <-p.stopAccepting:
			case <-p.ctx.Done():
			default:
			}
		}
	}()

	time.Sleep(20 * time.Millisecond)
	require.NotPanics(t, func() { close(p.stopAccepting) }, "shutdown must not panic a live dispatcher")
	wg.Wait()
}

// Workers must exit on the stop signal, since nothing closes the queue any more.
func TestWorkersExitOnStopAccepting(t *testing.T) {
	p := &P2P{
		ctx:           context.Background(),
		msgQueue:      make(chan queuedMessage, 1),
		stopAccepting: make(chan struct{}),
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-p.stopAccepting:
				return
			case <-p.msgQueue:
			}
		}
	}()

	close(p.stopAccepting)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("worker did not exit on stopAccepting")
	}
}
