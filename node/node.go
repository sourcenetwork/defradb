package node

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-peerstore/pstoreds"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/textileio/go-libp2p-pubsub-rpc/finalizer"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/net"
)

/*

Package node is responsible for interfacing a given DefraDB instance with
a networked peer instance and GRPC server.

Basically it combines db/DB, net/Peer, and net/Server into a single Node
object.
*/

var (
	log = logging.Logger("node")
)

type Node struct {
	// embed the DB interface into the node
	client.DB

	*net.Peer

	host     host.Host
	pubsub   *pubsub.PubSub
	litepeer *ipfslite.Peer

	ctx    context.Context
	cancel context.CancelFunc
}

// NewNode creates a new network node instance of DefraDB, wired into Libp2p
func NewNode(ctx context.Context, db client.DB, opts ...NodeOpt) (*Node, error) {
	// merge all the options args together
	var options Options
	for _, opt := range append(opts, DefaultOpts()) {
		if err := opt(&options); err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	fin := finalizer.NewFinalizer()

	// create our peerstore from the underlying defra rootstore
	// prefixed with "p2p"
	rootstore := db.Root()
	pstore := namespace.Wrap(rootstore, ds.NewKey("peers"))
	peerstore, err := pstoreds.NewPeerstore(ctx, pstore, pstoreds.DefaultOpts())
	if err != nil {
		return nil, fin.Cleanup(err)
	}
	fin.Add(peerstore)

	hostKey, err := getHostKey(options.RepoPath)
	if err != nil {
		return nil, fin.Cleanup(err)
	}

	libp2pOpts := []libp2p.Option{
		libp2p.Peerstore(peerstore),
		libp2p.ConnectionManager(options.ConnManager),
		libp2p.DisableRelay(), // @todo: Possibly bind this to an Option
	}

	h, d, err := ipfslite.SetupLibp2p(
		ctx,
		hostKey,
		nil,
		options.ListenAddrs,
		rootstore,
		libp2pOpts...,
	)
	log.Info("Created LibP2P host with Peer ID ", h.ID())
	log.Info(" -> listening on ", options.ListenAddrs)
	if err != nil {
		return nil, fin.Cleanup(err)
	}

	bstore := db.DAGstore()
	lite, err := ipfslite.New(ctx, rootstore, bstore, h, d, nil)
	if err != nil {
		return nil, fin.Cleanup(err)
	}

	var ps *pubsub.PubSub
	if options.EnablePubSub {
		ps, err = pubsub.NewGossipSub(
			ctx,
			h,
			pubsub.WithPeerExchange(true),
			pubsub.WithFloodPublish(true),
		)
		if err != nil {
			return nil, fin.Cleanup(err)
		}
	}

	peer, err := net.NewPeer(
		ctx,
		db,
		h,
		ps,
		lite,
		options.GRPCServerOptions,
		options.GRPCDialOptions,
	)
	if err != nil {
		return nil, fin.Cleanup(err)
	}

	return &Node{
		Peer:     peer,
		host:     h,
		pubsub:   ps,
		DB:       db,
		litepeer: lite,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

func (n *Node) Boostrap(addrs []peer.AddrInfo) {
	n.litepeer.Bootstrap(addrs)
}

// replace with proper keystore
func getHostKey(keypath string) (crypto.PrivKey, error) {
	// If a local datastore is used, the key is written to a file
	pth := filepath.Join(keypath, "key")
	_, err := os.Stat(pth)
	if os.IsNotExist(err) {
		key, bytes, err := newHostKey()
		if err != nil {
			return nil, err
		}
		if err := os.MkdirAll(keypath, os.ModePerm); err != nil {
			return nil, err
		}
		if err = ioutil.WriteFile(pth, bytes, 0400); err != nil {
			return nil, err
		}
		return key, nil
	} else if err != nil {
		return nil, err
	} else {
		bytes, err := ioutil.ReadFile(pth)
		if err != nil {
			return nil, err
		}
		return crypto.UnmarshalPrivateKey(bytes)
	}
}

func newHostKey() (crypto.PrivKey, []byte, error) {
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
	if err != nil {
		return nil, nil, err
	}
	key, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}
	return priv, key, nil
}
