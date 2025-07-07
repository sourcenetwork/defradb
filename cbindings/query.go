// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build cgo
// +build cgo

package main

/*
#include "defra_structs.h"
*/
import "C"

import (
	"context"
	"strings"
)

//export executeQuery
func executeQuery(
	cQuery *C.char,
	cIdentity *C.char,
	cTxnID C.ulonglong,
	cOperationName *C.char,
	cVariables *C.char,
) *C.Result {
	query := C.GoString(cQuery)
	identityStr := C.GoString(cIdentity)
	ctx := context.Background()
	opts, err := buildRequestOptions(cOperationName, cVariables)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Attach the identity
	newctx, err := contextWithIdentity(ctx, identityStr)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Set the transaction
	newctx, err = contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	res := globalNode.DB.ExecRequest(ctx, query, opts...)

	// Check for errors in the GQL response, wrangling them into a single string
	if len(res.GQL.Errors) > 0 {
		var sb strings.Builder
		sb.WriteString("Error executing query:\n")
		for _, err := range res.GQL.Errors {
			sb.WriteString("Error: ")
			sb.WriteString(err.Error())
			sb.WriteString("\n")
		}
		return returnC(1, sb.String(), "")
	}

	// Try to marshall the JSON and return it
	dataMap, ok := res.GQL.Data.(map[string]any)
	if !ok || dataMap == nil || len(dataMap) == 0 {
		return returnC(1, "GraphQL response data is nil or empty.", "")
	}
	return marshalJSONToCResult(dataMap)
}

func main() {}
