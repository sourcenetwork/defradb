// Copyright 2026 Democratized Data Foundation
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
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
)

type dummyHost struct {
	SimpleMockHost
}

func (d *dummyHost) Connect(ctx context.Context, addresses []string) error {
	return nil
}

type dummyPushProtocol struct{}

func (d *dummyPushProtocol) SendRequest(
	ctx context.Context,
	req protocol.PushLogRequest,
	peer string,
) (protocol.PushLogReply, error) {
	return protocol.PushLogReply{}, nil
}

func TestPushLogToReplicators_ConcurrentMapAccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p := &P2P{
		ctx:                ctx,
		host:               &dummyHost{},
		replicators:        make(map[string]map[string][]string),
		replicatorProtocol: &dummyPushProtocol{},
	}

	colID := "col1"
	p.replicators[colID] = map[string][]string{
		"peer1": {"addr1"},
		"peer2": {"addr2"},
	}

	var wg sync.WaitGroup
	// Concurrently read and push logs
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			p.pushLogToReplicators(event.Update{
				CollectionID: colID,
				DocID:        fmt.Sprintf("doc-%d", idx),
			})
		}(i)
	}

	// Concurrently mutate replicators
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			p.updateReplicators(
				ctx,
				fmt.Sprintf("peer-%d", idx),
				[]string{fmt.Sprintf("addr-%d", idx)},
				map[string]struct{}{colID: {}},
			)
		}(i)
	}

	wg.Wait()
}

func TestHandleCompletedReplicatorRetry_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := &P2P{}
	err := p.handleCompletedReplicatorRetry(ctx, "peer1", true)
	assert.ErrorIs(t, err, context.Canceled)
}
