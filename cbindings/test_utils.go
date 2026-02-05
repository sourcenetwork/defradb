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

/*
#include <stdlib.h>
#include "defra_structs.h"
*/
import "C"

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/cgo"
	"strings"
	"unsafe"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/utils"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

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

// identityFromOptions creates a cgo handle, wrapped as a pointer, from options that implement OptionWithIdentity.
// If the options slice is empty or nil, returns 0.
func identityFromOptions[T any, PT interface {
	*T
	utils.OptionWithIdentity[PT]
}](opts []options.Lister[T]) C.uintptr_t {
	if len(opts) == 0 {
		return C.uintptr_t(0)
	}
	optFuncs := opts[0].List()
	if len(optFuncs) == 0 {
		return C.uintptr_t(0)
	}
	var opt T
	for _, f := range optFuncs {
		f(&opt)
	}
	idf := PT(&opt).GetIdentity()
	if !idf.HasValue() {
		return C.uintptr_t(0)
	}
	val := idf.Value()
	handle := cgo.NewHandle(val)
	return C.uintptr_t(handle)
}

// isEncryptedFromCollectionCreateOptions is a helper function that extracts as a C.int
func isEncryptedFromCollectionCreateOptions(opts []options.Lister[options.CollectionCreateOptions]) C.int {
	if len(opts) == 0 {
		return 0
	}
	optFuncs := opts[0].List()
	if len(optFuncs) == 0 {
		return 0
	}
	var opt options.CollectionCreateOptions
	for _, f := range optFuncs {
		f(&opt)
	}
	if opt.EncryptDoc {
		return 1
	}
	return 0
}

// encryptedFieldsFromCollectionCreateOptions is a helper function that returns a comma separated
// C-string, or a blank string, representing the fields that should be encrypted
// After calling this, the caller is responsible for freeing the string returned
func encryptedFieldsFromCollectionCreateOptions(opts []options.Lister[options.CollectionCreateOptions]) *C.char {
	if len(opts) == 0 {
		return C.CString("")
	}
	optFuncs := opts[0].List()
	if len(optFuncs) == 0 {
		return C.CString("")
	}
	var opt options.CollectionCreateOptions
	for _, f := range optFuncs {
		f(&opt)
	}
	if len(opt.EncryptedFields) > 0 {
		return C.CString(strings.Join(opt.EncryptedFields, ","))
	}
	return C.CString("")
}

// extractStringsFromRequestOptions is a helper function that extracts operation name and variables
// as strings from the request option object. They will be blank strings if not present.
func extractStringsFromRequestOptions(opts []options.Lister[options.ExecRequestOptions]) (string, string, error) {
	if len(opts) == 0 {
		return "", "", nil
	}
	optFuncs := opts[0].List()
	if len(optFuncs) == 0 {
		return "", "", nil
	}
	var opt options.ExecRequestOptions
	for _, f := range optFuncs {
		f(&opt)
	}

	opName := ""
	if opt.OperationName.HasValue() {
		opName = opt.OperationName.Value()
	}

	varsJSON := ""
	if opt.Variables != nil {
		data, err := json.Marshal(opt.Variables)
		if err != nil {
			return "", "", err
		}
		varsJSON = string(data)
	}
	return opName, varsJSON, nil
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

// stringFromImmutableOptionString is a helper function to extract a simple string
func stringFromImmutableOptionString(s immutable.Option[string]) string {
	if !s.HasValue() {
		return ""
	}
	return s.Value()
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

// convertGoCResultToGQLResult is a helper function that make a GQLResult from a GoCResult
func convertGoCResultToGQLResult(res GoCResult) (client.GQLResult, error) {
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
				cSubID := C.CString(subID)
				res := ConvertAndFreeCResult(PollSubscription(cSubID))
				C.free(unsafe.Pointer(cSubID))
				if res.Value == "" {
					continue
				}
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

func getNodeOrTxnHandle(h cgo.Handle, ctx context.Context) C.uintptr_t {
	if txn, ok := datastore.CtxTryGetTxn(ctx); ok {
		if h, ok := txnHandleMap.Load(txn); ok {
			return C.uintptr_t(h.(cgo.Handle)) //nolint:forcetypeassert
		}
	}
	return C.uintptr_t(h)
}
