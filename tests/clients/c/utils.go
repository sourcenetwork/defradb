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

/*
#include <stdlib.h>
#include "defra_structs.h"
*/
import "C"

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/lens-vm/lens/host-go/config/model"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// Helper function which builds a return struct from Go to C
func returnC(status int, errortext string, valuetext string) *C.Result {
	result := (*C.Result)(C.malloc(C.size_t(unsafe.Sizeof(C.Result{}))))

	result.status = C.int(status)
	result.error = C.CString(errortext)
	result.value = C.CString(valuetext)

	return result
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
func contextWithTransaction(ctx context.Context, cTxnID C.ulonglong) (context.Context, error) {
	TxnIDu64 := uint64(cTxnID)
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
func marshalJSONToCResult(value any) *C.Result {
	dataJSON, err := json.Marshal(value)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrMarshallingJSON, err), "")
	}
	return returnC(0, "", string(dataJSON))
}

// Helper function that takes a comma separated const char * and returns an array of Go strings
func splitCommaSeparatedCString(cStr *C.char) []string {
	baseStr := C.GoString(cStr)
	var retArr []string
	if baseStr != "" {
		retArr = strings.Split(baseStr, ",")
	} else {
		retArr = []string{}
	}
	return retArr
}

// Helper function that tries to build a []client.RequestOption from C-string name and JSON variables
func buildRequestOptions(cOpName *C.char, cVars *C.char) ([]client.RequestOption, error) {
	var opts []client.RequestOption
	if cOpName != nil && C.GoString(cOpName) != "" {
		opts = append(opts, client.WithOperationName(C.GoString(cOpName)))
	}
	if cVars != nil && C.GoString(cVars) != "" {
		var variables map[string]any
		if err := json.Unmarshal([]byte(C.GoString(cVars)), &variables); err != nil {
			return nil, fmt.Errorf("invalid JSON in variables: %w", err)
		}
		opts = append(opts, client.WithVariables(variables))
	}
	return opts, nil
}

func loadIdentityFromString(cKeyType *C.char, cPrivKey *C.char) (*identity.FullIdentity, error) {
	goKeyType := C.GoString(cKeyType)
	goPrivKeyStr := C.GoString(cPrivKey)

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

// Helper function
// This initializes and starts the globalNode (in memory), so that other functionality works
func setupTests() {
	dbPath := C.CString("")
	listeningAddresses := C.CString("")
	replicatorRetryIntervals := C.CString("")
	peers := C.CString("")
	keyType := C.CString("")    // secp256k1
	privateKey := C.CString("") // 545cd8289a64f0442224cff3ab8cc459a18ec7fbb9f6a58ae4c64cc8b4d59101
	defer C.free(unsafe.Pointer(keyType))
	defer C.free(unsafe.Pointer(privateKey))
	defer C.free(unsafe.Pointer(dbPath))
	defer C.free(unsafe.Pointer(listeningAddresses))
	defer C.free(unsafe.Pointer(replicatorRetryIntervals))
	defer C.free(unsafe.Pointer(peers))

	var nodeOpts C.NodeInitOptions
	nodeOpts.dbPath = dbPath
	nodeOpts.listeningAddresses = listeningAddresses
	nodeOpts.replicatorRetryIntervals = replicatorRetryIntervals
	nodeOpts.peers = peers
	nodeOpts.maxTransactionRetries = 5
	nodeOpts.disableP2P = 0
	nodeOpts.disableAPI = 0
	nodeOpts.inMemory = 1
	nodeOpts.identityKeyType = keyType
	nodeOpts.identityPrivateKey = privateKey

	NodeInit(nodeOpts)
	NodeStart()
}

// Helper function
// Get TxnID, as a C.ulonglong, from a context, returning 0 if not present
func cTxnIDFromContext(ctx context.Context) C.ulonglong {
	var cTxnID C.ulonglong = 0

	tx, ok := datastore.CtxTryGetTxn(ctx)
	if ok {
		cTxnID = C.ulonglong(tx.ID())
	}
	return cTxnID
}

// Helper function
// Returns a C integer, representing a boolean value, for whether EncryptDoc flag is set
func cIsEncryptedFromDocCreateOption(opts []client.DocCreateOption) C.int {
	createDocOpts := client.DocCreateOptions{}
	createDocOpts.Apply(opts)
	var retVal C.int = 0
	if createDocOpts.EncryptDoc {
		retVal = 1
	}
	return retVal
}

// Helper function
// Get EncryptedFields as a comma separated C-String, returning "" if none exist
// After calling this, you are responsible for freeing the memory
func cEncryptedFieldsFromDocCreateOptions(opts []client.DocCreateOption) *C.char {
	createDocOpts := client.DocCreateOptions{}
	createDocOpts.Apply(opts)
	if len(createDocOpts.EncryptedFields) > 0 {
		joined := strings.Join(createDocOpts.EncryptedFields, ",")
		return C.CString(joined)
	}
	return C.CString("")
}

// Helper function
// Get Identity, as a *C.char, from a context, returning "" if not present
// After calling this, you are responsible for freeing the memory
func cIdentityFromContext(ctx context.Context) *C.char {
	idf := identity.FullFromContext(ctx)
	if !idf.HasValue() {
		return C.CString("")
	}
	return C.CString(idf.Value().PrivateKey().String())
}

// Helper function
// Unpacks a C Result into either a payload/error pair, freeing memory afterwards
func (w *CWrapper) callResult(r C.Result) (json.RawMessage, error) {
	defer C.free(unsafe.Pointer(r.value))
	defer C.free(unsafe.Pointer(r.error))

	if r.status != 0 {
		msg := C.GoString(r.error)
		return nil, errors.New(msg)
	}

	data := C.GoString(r.value)
	return json.RawMessage(data), nil
}

// Helper function
// Unmarshals the value of a JSON C-string into any desired type
func unmarshalResult[T any](value *C.char) (T, error) {
	var result T
	payload := C.GoString(value)
	err := json.Unmarshal([]byte(payload), &result)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to unmarshal JSON into %T: %w", result, err)
	}
	return result, nil
}

// Helper function
// Marshals an Option[T] to a C-String.
// After calling this, you are responsible for freeing the memory
func optionToCString[T any](opt immutable.Option[T]) (*C.char, error) {
	if opt.HasValue() {
		return nil, nil
	}
	value := opt.Value()
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return C.CString(string(jsonBytes)), nil
}

// Helper function
// Pulls out the Operation Name and Variables as strings from a []client.RequestOption.
// The strings may or may not be blank.
// After calling this, you are responsible for freeing the memory
func extractCStringsFromRequestOptions(opts []client.RequestOption) (opNameC, varsC *C.char) {
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
	opNameC = C.CString(opName)

	// Extract Variables (marshal to JSON), leaving the JSON blank if no variables exist
	varsJSON := ""
	if config.Variables != nil {
		data, _ := json.Marshal(config.Variables)
		varsJSON = string(data)
	}
	varsC = C.CString(varsJSON)
	return opNameC, varsC
}

// Helper function
// Creates a C-string from a client.CollectionDefinition
// After calling this, you are responsible for freeing the memory
func cStringFromCollectionDefinition(def client.CollectionDefinition) *C.char {
	jsonBytes, _ := json.Marshal(def)
	return C.CString(string(jsonBytes))
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
// Creates a C-string from an immutable.Option[string]
// After calling this, you are responsible for freeing the memory
func cStringFromImmutableOptionString(s immutable.Option[string]) *C.char {
	if !s.HasValue() {
		return C.CString("")
	}
	return C.CString(s.Value())
}

// Helper function
// Creates a C-string from an immutable.Option[model.Lens]
// After calling this, you are responsible for freeing the memory
func cStringFromLensOption(opt immutable.Option[model.Lens]) (*C.char, error) {
	if !opt.HasValue() {
		return C.CString(""), nil
	}
	lens := opt.Value()
	data, err := json.Marshal(lens)
	if err != nil {
		return C.CString(""), err
	}
	return C.CString(string(data)), nil
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
// Frees a *C.Result and the C-Strings it contains
func freeCResult(result *C.Result) {
	if result != nil {
		if result.value != nil {
			C.free(unsafe.Pointer(result.value))
		}
		if result.error != nil {
			C.free(unsafe.Pointer(result.error))
		}
		C.free(unsafe.Pointer(result))
	}
}

// Helper function
// Gets a client.Txn from a C.ulonglong representing it in the C-side TxnStore
// This function is only necessary to allow for the test wrapper to function
func GetTxnFromHandle(cTxnID C.ulonglong) any {
	TxnIDu64 := uint64(cTxnID)
	val, ok := TxnStore.Load(TxnIDu64)
	if !ok {
		return 0
	}
	return val
}

func convertCResultToGQLResult(res *C.Result) (client.GQLResult, error) {
	var gql client.GQLResult
	if res.status != 0 {
		return gql, fmt.Errorf(C.GoString(res.value))
	}
	err := json.Unmarshal([]byte(C.GoString(res.value)), &gql)
	return gql, err
}

func WrapSubscriptionAsChannel(subID string) <-chan client.GQLResult {
	ch := make(chan client.GQLResult)
	go func() {
		defer close(ch)
		cID := C.CString(subID)
		defer C.free(unsafe.Pointer(cID))
		for {
			res := PollSubscription(cID)
			if res == nil {
				return
			}
			goRes, err := convertCResultToGQLResult(res)
			freeCResult(res)
			if err != nil {
				goRes.Errors = append(goRes.Errors, err)
			}

			ch <- goRes
		}
	}()
	return ch
}
