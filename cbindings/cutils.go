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

package cbindings

// The following comment is to allow use of C structs in the Go code

/*
#include <stdlib.h>
#include "defra_structs.h"
*/
import "C"
import (
	"context"
	"fmt"
	"runtime/cgo"
	"unsafe"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/node"

	"github.com/sourcenetwork/immutable"
)

// Helper function which builds a return struct from Go to C
func returnC(gcr GoCResult) *C.Result {
	result := (*C.Result)(C.malloc(C.size_t(unsafe.Sizeof(C.Result{}))))

	result.status = C.int(gcr.Status)
	result.error = C.CString(gcr.Error)
	result.value = C.CString(gcr.Value)

	return result
}

func returnNewNodeResultC(status int, error string, n *node.Node) C.NewNodeResult {
	result := C.NewNodeResult{}
	result.status = C.int(status)
	if error != "" {
		result.error = C.CString(error)
	} else {
		result.error = nil
	}

	if n != nil {
		result.nodePtr = C.uintptr_t(cgo.NewHandle(n))
	} else {
		result.nodePtr = C.uintptr_t(0)
	}

	return result
}

func returnNewTxnResultC(status int, error string, n client.Txn) C.NewTxnResult {
	result := C.NewTxnResult{}
	result.status = C.int(status)
	if error != "" {
		result.error = C.CString(error)
	} else {
		result.error = nil
	}
	if n != nil {
		result.txnPtr = C.uintptr_t(cgo.NewHandle(n))
	} else {
		result.txnPtr = C.uintptr_t(0)
	}
	return result
}

func returnNewIdentityResultC(status int, error string, n identity.Identity) C.NewIdentityResult {
	result := C.NewIdentityResult{}
	result.status = C.int(status)
	if error != "" {
		result.error = C.CString(error)
	} else {
		result.error = nil
	}
	if n != nil {
		result.identityPtr = C.uintptr_t(cgo.NewHandle(n))
	} else {
		result.identityPtr = C.uintptr_t(0)
	}
	return result
}

func convertNodeInitOptionsToGoNodeInitOptions(cOptions C.NodeInitOptions) (GoNodeInitOptions, error) {
	ident, err := getIdentityFromPointer(cOptions.identityPtr)
	if err != nil {
		return GoNodeInitOptions{}, err
	}
	return GoNodeInitOptions{
		DbPath:                   C.GoString(cOptions.dbPath),
		ListeningAddresses:       C.GoString(cOptions.listeningAddresses),
		ReplicatorRetryIntervals: C.GoString(cOptions.replicatorRetryIntervals),
		Peers:                    C.GoString(cOptions.peers),
		Identity:                 ident,
		InMemory:                 int(cOptions.inMemory),
		DisableP2P:               int(cOptions.disableP2P),
		DisableAPI:               int(cOptions.disableAPI),
		MaxTransactionRetries:    int(cOptions.maxTransactionRetries),
		EnableNodeACP:            int(cOptions.enableNodeACP),
	}, nil
}

// recoverHandleValue is a helper function that recovers a handle's value from a pointer,
// and recovers from a panic if the handle is invalid
func recoverHandleValue(ptr C.uintptr_t) (v any, err error) {
	defer func() {
		if r := recover(); r != nil {
			v, err = nil, fmt.Errorf(errInvalidCGOHandle, uintptr(ptr))
		}
	}()
	h := cgo.Handle(ptr)
	return h.Value(), nil
}

// getStoreFromPointer should be used by functions that can work on a node pointer or
// on a transaction pointer.
func getStoreFromPointer(nodePtr C.uintptr_t) (store client.Store, err error) {
	v, err := recoverHandleValue(nodePtr)
	if err != nil {
		return nil, err
	}
	switch v := v.(type) {
	case *node.Node:
		return v.DB, nil
	case client.Txn:
		return v, nil
	default:
		return nil, fmt.Errorf(errInvalidCGOHandle, uintptr(nodePtr))
	}
}

// getNodeFromPointer should be used by functions that can only work on a node pointer.
func getNodeFromPointer(nodePtr C.uintptr_t) (n *node.Node, err error) {
	v, err := recoverHandleValue(nodePtr)
	if err != nil {
		return nil, err
	}
	n, ok := v.(*node.Node)
	if !ok || n == nil {
		return nil, fmt.Errorf(errInvalidCGOHandle, uintptr(nodePtr))
	}
	return n, nil
}

func getIdentityFromPointer(identityPtr C.uintptr_t) (ident identity.Identity, err error) {
	if identityPtr == 0 {
		return nil, nil
	}
	v, err := recoverHandleValue(identityPtr)
	if err != nil {
		return nil, err
	}
	switch v := v.(type) {
	case identity.Identity:
		return v, nil
	default:
		return nil, fmt.Errorf(errInvalidCGOHandle, uintptr(identityPtr))
	}
}

// contextWithIdentity is a helper function that attaches identity to a context
func contextWithIdentity(ctx context.Context, identityPtr C.uintptr_t) (context.Context, error) {
	ident, err := getIdentityFromPointer(identityPtr)
	if err != nil {
		return ctx, err
	}
	if ident == nil {
		return ctx, nil
	}
	return identity.WithContext(ctx, immutable.Some[identity.Identity](ident)), nil
}

// ConvertAndFreeCResult exists to convert C.Result to GoCResult for use in integration tests
// It will, in converting,consume the C.Result, freeing the memory for it
func ConvertAndFreeCResult(cResult *C.Result) GoCResult {
	defer C.free(unsafe.Pointer(cResult.error))
	defer C.free(unsafe.Pointer(cResult.value))
	return GoCResult{
		Status: int(cResult.status),
		Error:  C.GoString(cResult.error),
		Value:  C.GoString(cResult.value),
	}
}
