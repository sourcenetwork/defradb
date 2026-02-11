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

	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/internal/kms"
	"github.com/sourcenetwork/defradb/node"
	changeDetector "github.com/sourcenetwork/defradb/tests/change_detector"
	"github.com/sourcenetwork/defradb/tests/state"
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
	opts ...node.Option,
) (*state.NodeState, error) {
	opts = append(defaultNodeOpts(), opts...)
	opts = append(opts, db.WithEnabledSigning(testCase.EnableSigning))

	if s.EnableSearchableEncryption {
		seKey, err := crypto.GenerateAES256()
		if err != nil {
			return nil, fmt.Errorf("failed to generate searchable encryption key: %w", err)
		}
		opts = append(opts, db.WithSearchableEncryptionKey(seKey))
	}

	err := createBadgerEncryptionKey()
	if err != nil {
		return nil, err
	}
	if badgerEncryption && encryptionKey != nil {
		opts = append(opts, node.WithBadgerEncryptionKey(encryptionKey))
	}

	switch s.DocumentACPType {
	case state.LocalDocumentACPType:
		opts = append(opts, node.WithDocumentACPType(node.LocalDocumentACPType))

	case state.SourceHubDocumentACPType:
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
	if s.DbType == BadgerFileType || s.DbType == LevelStoreType {
		if databaseDir != "" {
			// restarting database
			path = databaseDir
		} else if changeDetector.Enabled {
			// change detector
			path = changeDetector.DatabaseDir(s.T)
		} else {
			// default test case
			path = s.T.TempDir()
		}
		opts = append(opts,
			node.WithStorePath(path),
			node.WithDocumentACPPath(path),
			node.WithNodeACPPath(path),
		)
	}

	switch s.DbType {
	case BadgerFileType:
		opts = append(opts, node.WithStoreType(node.BadgerStore))

	case BadgerIMType:
		opts = append(opts, node.WithStoreType(node.BadgerStore), node.WithBadgerInMemory(true))

	case DefraIMType:
		opts = append(opts, node.WithStoreType(node.MemoryStore))

	case LevelStoreType:
		opts = append(opts, node.WithStoreType(node.LevelStore))

	default:
		return nil, fmt.Errorf("invalid database type: %v", s.DbType)
	}

	if s.KMS == PubSubKMSType {
		opts = append(opts, node.WithKMS(kms.PubSubServiceType))
	}

	netOpts := getP2POptions(opts)

	if s.IsNetworkEnabled {
		opts = append(opts, node.WithDisableP2P(false))
	}

	nodeObj, err := node.New(s.Ctx, opts...)
	if err != nil {
		return nil, err
	}

	ctx := acpIdentity.WithContext(s.Ctx, identity)
	err = nodeObj.Start(ctx)

	if err != nil {
		return nil, err
	}

	c, err := setupClient(s, nodeObj)
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

	var addresses []string

	// Inject node identity to bypass NAC inorder to be able to call [PeerInfo] operation,
	// otherwise when NAC is enabled, we will get authorization error.
	nodeIdentity := NodeIdentity(s.CurrentSetupNodeID)
	peerInfoOpts := options.PeerInfo()
	identOption := getIdentityForRequestSpecificToNode(s, nodeIdentity, s.CurrentSetupNodeID)
	if identOption.HasValue() {
		peerInfoOpts.SetIdentity(identOption.Value())
	}
	addresses, err = nodeObj.DB.PeerInfo(s.Ctx, peerInfoOpts)
	require.NoError(s.T, err)

	// The addresses returned by PeerInfo include the /p2p/<peerID> part, but
	// the libp2p.ListenAddrStrings cannot include it, so we need to remove it
	// before caching the addresses on the state.
	addresses, err = removePeerIDFromAddr(addresses)
	require.NoError(s.T, err)
	st.CachedAddresses = addresses

	return st, nil
}

func removePeerIDFromAddr(addr []string) ([]string, error) {
	addrs := make([]string, len(addr))
	for i, a := range addr {
		justAddr, err := removePeerID(a)
		if err != nil {
			return nil, err
		}
		addrs[i] = justAddr
	}
	return addrs, nil
}

func removePeerID(addr string) (string, error) {
	maddrWithID, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return "", err
	}
	justAddr, p2ppart := multiaddr.SplitLast(maddrWithID)
	if p2ppart == nil || p2ppart.Protocol().Code != multiaddr.P_P2P {
		return "", errors.New("address does not contain a /p2p/ part")
	}
	return justAddr.String(), nil
}
