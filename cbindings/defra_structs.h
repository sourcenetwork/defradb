// defra_structs.h
#ifndef DEFRA_STRUCTS_H
#define DEFRA_STRUCTS_H
#include <stdint.h>

typedef struct {
    int status;
    char* error;
    char* value;
} Result;

typedef struct {
    int status;
    char* error;
    uintptr_t nodePtr;
} NewNodeResult;

typedef struct {
    int status;
    char* error;
    uintptr_t txnPtr;
} NewTxnResult;

typedef struct {
    int status;
    char* error;
    uintptr_t identityPtr;
} NewIdentityResult;

typedef struct {
    const char* version;
    const char* collectionID;
    const char* name;
    int getInactive;
    int enableSigning; // 0 unset, 1 enabled, -1 disabled
} CollectionOptions;

typedef struct {
    // Core / node-level
    const char* dbPath;
    const char* listeningAddresses;
    const char* replicatorRetryIntervals;
    const char* peers;
    uintptr_t identityPtr;
    int inMemory;
    int disableP2P;
    int disableAPI;
    int enableNodeACP;
    int maxTransactionRetries;

    // Store options
    const char* storeType;              // "badger" | "memory" | "" (default)
    int64_t badgerFileSize;             // 0 = use default
    const uint8_t* badgerEncryptionKey;
    int badgerEncryptionKeyLen;

    // DB options
    int enableSigning;
    const uint8_t* searchableEncryptionKey;
    int searchableEncryptionKeyLen;
    int64_t p2pBlockSyncTimeoutMs;      // 0 = use default
    int lensPoolSize;                   // 0 = use default
    int chunkSize;                      // 0 = use default

    // P2P options
    int enablePubSub;
    int enableRelay;
    int enableClearBackoffOnRetry;
    const uint8_t* p2pPrivateKey;
    int p2pPrivateKeyLen;

    // HTTP options
    const char* httpAddress;
    const char* httpAllowedOrigins;     // comma-separated
    const char* tlsCertPath;
    const char* tlsKeyPath;
    int64_t httpReadTimeoutMs;          // 0 = use default
    int64_t httpWriteTimeoutMs;         // 0 = use default
    int64_t httpIdleTimeoutMs;          // 0 = use default

    // Document ACP options
    const char* documentACPType;        // "none" | "local" | "source-hub" | "" (default)
    const char* documentACPPath;
    const char* sourceHubChainID;
    const char* sourceHubGRPCAddress;
    const char* sourceHubCometRPCAddress;

    // Node ACP options
    const char* nodeACPPath;
} NodeInitOptions;

#endif // DEFRA_STRUCTS_H
