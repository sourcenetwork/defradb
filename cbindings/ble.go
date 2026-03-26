// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build android

package cbindings

/*
#include <stdlib.h>
#include <jni.h>

extern void DefraBLE_SetBleInterface(JavaVM* jvm, jobject bleInterface);
extern void CallStart(const char* localPID);
extern jboolean CallDialPeer(const char* remotePID);
extern jboolean CallSendToPeer(const char* remotePID, void* payload, int length);
extern void CallCloseConnWithPeer(const char* remotePID);
*/
import "C"

import (
	"unsafe"

	proximity "berty.tech/weshnet/v2/pkg/proximitytransport"
	"github.com/sourcenetwork/defradb/internal/bledriver"
)

//export BLEHandleFoundPeer
func BLEHandleFoundPeer(remotePID *C.char) C.int {
	t := bledriver.GetTransport()
	if t == nil {
		return 0
	}
	pid := C.GoString(remotePID)
	if t.HandleFoundPeer(pid) {
		return 1
	}
	return 0
}

//export BLEHandleLostPeer
func BLEHandleLostPeer(remotePID *C.char) {
	t := getTransport()
	if t == nil {
		return
	}
	t.HandleLostPeer(C.GoString(remotePID))
}

//export BLEReceiveFromPeer
func BLEReceiveFromPeer(remotePID *C.char, payload unsafe.Pointer, length C.int) {
	t := getTransport()
	if t == nil {
		return
	}
	data := C.GoBytes(payload, length)
	t.ReceiveFromPeer(C.GoString(remotePID), data)
}

func getTransport() proximity.ProximityTransport {
	return bledriver.GetTransport()
}
