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
	"fmt"
	"unsafe"

	proximity "berty.tech/weshnet/v2/pkg/proximitytransport"
	"github.com/sourcenetwork/defradb/internal/bledriver"
)

//export BLEHandleFoundPeer
func BLEHandleFoundPeer(remotePID *C.char) C.int {
	t := bledriver.GetTransport()
	if t == nil {
		fmt.Println("BLE_LOG: BLEHandleFoundPeer: transport is nil")
		return 0
	}
	pid := C.GoString(remotePID)
	fmt.Println("BLE_LOG: BLEHandleFoundPeer: calling HandleFoundPeer for", pid)
	if t.HandleFoundPeer(pid) {
		fmt.Println("BLE_LOG: BLEHandleFoundPeer: returned true")
		return 1
	}
	fmt.Println("BLE_LOG: BLEHandleFoundPeer: returned false")
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
