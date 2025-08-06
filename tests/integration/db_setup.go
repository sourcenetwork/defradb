// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !js

package tests

import (
	"fmt"

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/internal/kms"
	netConfig "github.com/sourcenetwork/defradb/net/config"
	"github.com/sourcenetwork/defradb/node"
	changeDetector "github.com/sourcenetwork/defradb/tests/change_detector"
	"github.com/sourcenetwork/defradb/tests/state"

	"github.com/stretchr/testify/require"
)

func createBadgerEncryptionKey() error {
	if !badgerEncryption || encryptionKey != nil {
		return nil
	}
	key, err := crypto.GenerateAES256()
	if err != nil {
		return err
	}
	encryptionKey = key
	return nil
}

// setupNode returns the database implementation for the current
// testing state. The database type on the test state is used to
// select the datastore implementation to use.
//
// Note: If the signature of this function is updated, don't forget to
// also update the function in [tests/integration/db_setup_js.go] otherwise
// the js client build may fail (the failure might not be obvious to find).
func setupNode(
	s *state.State,
	identity immutable.Option[acpIdentity.Identity],
	testCase TestCase,
	enableNAC bool,
	opts ...node.Option,
) (*state.NodeState, error) {
	opts = append(defaultNodeOpts(), opts...)
	opts = append(opts, db.WithEnabledSigning(testCase.EnableSigning))

	err := createBadgerEncryptionKey()
	if err != nil {
		return nil, err
	}
	if badgerEncryption && encryptionKey != nil {
		opts = append(opts, node.WithBadgerEncryptionKey(encryptionKey))
	}

	switch documentACPType {
	case LocalDocumentACPType:
		opts = append(opts, node.WithDocumentACPType(node.LocalDocumentACPType))

	case SourceHubDocumentACPType:
		if len(s.DocumentACPOptions) == 0 {
			s.DocumentACPOptions, err = setupSourceHub(s, testCase)
			require.NoError(s.T, err)
		}

		opts = append(opts, node.WithDocumentACPType(node.SourceHubDocumentACPType))
		for _, opt := range s.DocumentACPOptions {
			opts = append(opts, opt)
		}

	default:
		// no-op, use the `node` package default
	}

	var path string
	switch s.DbType {
	case BadgerIMType:
		opts = append(opts, node.WithBadgerInMemory(true))

	case BadgerFileType:
		switch {
		case databaseDir != "":
			// restarting database
			path = databaseDir

		case changeDetector.Enabled:
			// change detector
			path = changeDetector.DatabaseDir(s.T)

		default:
			// default test case
			path = s.T.TempDir()
		}

		opts = append(
			opts,
			node.WithStorePath(path),
			node.WithDocumentACPPath(path),
			node.WithNodeACPPath(path),
		)

	case DefraIMType:
		opts = append(opts, node.WithStoreType(node.MemoryStore))

	default:
		return nil, fmt.Errorf("invalid database type: %v", s.DbType)
	}

	if s.KMS == PubSubKMSType {
		opts = append(opts, node.WithKMS(kms.PubSubServiceType))
	}

	netOpts := make([]netConfig.NodeOpt, 0)
	for _, opt := range opts {
		if opt, ok := opt.(netConfig.NodeOpt); ok {
			netOpts = append(netOpts, opt)
		}
	}

	if s.IsNetworkEnabled {
		opts = append(opts, node.WithDisableP2P(false))
	}

	nodeObj, err := node.New(s.Ctx, opts...)
	if err != nil {
		return nil, err
	}

	s.Ctx = acpIdentity.WithContext(s.Ctx, identity)
	err = nodeObj.Start(s.Ctx)

	if err != nil {
		return nil, err
	}

	c, err := setupClient(s, nodeObj, enableNAC)
	resetStateContext(s)
	require.Nil(s.T, err)

	eventState, err := state.NewEventState(c.Events())
	require.NoError(s.T, err)

	st := &state.NodeState{
		Client:  c,
		Event:   eventState,
		P2P:     state.NewP2PState(),
		DbPath:  path,
		NetOpts: netOpts,
	}

	if nodeObj.Peer != nil {
		st.AddrInfo = nodeObj.Peer.PeerInfo()
	}

	return st, nil
}
