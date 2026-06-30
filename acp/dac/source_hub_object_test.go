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

func TestSourceHubObjectID(t *testing.T) {
	const docID = "bae-9793af00-a131-5ef2-b2c9-22b8053a11e7"

	if got := sourceHubObjectID(docID); got != "9793af00-a131-5ef2-b2c9-22b8053a11e7" {
		t.Fatalf("Expected SourceHub object ID to use the DocID UUID, got %q", got)
	}

	if got := sourceHubObjectID("custom-object-id"); got != "custom-object-id" {
		t.Fatalf("Expected non-DocID object ID to pass through, got %q", got)
	}
}
