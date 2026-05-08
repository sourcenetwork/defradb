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

import (
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errNegativeReplicatorTime       = "negative time intervals are not allowed for replicator retries"
	errAmbiguousCollection          = "more than one collection matches the given criteria"
	errNoDocIDOrFilter              = "operation requires a DocID or filter"
	errInvalidAscensionOrder        = "invalid ascension order: expected ASC or DESC"
	errInvalidIndexFieldDescription = "invalid or malformed field description"
	errInvalidSubscriptionID        = "invalid subscription ID"
	errGettingSubscription          = "could not retrieve subscription"
	errInvalidCGOHandle             = "invalid handle"
)

func NewErrAmbiguousCollection() error {
	return errors.New(errAmbiguousCollection)
}

func NewErrInvalidIndexFieldDescription(field string) error {
	return errors.New(errInvalidIndexFieldDescription, errors.NewKV("Field", field))
}

func NewErrInvalidSubscriptionID(id string) error {
	return errors.New(errInvalidSubscriptionID, errors.NewKV("SubscriptionID", id))
}

func NewErrInvalidCGOHandle(id uintptr) error {
	return errors.New(errInvalidCGOHandle, errors.NewKV("Handle", id))
}
