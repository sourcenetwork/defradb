// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package mobile

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/sourcenetwork/immutable"
	_ "golang.org/x/mobile/bind" // keep golang.org/x/mobile in go.mod; go mod tidy would otherwise remove it

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	idb "github.com/sourcenetwork/defradb/internal/db"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/node"
)

// DB wraps a DefraDB node and stores the caller's identity for use in all operations.
type DB struct {
	node     *node.Node
	identity immutable.Option[acpIdentity.Identity]
}

// Open creates and starts a new DefraDB node at the given path.
//
// If privateKeyHex is non-empty, it is parsed as a secp256k1 private key and used
// to build the node identity. Otherwise the node runs without an identity.
func Open(path string, privateKeyHex string) (*DB, error) {
	ctx := context.Background()

	nodeOpt := options.Node().
		SetDisableP2P(false).
		SetDisableAPI(true).
		Store().SetType(options.NodeBadgerStore).SetPath(path).
		Node()

	var identity immutable.Option[acpIdentity.Identity]

	if privateKeyHex != "" {
		privKey, err := crypto.PrivateKeyFromString(crypto.KeyTypeSecp256k1, privateKeyHex)
		if err != nil {
			return nil, err
		}
		ident, err := acpIdentity.FromPrivateKey(privKey)
		if err != nil {
			return nil, err
		}
		identity = immutable.Some[acpIdentity.Identity](ident)
		nodeOpt.DB().SetNodeIdentity(ident)
	}

	n, err := node.New(ctx, nodeOpt)
	if err != nil {
		return nil, err
	}
	if err := n.Start(ctx); err != nil {
		return nil, err
	}
	return &DB{node: n, identity: identity}, nil
}

// Close shuts down the node.
func (db *DB) Close() error {
	return db.node.Close(context.Background())
}

// context returns a context with the DB's identity injected.
func (db *DB) context() context.Context {
	return iIdentity.WithContext(context.Background(), db.identity)
}

// contextWithTxn returns a context with the DB's identity and a transaction injected.
func (db *DB) contextWithTxn(txn client.Txn) context.Context {
	ctx := iIdentity.WithContext(context.Background(), db.identity)
	return idb.InitContext(ctx, txn)
}

// marshalJSON serializes v to a JSON string.
func marshalJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// setOptIdentity sets the identity on any option builder that has a SetIdentity method,
// using reflection to remain generic across all option builder types.
func setOptIdentity[B any](opt B, id immutable.Option[acpIdentity.Identity]) {
	if !id.HasValue() {
		return
	}
	v := reflect.ValueOf(opt)
	m := v.MethodByName("SetIdentity")
	if m.IsValid() {
		m.Call([]reflect.Value{reflect.ValueOf(id.Value())})
	}
}
