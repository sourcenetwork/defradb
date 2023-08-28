// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

import "errors"

// Errors returnable from this package.
//
// This list is incomplete. Undefined errors may also be returned.
// Errors returned from this package may be tested against these errors with errors.Is.
var (
	ErrNoListener           = errors.New("cannot serve with no listener")
	ErrSchema               = errors.New("base must start with the http or https scheme")
	ErrDatabaseNotAvailable = errors.New("no database available")
	ErrFormNotSupported     = errors.New("content type application/x-www-form-urlencoded not yet supported")
	ErrBodyEmpty            = errors.New("body cannot be empty")
	ErrMissingGQLRequest    = errors.New("missing GraphQL request")
	ErrPeerIdUnavailable    = errors.New("no PeerID available. P2P might be disabled")
	ErrStreamingUnsupported = errors.New("streaming unsupported")
	ErrNoEmail              = errors.New("email address must be specified for tls with autocert")
	ErrPayloadFormat        = errors.New("invalid payload format")
	ErrMissingNewKey        = errors.New("missing _newKey for imported doc")
)
