package bledriver

import (
	"sync"

	proximity "berty.tech/weshnet/v2/pkg/proximitytransport"
)

// Driver is the global BLE driver instance.
// We set it to a NoopDriver by default, which will safely do nothing.
// If the user wishes to use BLE, they will inject a real driver implementation
// from the Java side.

var Driver proximity.ProximityDriver = &NoopDriver{}

var (
	transport   proximity.ProximityTransport
	transportMu sync.RWMutex
)

func SetTransport(t proximity.ProximityTransport) {
	transportMu.Lock()
	defer transportMu.Unlock()
	transport = t
}

func GetTransport() proximity.ProximityTransport {
	transportMu.RLock()
	defer transportMu.RUnlock()
	return transport
}

// This will return true if a real driver has been injected, false otherwise
func IsEnabled() bool {
	_, ok := Driver.(*NoopDriver)
	return !ok
}

type NoopDriver struct{}

func (d *NoopDriver) Start(localPID string)                            {}
func (d *NoopDriver) Stop()                                            {}
func (d *NoopDriver) DialPeer(remotePID string) bool                   { return false }
func (d *NoopDriver) SendToPeer(remotePID string, payload []byte) bool { return false }
func (d *NoopDriver) CloseConnWithPeer(remotePID string)               {}
func (d *NoopDriver) ProtocolCode() int                                { return 0x0042 }
func (d *NoopDriver) ProtocolName() string                             { return "ble" }
func (d *NoopDriver) DefaultAddr() string {
	return "/ble/Qmeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
}
