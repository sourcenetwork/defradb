// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package message

import "github.com/sourcenetwork/defradb/internal/wire"

func init() {
	// Every message sent over a stream is CBOR-encoded through this interface, so
	// its concrete type is not visible at the encode site. Registering the
	// interface acknowledges that its implementers cross the wire; each concrete
	// implementer is registered in its own package.
	wire.Register[Message]()
}
