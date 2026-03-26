// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !android

// This file is used to stub the BLE driver for non-Android builds.

package cbindings

import (
	proximity "berty.tech/weshnet/v2/pkg/proximitytransport"
	"github.com/sourcenetwork/defradb/internal/bledriver"
)

func init() {
	bledriver.Driver = &NoopJNIDriver{}
}

type NoopJNIDriver struct{}

func (d *NoopJNIDriver) Start(localPID string)                            {}
func (d *NoopJNIDriver) Stop()                                            {}
func (d *NoopJNIDriver) DialPeer(remotePID string) bool                   { return false }
func (d *NoopJNIDriver) SendToPeer(remotePID string, payload []byte) bool { return false }
func (d *NoopJNIDriver) CloseConnWithPeer(remotePID string)               {}
func (d *NoopJNIDriver) ProtocolCode() int                                { return 0x0042 }
func (d *NoopJNIDriver) ProtocolName() string                             { return "ble" }
func (d *NoopJNIDriver) DefaultAddr() string {
	return "/ble/Qmeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
}

var _ proximity.ProximityDriver = &NoopJNIDriver{}
