// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cbindings

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/node"
)

const (
	KeyTypeEd25519   = "ed25519"
	KeyTypeSecp256k1 = "secp256k1"
)

type GoCResult struct {
	Status int
	Error  string
	Value  string
}

type GoCOptions struct {
	TxID         uint64
	Version      string
	CollectionID string
	Name         string
	Identity     string
	GetInactive  int
}

type GoNodeInitOptions struct {
	DbPath                   string
	ListeningAddresses       string
	ReplicatorRetryIntervals string
	Peers                    string
	IdentityKeyType          string
	IdentityPrivateKey       string
	InMemory                 int
	DisableP2P               int
	DisableAPI               int
	MaxTransactionRetries    int
}

// returnGoC is a helper function that wraps a status, error, and value into a return object
func returnGoC(status int, errortext string, valuetext string) GoCResult {
	return GoCResult{
		Status: status,
		Error:  errortext,
		Value:  valuetext,
	}
}

// marshalJSONToGoCResult is a helper function that marshals an interface into a return object
func marshalJSONToGoCResult(value any) GoCResult {
	dataJSON, err := json.Marshal(value)
	if err != nil {
		return returnGoC(1, fmt.Sprintf(errMarshallingJSON, err), "")
	}
	return returnGoC(0, "", string(dataJSON))
}

// contextWithIdentity is a helper function that attaches identity to a context
func contextWithIdentity(ctx context.Context, privateKeyHex string) (context.Context, error) {
	if privateKeyHex == "" {
		return ctx, nil
	}
	data, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return ctx, err
	}
	privKey := secp256k1.PrivKeyFromBytes(data)
	newIdentity, err := identity.FromPrivateKey(crypto.NewPrivateKey(privKey))
	if err != nil {
		return ctx, err
	}
	immutableIdentity := immutable.Some[identity.Identity](newIdentity)
	newctx := identity.WithContext(ctx, immutableIdentity)
	return newctx, nil
}

// contextWithTransaction is a helper function that attaches transaction to a context
func contextWithTransaction(n int, ctx context.Context, TxnIDu64 uint64) (context.Context, error) {
	if TxnIDu64 == 0 {
		return ctx, nil
	}
	tx, ok := TxnStoreMap[n].Load(TxnIDu64)
	if !ok {
		return ctx, fmt.Errorf(errTxnDoesNotExist, TxnIDu64)
	}
	txn := tx.(datastore.Txn) //nolint:forcetypeassert
	ctx2 := datastore.CtxSetTxn(ctx, txn)
	return ctx2, nil
}

// splitCommaSeparatedString is a helper function that turns a single string into an array
func splitCommaSeparatedString(baseStr string) []string {
	var retArr []string
	if baseStr != "" {
		retArr = strings.Split(baseStr, ",")
	} else {
		retArr = []string{}
	}
	return retArr
}

// buildRequestOptions is a helper function that builds the RequestOption from an operation name,
// and a set of variables (as strings)
func buildRequestOptions(opName string, vars string) ([]client.RequestOption, error) {
	var opts []client.RequestOption
	if opName != "" {
		opts = append(opts, client.WithOperationName(opName))
	}
	if vars != "" {
		var variables map[string]any
		if err := json.Unmarshal([]byte(vars), &variables); err != nil {
			return nil, fmt.Errorf("invalid JSON in variables: %w", err)
		}
		opts = append(opts, client.WithVariables(variables))
	}
	return opts, nil
}

// identityFromKey is a helper function that takes a key type/private key pair, and returns Identity
func identityFromKey(goKeyType string, goPrivKeyStr string) (*identity.FullIdentity, error) {
	if goKeyType == "" || goPrivKeyStr == "" {
		return nil, nil
	}

	var keyType crypto.KeyType
	switch goKeyType {
	case KeyTypeEd25519:
		keyType = crypto.KeyTypeEd25519
	case KeyTypeSecp256k1:
		keyType = crypto.KeyTypeSecp256k1
	default:
		return nil, fmt.Errorf("invalid key type: %s", goKeyType)
	}

	privKey, err := crypto.PrivateKeyFromString(keyType, goPrivKeyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to construct private key: %w", err)
	}

	id, err := identity.FromPrivateKey(privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity from private key: %w", err)
	}

	return &id, nil
}

// GetNode is a thread-safe getter for a global node
func GetNode(n int) *node.Node {
	globalNodesMu.RLock()
	defer globalNodesMu.RUnlock()
	return GlobalNodes[n]
}

// SetNode is a thread-safe setter for a global node
func SetNode(n int, node *node.Node) {
	globalNodesMu.Lock()
	defer globalNodesMu.Unlock()
	GlobalNodes[n] = node
}
