// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package mobile

import "errors"

// ErrSubscriptionsNotSupported is returned when a subscription request is made
// via the mobile bindings. Subscriptions require persistent connections which
// are not supported in the gomobile environment.
var ErrSubscriptionsNotSupported = errors.New("subscriptions are not supported in mobile bindings")
