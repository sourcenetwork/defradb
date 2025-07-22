// defra_structs.h
#ifndef DEFRA_STRUCTS_H
#define DEFRA_STRUCTS_H

typedef struct {
    int status;
    char* error;
    char* value;
} Result;

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
    int maxTransactionRetries;
} NodeInitOptions;

#endif // DEFRA_STRUCTS_H
