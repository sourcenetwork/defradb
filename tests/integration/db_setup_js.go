// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

import (
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/node"
)

// setupNode returns the database implementation for the current
// testing state. The database type on the test state is used to
// select the datastore implementation to use.
func setupNode(s *state, opts ...node.Option) (*nodeState, error) {
	opts = append(defaultNodeOpts(), opts...)
	opts = append(opts, db.WithEnabledSigning(s.testCase.EnableSigning))
	opts = append(opts, node.WithBadgerInMemory(true))

	node, err := node.New(s.ctx, opts...)
	if err != nil {
		return nil, err
	}
	err = node.Start(s.ctx)
	if err != nil {
		return nil, err
	}
	c, err := setupClient(s, node)
	if err != nil {
		return nil, err
	}
	eventState, err := newEventState(c.Events())
	if err != nil {
		return nil, err
	}
	return &nodeState{
		Client: c,
		event:  eventState,
		p2p:    newP2PState(),
	}, nil
}
