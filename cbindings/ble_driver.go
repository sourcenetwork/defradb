package cbindings

/*
#include <stdlib.h>
#include <jni.h>

extern void DefraBLE_SetBleInterface(JavaVM* jvm, jobject bleInterface);
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

func init() {
	bledriver.Driver = &JNIDriver{}
}

type JNIDriver struct{}

func (d *JNIDriver) Start(localPID string) {
	// No-op because handling starting should occur on the Java side
}

func (d *JNIDriver) Stop() {
	// No-op because handling stopping should occur on the Java side
}

func (d *JNIDriver) DialPeer(remotePID string) bool {
	cPID := C.CString(remotePID)
	defer C.free(unsafe.Pointer(cPID))
	return C.CallDialPeer(cPID) != 0
}

func (d *JNIDriver) SendToPeer(remotePID string, payload []byte) bool {
	if len(payload) == 0 {
		return true
	}
	cPID := C.CString(remotePID)
	defer C.free(unsafe.Pointer(cPID))
	ptr := unsafe.Pointer(&payload[0])
	return C.CallSendToPeer(cPID, ptr, C.int(len(payload))) != 0
}

func (d *JNIDriver) CloseConnWithPeer(remotePID string) {
	cPID := C.CString(remotePID)
	defer C.free(unsafe.Pointer(cPID))
	C.CallCloseConnWithPeer(cPID)
}

func (d *JNIDriver) ProtocolCode() int    { return 0x0042 }
func (d *JNIDriver) ProtocolName() string { return "ble" }
func (d *JNIDriver) DefaultAddr() string {
	return "/ble/Qmeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
}

var _ proximity.ProximityDriver = &JNIDriver{}
