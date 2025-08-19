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
	EnableNodeACP            int
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
		return identity.WithContext(ctx, immutable.None[identity.Identity]()), nil
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
	return identity.WithContext(ctx, immutableIdentity), nil
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
func identityFromKey(goKeyType string, goPrivKeyStr string) (identity.FullIdentity, error) {
	if goKeyType == "" || goPrivKeyStr == "" {
		return nil, nil
	}

	var keyType crypto.KeyType
	switch goKeyType {
	case string(crypto.KeyTypeEd25519):
		keyType = crypto.KeyTypeEd25519
	case string(crypto.KeyTypeSecp256k1):
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

	return id, nil
}
