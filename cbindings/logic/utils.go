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

// Helper function which builds a return struct from Go to C
func returnGoC(status int, errortext string, valuetext string) GoCResult {
	return GoCResult{
		Status: status,
		Error:  errortext,
		Value:  valuetext,
	}
}

// Helper function that attaches an identity to a context, returning the new context
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

// Helper function that attaches a transaction to a context, returning a new context
func contextWithTransaction(ctx context.Context, TxnIDu64 uint64) (context.Context, error) {
	if TxnIDu64 == 0 {
		return ctx, nil
	}
	tx, ok := TxnStore.Load(TxnIDu64)
	if !ok {
		return ctx, fmt.Errorf(cerrTxnDoesNotExist, TxnIDu64)
	}
	txn := tx.(datastore.Txn) //nolint:forcetypeassert
	ctx2 := datastore.CtxSetTxn(ctx, txn)
	return ctx2, nil
}

// Helper function that seeks to marshall JSON into a CResult
// The Result object will either contain the payload, if it works, or an error if it doesn't
func marshalJSONToGoCResult(value any) GoCResult {
	dataJSON, err := json.Marshal(value)
	if err != nil {
		return returnGoC(1, fmt.Sprintf(cerrMarshallingJSON, err), "")
	}
	return returnGoC(0, "", string(dataJSON))
}

func splitCommaSeparatedString(baseStr string) []string {
	var retArr []string
	if baseStr != "" {
		retArr = strings.Split(baseStr, ",")
	} else {
		retArr = []string{}
	}
	return retArr
}

// Helper function that tries to build a []client.RequestOption from string name and JSON variables
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

func loadIdentityFromString(goKeyType string, goPrivKeyStr string) (*identity.FullIdentity, error) {
	if goKeyType == "" || goPrivKeyStr == "" {
		return nil, nil
	}

	// Convert string key type to crypto.KeyType
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
