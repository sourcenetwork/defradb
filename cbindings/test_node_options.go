// Copyright 2026 Democratized Data Foundation
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
	"encoding/json"
	"testing"
	"unsafe"

	"github.com/sourcenetwork/defradb/node"
)

// Note: This file deliberately avoids the _test.go naming suffix. If it were a _test.go file, `go test`
// would compile it in a separate test compilation unit, conflicting with the //export directives and
// func main() present in this package. The tests here are called via wrappers in node_options_test.go.

// getNestedField traverses a nested map[string]any by following keys in order,
// returning the value at the final key and true, or nil and false if any key is
// missing, or if intermediate value is not itself a map.
func getNestedField(m map[string]any, keys ...string) (any, bool) {
	cur := any(m)
	for _, k := range keys {
		sub, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = sub[k]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}

// assertNodeOptionField asserts that a field in the options map is a particular value.
// It is a helper function for the tests.
func assertNodeOptionField(t *testing.T, m map[string]any, expected any, keys ...string) {
	t.Helper()
	val, ok := getNestedField(m, keys...)
	if !ok {
		t.Errorf("field %v not found in options", keys)
		return
	}
	if val != expected {
		t.Errorf("field %v: want %v (%T), got %v (%T)", keys, expected, expected, val, val)
	}
}

// assertNodeOptionArrayField asserts that a field in the options map is an array of any values
// It is a helper function for the tests.
func assertNodeOptionArrayField(t *testing.T, m map[string]any, expected []any, keys ...string) {
	t.Helper()
	val, ok := getNestedField(m, keys...)
	if !ok {
		t.Errorf("field %v not found in options", keys)
		return
	}
	arr, ok := val.([]any)
	if !ok {
		t.Errorf("field %v: want []any, got %T", keys, val)
		return
	}
	if len(arr) != len(expected) {
		t.Errorf("field %v: want len %d, got len %d; full value: %v", keys, len(expected), len(arr), arr)
		return
	}
	for i := range expected {
		if arr[i] != expected[i] {
			t.Errorf("field %v[%d]: want %v, got %v", keys, i, expected[i], arr[i])
		}
	}
}

// startTestNodeAndGetOptionMap starts a test node and returns the options as a map.
// It is a significant helper function for the tests below.
func startTestNodeAndGetOptionMap(t *testing.T, opts C.NodeInitOptions) map[string]any {
	t.Helper()
	nodeResult := NewNode(opts)
	if nodeResult.status != 0 {
		t.Fatalf("NewNode failed: %s", C.GoString(nodeResult.error))
	}
	t.Cleanup(func() {
		_ = ConvertAndFreeCResult(CloseNode(nodeResult.nodePtr))
	})

	result := ConvertAndFreeCResult(GetNodeOptions(nodeResult.nodePtr))
	if result.Status != 0 {
		t.Fatalf("GetNodeOptions failed: %s", result.Error)
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(result.Value), &m); err != nil {
		t.Fatalf("failed to unmarshal node options: %v", err)
	}
	return m
}

// TestGetNodeOptions_SetA verifies that all node options from set A are reflected by GetNodeOptions.
// It uses specific non-default values so that a separate run (set B) with different values can
// confirm the results are not coincidentally equal to the defaults.
func TestGetNodeOptions_SetA(t *testing.T) {
	// Create an identity so that EnableNodeACP can be set (NAC requires an identity to start).
	identResult := NewIdentity(nil)
	if identResult.status != 0 {
		t.Fatalf("NewIdentity failed: %s", C.GoString(identResult.error))
	}
	defer FreeIdentity(identResult.identityPtr)

	// Allocate C strings
	storeType := C.CString("badger")
	defer C.free(unsafe.Pointer(storeType))
	retryIntervals := C.CString("5,10")
	defer C.free(unsafe.Pointer(retryIntervals))
	listenAddrs := C.CString("/ip4/127.0.0.1/tcp/9171")
	defer C.free(unsafe.Pointer(listenAddrs))
	bootstrapPeers := C.CString("/ip4/127.0.0.2/tcp/9172")
	defer C.free(unsafe.Pointer(bootstrapPeers))
	httpAddr := C.CString("localhost:9191")
	defer C.free(unsafe.Pointer(httpAddr))
	allowedOrigins := C.CString("http://a.example.com")
	defer C.free(unsafe.Pointer(allowedOrigins))
	certPath := C.CString("/tls/cert-a.pem")
	defer C.free(unsafe.Pointer(certPath))
	keyPath := C.CString("/tls/key-a.pem")
	defer C.free(unsafe.Pointer(keyPath))
	dacType := C.CString("local")
	defer C.free(unsafe.Pointer(dacType))
	dacPathStr := t.TempDir()
	dacPath := C.CString(dacPathStr)
	defer C.free(unsafe.Pointer(dacPath))
	logID := C.CString("log-a")
	defer C.free(unsafe.Pointer(logID))
	grpcAddr := C.CString("grpc-a:9090")
	defer C.free(unsafe.Pointer(grpcAddr))
	cometAddr := C.CString("comet-a:26657")
	defer C.free(unsafe.Pointer(cometAddr))
	nodeACPPath := C.CString("")
	defer C.free(unsafe.Pointer(nodeACPPath))

	// Allocate byte-slice fields
	// SearchableEncryptionKey is stored in DB options and redacted in the sanitized output.
	searchKeyA := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	cSearchKeyA := C.CBytes(searchKeyA)
	defer C.free(cSearchKeyA)
	// P2PPrivateKey is stored but never validated when DisableP2P is true.
	p2pKeyA := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	cP2PKeyA := C.CBytes(p2pKeyA)
	defer C.free(cP2PKeyA)

	// Populate NodeInitOptions
	var opts C.NodeInitOptions
	opts.inMemory = 1
	opts.disableP2P = 1
	opts.disableAPI = 1
	// Store
	opts.storeType = storeType
	opts.badgerFileSize = C.int64_t(1048576) // 1 MB
	// DB
	opts.enableSigning = 1
	opts.maxTransactionRetries = C.int(7)
	opts.p2pBlockSyncTimeoutMs = C.int64_t(3000)
	opts.lensPoolSize = C.int(3)
	opts.chunkSize = C.int(512)
	opts.replicatorRetryIntervals = retryIntervals
	opts.searchableEncryptionKey = (*C.uint8_t)(cSearchKeyA)
	opts.searchableEncryptionKeyLen = C.int(len(searchKeyA))
	// P2P
	// The following are stored but not used because P2P is disabled
	opts.enablePubSub = 1
	opts.enableRelay = 1
	opts.enableClearBackoffOnRetry = 1
	opts.listeningAddresses = listenAddrs
	opts.peers = bootstrapPeers
	opts.p2pPrivateKey = (*C.uint8_t)(cP2PKeyA)
	opts.p2pPrivateKeyLen = C.int(len(p2pKeyA))
	// HTTP
	// The following are stored but not used because API is disabled
	opts.httpAddress = httpAddr
	opts.httpAllowedOrigins = allowedOrigins
	opts.tlsCertPath = certPath
	opts.tlsKeyPath = keyPath
	opts.httpReadTimeoutMs = C.int64_t(1000)
	opts.httpWriteTimeoutMs = C.int64_t(2000)
	opts.httpIdleTimeoutMs = C.int64_t(3000)
	// Document ACP
	opts.documentACPType = dacType
	opts.documentACPPath = dacPath
	opts.remoteDACLogID = logID
	opts.remoteDACGRPCAddress = grpcAddr
	opts.remoteDACCometRPCAddress = cometAddr
	// Identity + Node ACP
	opts.identityPtr = identResult.identityPtr
	opts.enableNodeACP = 1
	opts.nodeACPPath = nodeACPPath

	m := startTestNodeAndGetOptionMap(t, opts)

	// Node-level
	assertNodeOptionField(t, m, true, "DisableP2P")
	assertNodeOptionField(t, m, true, "DisableAPI")

	// Store
	assertNodeOptionField(t, m, "badger", "Store", "Store")
	assertNodeOptionField(t, m, true, "Store", "BadgerInMemory")
	assertNodeOptionField(t, m, float64(1048576), "Store", "BadgerFileSize")

	// DB — ChunkSize is always overridden to defaultChunkSize when BadgerInMemory is true.
	assertNodeOptionField(t, m, true, "DB", "EnableSigning")
	assertNodeOptionField(t, m, "<redacted>", "DB", "Identity")
	assertNodeOptionField(t, m, float64(7), "DB", "MaxTxnRetries")
	assertNodeOptionField(t, m, float64(3000000000), "DB", "P2PBlockSyncTimeout") // 3000 ms
	assertNodeOptionField(t, m, float64(3), "DB", "LensPoolSize")
	assertNodeOptionField(t, m, float64(node.GetDefaultChunkSize()), "DB", "ChunkSize")
	assertNodeOptionField(t, m, "wazero", "DB", "LensRuntime")
	assertNodeOptionField(t, m, "<redacted>", "DB", "SearchableEncryptionKey")
	assertNodeOptionArrayField(t, m, []any{float64(5000000000), float64(10000000000)}, "DB", "RetryIntervals")

	// P2P
	assertNodeOptionField(t, m, true, "P2P", "EnablePubSub")
	assertNodeOptionField(t, m, true, "P2P", "EnableRelay")
	assertNodeOptionField(t, m, true, "P2P", "EnableClearBackoffOnRetry")
	assertNodeOptionArrayField(t, m, []any{"/ip4/127.0.0.1/tcp/9171"}, "P2P", "ListenAddresses")
	assertNodeOptionArrayField(t, m, []any{"/ip4/127.0.0.2/tcp/9172"}, "P2P", "BootstrapPeers")
	assertNodeOptionField(t, m, "<redacted>", "P2P", "PrivateKey")

	// HTTP
	assertNodeOptionField(t, m, "localhost:9191", "HTTP", "Address")
	assertNodeOptionArrayField(t, m, []any{"http://a.example.com"}, "HTTP", "AllowedOrigins")
	assertNodeOptionField(t, m, "/tls/cert-a.pem", "HTTP", "TLSCertPath")
	assertNodeOptionField(t, m, "/tls/key-a.pem", "HTTP", "TLSKeyPath")
	assertNodeOptionField(t, m, float64(1000000000), "HTTP", "ReadTimeout")  // 1000 ms
	assertNodeOptionField(t, m, float64(2000000000), "HTTP", "WriteTimeout") // 2000 ms
	assertNodeOptionField(t, m, float64(3000000000), "HTTP", "IdleTimeout")  // 3000 ms

	// Document ACP
	assertNodeOptionField(t, m, "local", "DocumentACP", "DocumentACPType")
	assertNodeOptionField(t, m, dacPathStr, "DocumentACP", "Path")
	assertNodeOptionField(t, m, "log-a", "DocumentACP", "RemoteDACLogID")
	assertNodeOptionField(t, m, "grpc-a:9090", "DocumentACP", "RemoteDACGRPCAddress")
	assertNodeOptionField(t, m, "comet-a:26657", "DocumentACP", "RemoteDACCometRPCAddress")

	// Node ACP
	assertNodeOptionField(t, m, true, "NodeACP", "IsEnabled")
	assertNodeOptionField(t, m, "", "NodeACP", "Path")
}

// TestGetNodeOptions_SetB verifies GetNodeOptions with a distinct set of values, confirming that
// the results from set A were not coincidentally equal to defaults or fixed constants.
// Set B uses disk-based storage (no BadgerInMemory) so that ChunkSize is not overridden by
// node.Start(), giving a genuinely different value from set A.
func TestGetNodeOptions_SetB(t *testing.T) {
	// Disk-based storage requires a real path; t.TempDir() is cleaned up automatically.
	dbPath := C.CString(t.TempDir())
	defer C.free(unsafe.Pointer(dbPath))

	// Allocate C strings
	retryIntervals := C.CString("15,30")
	defer C.free(unsafe.Pointer(retryIntervals))
	listenAddrs := C.CString("/ip4/127.0.0.3/tcp/9173")
	defer C.free(unsafe.Pointer(listenAddrs))
	bootstrapPeers := C.CString("/ip4/127.0.0.4/tcp/9174")
	defer C.free(unsafe.Pointer(bootstrapPeers))
	httpAddr := C.CString("localhost:9292")
	defer C.free(unsafe.Pointer(httpAddr))
	allowedOrigins := C.CString("http://b.example.com")
	defer C.free(unsafe.Pointer(allowedOrigins))
	certPath := C.CString("/tls/cert-b.pem")
	defer C.free(unsafe.Pointer(certPath))
	keyPath := C.CString("/tls/key-b.pem")
	defer C.free(unsafe.Pointer(keyPath))
	// DocumentACPType is left unset, defaulting to "local" because of the in-memory flag.
	// DocumentACPPath left unset,  defaulting to ""
	logID := C.CString("log-b")
	defer C.free(unsafe.Pointer(logID))
	grpcAddr := C.CString("grpc-b:9090")
	defer C.free(unsafe.Pointer(grpcAddr))
	cometAddr := C.CString("comet-b:26657")
	defer C.free(unsafe.Pointer(cometAddr))
	nodeACPPath := C.CString("")
	defer C.free(unsafe.Pointer(nodeACPPath))

	// Populate NodeInitOptions
	var opts C.NodeInitOptions
	// inMemory is intentionally not set, resulting in BadgerInMemory staying false (vs true in set A).
	opts.disableP2P = 1
	opts.disableAPI = 1
	// Store
	opts.dbPath = dbPath
	opts.badgerFileSize = C.int64_t(2097152) // 2 MB
	// DB
	// enableSigning is not set, defaulting to true
	// enableNodeACP not set, defaulting to false
	opts.maxTransactionRetries = C.int(14)
	opts.p2pBlockSyncTimeoutMs = C.int64_t(8000)
	opts.lensPoolSize = C.int(6)
	opts.chunkSize = C.int(256)
	opts.replicatorRetryIntervals = retryIntervals
	// searchableEncryptionKey intentionally not set, so is nil in the output (vs "<redacted>" in set A)
	// P2P flags are intentionally not set, and will all default to false (vs true in set A)
	opts.listeningAddresses = listenAddrs
	opts.peers = bootstrapPeers
	// p2pPrivateKey intentionally not set, so is nil in output (vs "<redacted>" in set A)
	// HTTP
	opts.httpAddress = httpAddr
	opts.httpAllowedOrigins = allowedOrigins
	opts.tlsCertPath = certPath
	opts.tlsKeyPath = keyPath
	opts.httpReadTimeoutMs = C.int64_t(4000)
	opts.httpWriteTimeoutMs = C.int64_t(5000)
	opts.httpIdleTimeoutMs = C.int64_t(6000)
	// Document ACP
	opts.remoteDACLogID = logID
	opts.remoteDACGRPCAddress = grpcAddr
	opts.remoteDACCometRPCAddress = cometAddr
	// Node ACP
	// enableNodeACP is not set, defaulting to false (vs true in set A)
	opts.nodeACPPath = nodeACPPath

	m := startTestNodeAndGetOptionMap(t, opts)

	// Node-level
	assertNodeOptionField(t, m, true, "DisableP2P")
	assertNodeOptionField(t, m, true, "DisableAPI")

	// Store
	assertNodeOptionField(t, m, "", "Store", "Store")
	assertNodeOptionField(t, m, false, "Store", "BadgerInMemory")
	assertNodeOptionField(t, m, float64(2097152), "Store", "BadgerFileSize")

	// DB
	assertNodeOptionField(t, m, true, "DB", "EnableSigning")
	assertNodeOptionField(t, m, nil, "DB", "Identity")
	assertNodeOptionField(t, m, float64(14), "DB", "MaxTxnRetries")
	assertNodeOptionField(t, m, float64(8000000000), "DB", "P2PBlockSyncTimeout") // 8000 ms
	assertNodeOptionField(t, m, float64(6), "DB", "LensPoolSize")
	assertNodeOptionField(t, m, float64(256), "DB", "ChunkSize")
	assertNodeOptionField(t, m, "wazero", "DB", "LensRuntime")
	assertNodeOptionField(t, m, nil, "DB", "SearchableEncryptionKey")
	assertNodeOptionArrayField(t, m, []any{float64(15000000000), float64(30000000000)}, "DB", "RetryIntervals")

	// P2P — boolean flags not set → all false (vs true in set A)
	assertNodeOptionField(t, m, false, "P2P", "EnablePubSub")
	assertNodeOptionField(t, m, false, "P2P", "EnableRelay")
	assertNodeOptionField(t, m, false, "P2P", "EnableClearBackoffOnRetry")
	assertNodeOptionArrayField(t, m, []any{"/ip4/127.0.0.3/tcp/9173"}, "P2P", "ListenAddresses")
	assertNodeOptionArrayField(t, m, []any{"/ip4/127.0.0.4/tcp/9174"}, "P2P", "BootstrapPeers")
	assertNodeOptionField(t, m, nil, "P2P", "PrivateKey") // not set → nil (vs "<redacted>" in set A)

	// HTTP
	assertNodeOptionField(t, m, "localhost:9292", "HTTP", "Address")
	assertNodeOptionArrayField(t, m, []any{"http://b.example.com"}, "HTTP", "AllowedOrigins")
	assertNodeOptionField(t, m, "/tls/cert-b.pem", "HTTP", "TLSCertPath")
	assertNodeOptionField(t, m, "/tls/key-b.pem", "HTTP", "TLSKeyPath")
	assertNodeOptionField(t, m, float64(4000000000), "HTTP", "ReadTimeout")  // 4000 ms
	assertNodeOptionField(t, m, float64(5000000000), "HTTP", "WriteTimeout") // 5000 ms
	assertNodeOptionField(t, m, float64(6000000000), "HTTP", "IdleTimeout")  // 6000 ms

	// Document ACP
	// Type is not set, defaulting to "local" (vs explicit "local" in set A)
	assertNodeOptionField(t, m, "local", "DocumentACP", "DocumentACPType")
	assertNodeOptionField(t, m, "", "DocumentACP", "Path")
	assertNodeOptionField(t, m, "log-b", "DocumentACP", "RemoteDACLogID")
	assertNodeOptionField(t, m, "grpc-b:9090", "DocumentACP", "RemoteDACGRPCAddress")
	assertNodeOptionField(t, m, "comet-b:26657", "DocumentACP", "RemoteDACCometRPCAddress")

	// Node ACP
	// enableNodeACP is not set, defaulting to false (vs true in set A)
	assertNodeOptionField(t, m, false, "NodeACP", "IsEnabled")
	assertNodeOptionField(t, m, "", "NodeACP", "Path")
}
