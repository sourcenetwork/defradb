// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

#include <stdio.h>
#include <stdlib.h>
#include "defra_structs.h"
#include "libdefradb.h"

// Basic smoke test for the C bindings
// This will simply try tocreate a node, and close it.
// The return status of 1 will propogate to the caller (presumably the CI runner)
int main() {
    NodeInitOptions nodeOpts = {0};
    nodeOpts.inMemory = 1;
    nodeOpts.disableP2P = 1;
    nodeOpts.disableAPI = 1;
    nodeOpts.enableNodeACP = 0;

    NewNodeResult node = NewNode(opts);
    if (node.status != 0) {
        fprintf(stderr, "NewNode failed: %s\n", node.error);
        return 1;
    }

    Result close_result = CloseNode(node.nodePtr);
    if (close_result.status != 0) {
        fprintf(stderr, "CloseNode failed: %s\n", close_result.error);
        return 1;
    }

    printf("Basic C smoke test passed\n");
    return 0;
}
