// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package tests

import (
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client/options"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/state"
)

// setupNode returns the database implementation for the current
// testing state. The database type on the test state is used to
// select the datastore implementation to use.
func setupNode(
	s *state.State,
	identity immutable.Option[acpIdentity.Identity],
	testCase TestCase,
	opts *options.NodeOptionsBuilder,
) (*state.NodeState, error) {
	if opts == nil {
		opts = defaultNodeOpts()
	}
	opts.DB().
		SetEnableSigning(testCase.EnableSigning).
		SetLensRuntime(options.NodeJSLensRuntime)
	// Note: Since we are hard-coding to run with badger in-mem only, we have a function that
	// handles some edge-cases by skipping js client testing when a db type is something else.
	// If this hard-coding is changed in future, don't forget to tweak the following func:
	// [skipJSClientIfUnsupportedDBType]
	opts.Store().SetBadgerInMemory(true)

	switch documentACPType {
	case state.LocalDocumentACPType:
		opts.DocumentACP().SetType(options.NodeLocalDocumentACPType)

	case state.SourceHubDocumentACPType:
		if s.DocumentACPOptions == nil {
			var err error
			s.DocumentACPOptions, err = setupSourceHub(s)
			require.NoError(s.T, err)
		}
		opts.DocumentACP().
			SetType(options.NodeSourceHubDocumentACPType).
			SetAll(*s.DocumentACPOptions)

	default:
		// no-op, use the `node` package default
	}

	nodeObj, err := node.New(s.Ctx, opts)
	if err != nil {
		return nil, err
	}
	ctx := iIdentity.WithContext(s.Ctx, identity)
	err = nodeObj.Start(ctx)
	if err != nil {
		return nil, err
	}
	c, err := setupClient(s, nodeObj)
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
