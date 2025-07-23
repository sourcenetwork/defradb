// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cwrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	cbindings "github.com/sourcenetwork/defradb/cbindings/logic"

	"github.com/lens-vm/lens/host-go/config/model"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// Helper function
// This initializes and starts the globalNode (in memory), so that other functionality works
func setupTests() {
	dbPath := ""
	listeningAddresses := ""
	replicatorRetryIntervals := ""
	peers := ""
	keyType := ""
	privateKey := ""

	var nodeOpts cbindings.GoNodeInitOptions
	nodeOpts.DbPath = dbPath
	nodeOpts.ListeningAddresses = listeningAddresses
	nodeOpts.ReplicatorRetryIntervals = replicatorRetryIntervals
	nodeOpts.Peers = peers
	nodeOpts.MaxTransactionRetries = 5
	nodeOpts.DisableP2P = 0
	nodeOpts.DisableAPI = 0
	nodeOpts.InMemory = 1
	nodeOpts.IdentityKeyType = keyType
	nodeOpts.IdentityPrivateKey = privateKey

	cbindings.NodeInit(nodeOpts)
	cbindings.NodeStart()
}

// Helper function
// Get TxnID, as a uint64, from a context, returning 0 if not present
func txnIDFromContext(ctx context.Context) uint64 {
	tx, ok := datastore.CtxTryGetTxn(ctx)
	if ok {
		return tx.ID()
	}
	return 0
}

// Helper function
// Returns a boolean value, for whether EncryptDoc flag is set
func isEncryptedFromDocCreateOption(opts []client.DocCreateOption) bool {
	createDocOpts := client.DocCreateOptions{}
	createDocOpts.Apply(opts)
	return createDocOpts.EncryptDoc
}

// Helper function
// Get EncryptedFields as a comma separated string, returning "" if none exist
func encryptedFieldsFromDocCreateOptions(opts []client.DocCreateOption) string {
	createDocOpts := client.DocCreateOptions{}
	createDocOpts.Apply(opts)
	if len(createDocOpts.EncryptedFields) > 0 {
		return strings.Join(createDocOpts.EncryptedFields, ",")
	}
	return ""
}

// Helper function
// Get a private key, or blank string, used to pass in identity, as a string
func identityFromContext(ctx context.Context) string {
	idf := identity.FullFromContext(ctx)
	if !idf.HasValue() {
		return ""
	}
	return idf.Value().PrivateKey().String()
}

// Helper function
// Unmarshals the value of a string into any desired type
func unmarshalResult[T any](value string) (T, error) {
	var result T
	err := json.Unmarshal([]byte(value), &result)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to unmarshal JSON into %T: %w", result, err)
	}
	return result, nil
}

// Helper function
// Marshals an Option[T] to a string.
func optionToString[T any](opt immutable.Option[T]) (string, error) {
	if !opt.HasValue() {
		return "", nil
	}
	value := opt.Value()
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// Helper function
// Pulls out the Operation Name and Variables as strings from a []client.RequestOption.
// The strings may or may not be blank.
func extractCStringsFromRequestOptions(opts []client.RequestOption) (string, string) {
	// Create a structure, and modify it with the config options
	config := &client.GQLOptions{}
	for _, opt := range opts {
		opt(config)
	}

	// Extract OperationName, leaving it blank if one does not exist
	opName := ""
	if config.OperationName != "" {
		opName = config.OperationName
	}

	// Extract Variables (marshal to JSON), leaving the JSON blank if no variables exist
	varsJSON := ""
	if config.Variables != nil {
		data, _ := json.Marshal(config.Variables)
		varsJSON = string(data)
	}
	return opName, varsJSON
}

// Helper function
// Creates a string from a client.CollectionDefinition
func stringFromCollectionDefinition(def client.CollectionDefinition) string {
	jsonBytes, _ := json.Marshal(def)
	return string(jsonBytes)
}

// Helper function
// Builds a collection from a definition
func collectionsFromDefinitions(defs []client.CollectionDefinition) ([]client.Collection, error) {
	cols := make([]client.Collection, len(defs))
	for i, def := range defs {
		cols[i] = &Collection{def: def}
	}
	return cols, nil
}

// Helper function
// Creates a string from an immutable.Option[string]
func stringFromImmutableOptionString(s immutable.Option[string]) string {
	if !s.HasValue() {
		return ""
	}
	return s.Value()
}

// Helper function
// Creates a string from an immutable.Option[model.Lens]
func stringFromLensOption(opt immutable.Option[model.Lens]) (string, error) {
	if !opt.HasValue() {
		return "", nil
	}
	lens := opt.Value()
	data, err := json.Marshal(lens)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Helper function
// Wrangle data from enumerable.Enumerable[map[string]any] into []map[string]any
func collectEnumerable(e enumerable.Enumerable[map[string]any]) ([]map[string]any, error) {
	var result []map[string]any
	err := enumerable.ForEach(e, func(item map[string]any) {
		result = append(result, item)
	})
	return result, err
}

// Helper function
// Gets a client.Txn from a uint64 representing it in the C-side TxnStore
func getTxnFromHandle(txnID uint64) any {
	val, ok := cbindings.TxnStore.Load(txnID)
	if !ok {
		return 0
	}
	return val
}

func convertGoCResultToGQLResult(res cbindings.GoCResult) (client.GQLResult, error) {
	var gql client.GQLResult
	if res.Status != 0 {
		return gql, errors.New(res.Value)
	}
	err := json.Unmarshal([]byte(res.Value), &gql)
	return gql, err
}

func WrapSubscriptionAsChannel(subID string) <-chan client.GQLResult {
	ch := make(chan client.GQLResult)
	go func() {
		defer close(ch)
		for {
			res := cbindings.PollSubscription(subID)
			goRes, err := convertGoCResultToGQLResult(res)
			if err != nil {
				goRes.Errors = append(goRes.Errors, err)
			}
			ch <- goRes
		}
	}()
	return ch
}
