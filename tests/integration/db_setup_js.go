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
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"
)

// setupNode returns the database implementation for the current
// testing state. The database type on the test state is used to
// select the datastore implementation to use.
func setupNode(
	s *state.State,
	identity immutable.Option[acpIdentity.Identity],
	isNACEnabled bool,
	testCase TestCase,
	opts ...node.Option,
) (*state.NodeState, error) {
	opts = append(defaultNodeOpts(), opts...)
	opts = append(opts, db.WithEnabledSigning(testCase.EnableSigning))
	opts = append(opts, node.WithLensRuntime(node.JSLensRuntime))
	opts = append(opts, node.WithEnableNodeACP(isNACEnabled))
	// Note: Since we are hard-coding to run with badger in-mem only, we have a function that
	// handles some edge-cases by skipping js client testing when a db type is something else.
	// If this hard-coding is changed in future, don't forget to tweak the following func:
	// [skipJSClientIfUnsupportedDBType]
	opts = append(opts, node.WithBadgerInMemory(true))

	switch documentACPType {
	case LocalDocumentACPType:
		opts = append(opts, node.WithDocumentACPType(node.LocalDocumentACPType))

	case SourceHubDocumentACPType:
		if len(s.DocumentACPOptions) == 0 {
			var err error
			s.DocumentACPOptions, err = setupSourceHub(s)
			require.NoError(s.T, err)
		}

		opts = append(opts, node.WithDocumentACPType(node.SourceHubDocumentACPType))
		for _, opt := range s.DocumentACPOptions {
			opts = append(opts, opt)
		}

	default:
		// no-op, use the `node` package default
	}

	node, err := node.New(s.Ctx, opts...)
	if err != nil {
		return nil, err
	}
	s.Ctx = acpIdentity.WithContext(s.Ctx, identity)
	err = node.Start(s.Ctx)
	resetStateContext(s)
	if err != nil {
		return nil, err
	}
	c, err := setupClient(s, node)
	if err != nil {
		return nil, err
	}
	eventState, err := state.NewEventState(c.Events())
	if err != nil {
		return nil, err
	}
	return &state.NodeState{
		Client: c,
		Event:  eventState,
		P2P:    state.NewP2PState(),
	}, nil
}
