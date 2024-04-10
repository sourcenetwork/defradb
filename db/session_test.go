// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionWithTxn(t *testing.T) {
	sess := NewSession(context.Background()).WithTxn(&explicitTxn{})
	_, ok := sess.Value(TxnContextKey{}).(*explicitTxn)
	assert.True(t, ok)
}
