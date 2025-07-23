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
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/immutable"
)

// setupNode returns the database implementation for the current
// testing state. The database type on the test state is used to
// select the datastore implementation to use.
func setupNode(
	s *state,
	identity immutable.Option[acpIdentity.Identity],
	isAACEnabled bool,
	opts ...node.Option,
) (*nodeState, error) {
	opts = append(defaultNodeOpts(), opts...)
	opts = append(opts, db.WithEnabledSigning(s.testCase.EnableSigning))
	opts = append(opts, node.WithLensRuntime(node.JSLensRuntime))
	opts = append(opts, node.WithEnableAdminACP(isAACEnabled))
	// Note: Since we are hard-coding to run with badger in-mem only, we have a function that
	// handles some edge-cases by skipping js client testing when a db type is something else.
	// If this hard-coding is changed in future, don't forget to tweak the following func:
	// [skipJSClientIfUnsupportedDBType]
	opts = append(opts, node.WithBadgerInMemory(true))

	node, err := node.New(s.ctx, opts...)
	if err != nil {
		return nil, err
	}
	s.ctx = acpIdentity.WithContext(s.ctx, identity)
	err = node.Start(s.ctx)
	resetStateContext(s)
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
