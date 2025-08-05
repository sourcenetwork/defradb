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

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/immutable/enumerable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// setupTests is a function that initializes and starts the globalNode (in memory), for the tests
func setupTests(n int, identityString string, enableNAC bool) {
	var nodeOpts cbindings.GoNodeInitOptions
	nodeOpts.DbPath = ""
	nodeOpts.ListeningAddresses = ""
	nodeOpts.ReplicatorRetryIntervals = ""
	nodeOpts.Peers = ""
	nodeOpts.MaxTransactionRetries = 5
	nodeOpts.DisableP2P = 0
	nodeOpts.DisableAPI = 0
	nodeOpts.InMemory = 1
	nodeOpts.IdentityPrivateKey = identityString
	if identityString != "" {
		if enableNAC {
			nodeOpts.EnableNodeACP = 1
		}
		nodeOpts.IdentityKeyType = "secp256k1"
	}

	cbindings.NodeInit(n, nodeOpts)
	//cbindings.NodeStart(n, identityString)
}

// txnIDFromContext is a helper function that extracts a transaction ID from a context
func txnIDFromContext(ctx context.Context) uint64 {
	tx, ok := datastore.CtxTryGetTxn(ctx)
	if ok {
		return tx.ID()
	}
	return 0
}

// isEncryptedFromDocCreateOption is a helper function that extracts a boolean
func isEncryptedFromDocCreateOption(opts []client.DocCreateOption) bool {
	createDocOpts := client.DocCreateOptions{}
	createDocOpts.Apply(opts)
	return createDocOpts.EncryptDoc
}

// encryptedFieldsFromDocCreateOptions is a helper function that returns a comma separated string,
// or a blank string, representing the fields that should be encrypted
func encryptedFieldsFromDocCreateOptions(opts []client.DocCreateOption) string {
	createDocOpts := client.DocCreateOptions{}
	createDocOpts.Apply(opts)
	if len(createDocOpts.EncryptedFields) > 0 {
		return strings.Join(createDocOpts.EncryptedFields, ",")
	}
	return ""
}

// identityFromContext is a helper function that extracts identity (or blank string) from a context
func identityFromContext(ctx context.Context) string {
	idf := identity.FullFromContext(ctx)
	if !idf.HasValue() {
		return ""
	}
	return idf.Value().PrivateKey().String()
}

// unmarshalResult is a helper function that unmarshals JSON string into another type
func unmarshalResult[T any](value string) (T, error) {
	var result T
	err := json.Unmarshal([]byte(value), &result)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to unmarshal JSON into %T: %w", result, err)
	}
	return result, nil
}

// optionToString is a helper function that extracts a string from an immutable.Option
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

// extractStringsFromRequestOptions is a helper function that extracts operation name and variables
// as strings from the request option object. They will be blank strings if not present.
func extractStringsFromRequestOptions(opts []client.RequestOption) (string, string, error) {
	config := &client.GQLOptions{}
	for _, opt := range opts {
		opt(config)
	}

	opName := ""
	if config.OperationName != "" {
		opName = config.OperationName
	}

	varsJSON := ""
	if config.Variables != nil {
		data, err := json.Marshal(config.Variables)
		if err != nil {
			return "", "", err
		}
		varsJSON = string(data)
	}
	return opName, varsJSON, nil
}

// stringFromImmutableOptionString is a helper function to extract a simple string
func stringFromImmutableOptionString(s immutable.Option[string]) string {
	if !s.HasValue() {
		return ""
	}
	return s.Value()
}

// stringFromLensOption is a helper function to extract a simple string
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

// collectEnumerable is a helper function for wrangling data from an Enumerable:
// enumerable.Enumerable[map[string]any] -> []map[string]any
func collectEnumerable(e enumerable.Enumerable[map[string]any]) ([]map[string]any, error) {
	var result []map[string]any
	err := enumerable.ForEach(e, func(item map[string]any) {
		result = append(result, item)
	})
	return result, err
}

// getTxnFromHandle is a helper function that gets a transaction from the C-side TxnStore
func getTxnFromHandle(n int, txnID uint64) any {
	val, ok := cbindings.TxnStoreMap[n].Load(txnID)
	if !ok {
		return 0
	}
	return val
}

// convertGoCResultToGQLResult is a helper function that make a GQLResult from a GoCResult
func convertGoCResultToGQLResult(res cbindings.GoCResult) (client.GQLResult, error) {
	var gql client.GQLResult
	if res.Status != 0 {
		return gql, errors.New(res.Value)
	}
	err := json.Unmarshal([]byte(res.Value), &gql)
	return gql, err
}

// wrapSubscriptionAsChannel is a function that takes a subscription ID and returns a GQLResult
// channel that is populated by polling the subscription in a loop. It takes in a context as
// well, so that it will terminate when the context is done
func wrapSubscriptionAsChannel(ctx context.Context, subID string) <-chan client.GQLResult {
	ch := make(chan client.GQLResult)
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				res := cbindings.PollSubscription(subID)
				goRes, err := convertGoCResultToGQLResult(res)
				if err != nil {
					goRes.Errors = append(goRes.Errors, err)
				}
				select {
				case ch <- goRes:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return ch
}
