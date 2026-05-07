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

// PollSubscription will get the subscription object associated with an ID, and if
// it exists will see if there's a message in its result channel. If there isn't, it will
// return with status 2, and a blank payload. If there is, it will return with status 0,
// and the payload of the message. If an error occurs, status 1 is returned.

//export PollSubscription
func PollSubscription(id *C.char) C.Result {
	subID := C.GoString(id)
	sub, ok := getSubscription(subID)
	if !ok {
		return returnC(returnGoC(1, NewErrInvalidSubscriptionID(subID).Error(), ""))
	}
	select {
	case msg, ok := <-sub.resultChan:
		if !ok {
			removeSubscription(subID)
			return returnC(returnGoC(1, errGettingSubscription, ""))
		}
		return returnC(marshalJSONToGoCResult(msg))
	default:
		return returnC(returnGoC(2, "", ""))
	}
}
