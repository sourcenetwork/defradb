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

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
)

// newClosableP2P builds the smallest P2P that Close can run against: the queue and its stop
// signal, plus the two subsystems Close shuts down after the workers.
func newClosableP2P(queueSize int) *P2P {
	return &P2P{
		ctx:              context.Background(),
		host:             &SimpleMockHost{},
		msgQueue:         make(chan queuedMessage, queueSize),
		msgQueueMaxBytes: 1 << 20,
		stopAccepting:    make(chan struct{}),
		batcher:          newPubsubBatcher("peerID", func(string, []byte) error { return nil }),
		processQueue:     newProcessQueue(),
	}
}

// The pubsub dispatcher keeps delivering until libp2p tears the subscription down, which Close
// does not do, so the handler runs after Close has begun. A send on a closed channel is a ready
// case in a select rather than a fallthrough to the default arm, so closing the queue to stop
// the workers panics whichever goroutine is mid-enqueue.
func TestEnqueueDuringCloseDoesNotPanic(t *testing.T) {
	p := newClosableP2P(4)

	msg, err := cbor.Marshal(protocol.PushLogRequest{DocID: "docID"})
	require.NoError(t, err)

	// Drains so the queue keeps accepting and the handler stays on the enqueue arm, which is
	// the arm that panics. Stopped by its own signal rather than by closing the queue, so the
	// only thing that can close it is the code under test.
	drained := make(chan struct{})
	stopDrain := make(chan struct{})
	go func() {
		defer close(drained)
		for {
			select {
			case <-stopDrain:
				return
			case <-p.msgQueue:
			}
		}
	}()

	var handlers sync.WaitGroup
	panicked := make(chan any, 1)
	stop := make(chan struct{})
	for range 4 {
		handlers.Add(1)
		go func() {
			defer handlers.Done()
			defer func() {
				if r := recover(); r != nil {
					select {
					case panicked <- r:
					default:
					}
				}
			}()
			for {
				select {
				case <-stop:
					return
				default:
				}
				_, _ = p.pubSubMessageHandler("sender", "topic", msg)
			}
		}()
	}

	time.Sleep(20 * time.Millisecond)
	p.Close()
	close(stop)
	handlers.Wait()
	close(stopDrain)
	<-drained

	select {
	case r := <-panicked:
		t.Fatalf("a dispatcher enqueueing during Close panicked: %v", r)
	default:
	}
}

// Nothing closes the queue any more, so the stop signal is the only thing that ends a worker.
func TestWorkersExitOnStopAccepting(t *testing.T) {
	p := newClosableP2P(1)

	done := make(chan struct{})
	p.msgWorkers.Add(1)
	go func() {
		defer close(done)
		p.processMessageWorker()
	}()

	close(p.stopAccepting)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("worker did not exit on stopAccepting")
	}
}
