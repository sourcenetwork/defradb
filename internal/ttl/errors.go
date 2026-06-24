// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package ttl

import "github.com/sourcenetwork/defradb/errors"

var (
	// ErrInvalidSlotCount is returned when a wheel is configured with no slots.
	ErrInvalidSlotCount = errors.New("invalid slot count, must be greater than 0")
	// ErrInvalidTick is returned when a wheel is configured with no tick duration.
	ErrInvalidTick = errors.New("invalid tick rate, must be greater than 0")
	// ErrNegativeTTL is returned when an entry is stored with a negative TTL.
	ErrNegativeTTL = errors.New("ttl value can not be negative")
	// ErrBeyondMaxTTL is returned when an entry TTL exceeds the wheel's maximum window.
	ErrBeyondMaxTTL = errors.New("ttl larger than max allowed")
)
