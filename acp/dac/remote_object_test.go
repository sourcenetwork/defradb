// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package dac

import "testing"

func TestRemoteDACObjectID(t *testing.T) {
	const docID = "bae-9793af00-a131-5ef2-b2c9-22b8053a11e7"

	if got := remoteDACObjectID(docID); got != "9793af00-a131-5ef2-b2c9-22b8053a11e7" {
		t.Fatalf("expected Vera object ID to use the DocID UUID, got %q", got)
	}

	if got := remoteDACObjectID("custom-object-id"); got != "custom-object-id" {
		t.Fatalf("Expected non-DocID object ID to pass through, got %q", got)
	}
}
