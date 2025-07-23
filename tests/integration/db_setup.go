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
	s *state,
	identity immutable.Option[acpIdentity.Identity],
	isAACEnabled bool,
	opts ...node.Option,
) (*nodeState, error) {
	opts = append(defaultNodeOpts(), opts...)
	opts = append(opts, db.WithEnabledSigning(s.testCase.EnableSigning))
	opts = append(opts, node.WithEnableAdminACP(isAACEnabled))

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
		if len(s.documentACPOptions) == 0 {
			s.documentACPOptions, err = setupSourceHub(s)
			require.NoError(s.t, err)
		}

		opts = append(opts, node.WithDocumentACPType(node.SourceHubDocumentACPType))
		for _, opt := range s.documentACPOptions {
			opts = append(opts, opt)
		}

	default:
		// no-op, use the `node` package default
	}

	var path string
	switch s.dbt {
	case BadgerIMType:
		opts = append(opts, node.WithBadgerInMemory(true))

	case BadgerFileType:
		switch {
		case databaseDir != "":
			// restarting database
			path = databaseDir

		case changeDetector.Enabled:
			// change detector
			path = changeDetector.DatabaseDir(s.t)

		default:
			// default test case
			path = s.t.TempDir()
		}

		opts = append(
			opts,
			node.WithStorePath(path),
			node.WithDocumentACPPath(path),
			node.WithAdminACPPath(path),
		)

	case DefraIMType:
		opts = append(opts, node.WithStoreType(node.MemoryStore))

	default:
		return nil, fmt.Errorf("invalid database type: %v", s.dbt)
	}

	if s.kms == PubSubKMSType {
		opts = append(opts, node.WithKMS(kms.PubSubServiceType))
	}

	netOpts := make([]netConfig.NodeOpt, 0)
	for _, opt := range opts {
		if opt, ok := opt.(netConfig.NodeOpt); ok {
			netOpts = append(netOpts, opt)
		}
	}

	if s.isNetworkEnabled {
		opts = append(opts, node.WithDisableP2P(false))
	}

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
	require.Nil(s.t, err)

	eventState, err := newEventState(c.Events())
	require.NoError(s.t, err)

	st := &nodeState{
		Client:  c,
		event:   eventState,
		p2p:     newP2PState(),
		dbPath:  path,
		netOpts: netOpts,
	}

	if node.Peer != nil {
		st.peerInfo = node.Peer.PeerInfo()
	}

	return st, nil
}
