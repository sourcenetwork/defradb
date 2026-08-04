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

//go:build !js

package action

import (
	"context"
	"fmt"

	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/errors"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/kms"
	"github.com/sourcenetwork/defradb/node"
	changeDetector "github.com/sourcenetwork/defradb/tests/change_detector"
	"github.com/sourcenetwork/defradb/tests/clients"
	"github.com/sourcenetwork/defradb/tests/clients/external"
	"github.com/sourcenetwork/defradb/tests/integration/version"
	"github.com/sourcenetwork/defradb/tests/state"
)

func createBadgerEncryptionKey() error {
	if !BadgerEncryption {
		return nil
	}
	encryptionKeyOnce.Do(func() {
		encryptionKey, encryptionKeyErr = crypto.GenerateAES256()
	})
	return encryptionKeyErr
}

// setupNode returns the database implementation for the current
// testing state. The database type on the test state is used to
// select the datastore implementation to use.
//
// If ver is non-empty, the node runs as a separate process from a downloaded
// release binary of that version instead of natively in-process; opts is
// ignored in that case (the external process gets its own flags).
//
// Note: If the signature of this function is updated, don't forget to
// also update the function in [tests/integration/db_setup_js.go] otherwise
// the js client build may fail (the failure might not be obvious to find).
func SetupNode(
	s *state.State,
	identity immutable.Option[acpIdentity.Identity],
	cfg NodeSetupConfig,
	opts *options.NodeOptionsBuilder,
	ver string,
) (*state.NodeState, error) {
	if ver != "" {
		return setupExternalNode(s, ver)
	}

	if opts == nil {
		opts = DefaultNodeOpts()
	}
	opts.DB().SetEnableSigning(cfg.EnableSigning)
	if cfg.HTTP.HasValue() {
		applyHTTPOptions(opts, cfg.HTTP.Value())
	}

	if s.EnableSearchableEncryption {
		seKey, err := crypto.GenerateAES256()
		if err != nil {
			return nil, fmt.Errorf("failed to generate searchable encryption key: %w", err)
		}
		opts.DB().SetSearchableEncryptionKey(seKey)
	}

	err := createBadgerEncryptionKey()
	if err != nil {
		return nil, err
	}
	if BadgerEncryption && encryptionKey != nil {
		opts.Store().SetBadgerEncryptionKey(encryptionKey)
	}

	switch s.DocumentACPType {
	case state.LocalDocumentACPType:
		opts.DocumentACP().SetType(options.NodeLocalDocumentACPType)

	case state.SourceHubDocumentACPType:
		if s.DocumentACPOptions == nil {
			s.DocumentACPOptions, err = setupSourceHub(s, cfg.IsDocumentACPTest)
			require.NoError(s.T, err)
		}
		opts.DocumentACP().SetAll(*s.DocumentACPOptions)

	default:
		// no-op, use the `node` package default
	}

	var path string
	if s.DbType == BadgerFileType || s.DbType == LevelStoreType {
		if DatabaseDir != "" {
			// restarting database
			path = DatabaseDir
		} else if changeDetector.Enabled {
			// change detector
			path = changeDetector.DatabaseDir(s.T)
		} else {
			// default test case
			path = s.T.TempDir()
		}
		opts.Store().SetPath(path).
			DocumentACP().SetPath(path).
			NodeACP().SetPath(path)
	}

	switch s.DbType {
	case BadgerFileType:
		opts.Store().SetType(options.NodeBadgerStore)

	case BadgerIMType:
		opts.Store().SetType(options.NodeBadgerStore).SetBadgerInMemory(true)

	case DefraIMType:
		opts.Store().SetType(options.NodeMemoryStore)

	case LevelStoreType:
		opts.Store().SetType(options.NodeStoreType("level"))

	default:
		return nil, fmt.Errorf("invalid database type: %v", s.DbType)
	}

	if s.KMS == PubSubKMSType {
		opts.SetKMS(options.NodeKMSType(kms.PubSubServiceType))
	}

	if s.IsNetworkEnabled {
		opts.SetDisableP2P(false)
	}

	opts.SetEnableDevelopment(true)

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
	require.Nil(s.T, err)

	// A native node discovers its addresses through the in-process DB, which
	// bypasses the HTTP auth middleware. Routing this through the client would
	// send an unauthenticated request that NAC rejects.
	st, err := newNodeState(s, c, nodeObj.DB, path, false)
	if err != nil {
		return nil, err
	}
	return st, nil
}

// peerInfoProvider reads a node's listen addresses. It is satisfied by both the
// in-process DB (native nodes) and the HTTP client (external nodes).
type peerInfoProvider interface {
	PeerInfo(ctx context.Context, opts ...options.Enumerable[options.PeerInfoOptions]) ([]string, error)
}

// newNodeState builds a NodeState around a set-up client: it subscribes to the
// client's events and discovers the node's cached peer addresses via peers. It
// is shared by the native and external node setup paths.
func newNodeState(
	s *state.State,
	c clients.Client,
	peers peerInfoProvider,
	path string,
	isExternal bool,
) (*state.NodeState, error) {
	eventState, err := state.NewEventState(c.Events())
	require.NoError(s.T, err)

	st := &state.NodeState{
		Client:     c,
		Event:      eventState,
		P2P:        state.NewP2PState(),
		DbPath:     path,
		IsExternal: isExternal,
	}

	addresses, err := discoverPeerAddresses(s, peers, isExternal)
	if err != nil {
		return nil, err
	}
	st.CachedAddresses = addresses

	return st, nil
}

// discoverPeerAddresses reads the node's listen addresses via PeerInfo and
// strips the trailing /p2p/<peerID> so they can be reused as listen addresses
// on restart.
func discoverPeerAddresses(s *state.State, peers peerInfoProvider, isExternal bool) ([]string, error) {
	// Inject node identity to bypass NAC in order to be able to call PeerInfo,
	// otherwise when NAC is enabled we get an authorization error.
	//
	// A native node reads through the in-process DB, so it always gets the
	// identity. An external node reads over HTTP, and this runs before the node
	// is on s.Nodes so no bearer token is generated yet. An empty "Bearer "
	// header is rejected by some older released servers, so skip it for an
	// external node when the token is empty.
	nodeIdentity := NodeIdentity(s.CurrentSetupNodeID)
	peerInfoOpts := options.PeerInfo()
	identOption := getIdentityForRequestSpecificToNode(s, nodeIdentity, s.CurrentSetupNodeID)
	if identOption.HasValue() {
		tokenIdent, ok := identOption.Value().(acpIdentity.TokenIdentity)
		if !isExternal || !ok || tokenIdent.BearerToken() != "" {
			peerInfoOpts.SetIdentity(identOption.Value())
		}
	}

	addresses, err := peers.PeerInfo(s.Ctx, peerInfoOpts)
	if err != nil {
		return nil, err
	}

	// The addresses returned by PeerInfo include the /p2p/<peerID> part, but
	// the libp2p.ListenAddrStrings cannot include it, so we need to remove it
	// before caching the addresses on the state.
	return removePeerIDFromAddr(addresses)
}

// setupExternalNode starts a node as a separate OS process from a downloaded
// release binary of the given version, and wraps it in the same NodeState
// shape a native node would produce.
//
// If no release asset exists for this platform, the test is skipped (not
// failed) and a nil error is returned.
func setupExternalNode(s *state.State, ver string) (*state.NodeState, error) {
	path, skip, err := version.BinaryPath(s.Ctx, ver)
	if err != nil {
		return nil, err
	}
	if skip {
		s.T.Skipf("no %s release asset for this platform", ver)
		return nil, nil
	}

	w, err := external.NewWrapper(s.Ctx, s.T, path)
	if err != nil {
		return nil, err
	}

	// An external node has no in-process DB, so it discovers its addresses over
	// the HTTP client.
	return newNodeState(s, w, w, "", true)
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
