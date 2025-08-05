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
	"unsafe"

	cbindings "github.com/sourcenetwork/defradb/cbindings/logic"
)

// Helper function which builds a return struct from Go to C
func returnC(gcr cbindings.GoCResult) *C.Result {
	result := (*C.Result)(C.malloc(C.size_t(unsafe.Sizeof(C.Result{}))))

	result.status = C.int(gcr.Status)
	result.error = C.CString(gcr.Error)
	result.value = C.CString(gcr.Value)

	return result
}

func convertCOptionsToGoCOptions(cOptions C.CollectionOptions) cbindings.GoCOptions {
	return cbindings.GoCOptions{
		TxID:         uint64(cOptions.tx),
		Version:      C.GoString(cOptions.version),
		CollectionID: C.GoString(cOptions.collectionID),
		Name:         C.GoString(cOptions.name),
		Identity:     C.GoString(cOptions.identity),
		GetInactive:  int(cOptions.getInactive),
	}
}

func convertNodeInitOptionsToGoNodeInitOptions(cOptions C.NodeInitOptions) cbindings.GoNodeInitOptions {
	return cbindings.GoNodeInitOptions{
		DbPath:                   C.GoString(cOptions.dbPath),
		ListeningAddresses:       C.GoString(cOptions.listeningAddresses),
		ReplicatorRetryIntervals: C.GoString(cOptions.replicatorRetryIntervals),
		Peers:                    C.GoString(cOptions.peers),
		IdentityKeyType:          C.GoString(cOptions.identityKeyType),
		IdentityPrivateKey:       C.GoString(cOptions.identityPrivateKey),
		InMemory:                 int(cOptions.inMemory),
		DisableP2P:               int(cOptions.disableP2P),
		DisableAPI:               int(cOptions.disableAPI),
		MaxTransactionRetries:    int(cOptions.maxTransactionRetries),
		EnableNodeACP:            int(cOptions.enableodeACP),
	}
}
