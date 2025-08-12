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
    unsigned long long tx;
    const char* version;
    const char* collectionID;
    const char* name;
    const char* identity;
    int getInactive;
} CollectionOptions;

typedef struct {
    const char* dbPath;
    const char* listeningAddresses;
    const char* replicatorRetryIntervals;
    const char* peers;
    const char* identityKeyType;
    const char* identityPrivateKey;
    int inMemory;
    int disableP2P;
    int disableAPI;
    int enableNodeACP;
    int maxTransactionRetries;
} NodeInitOptions;

#endif // DEFRA_STRUCTS_H
