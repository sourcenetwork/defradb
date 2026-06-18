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

import "github.com/sourcenetwork/defradb/client"

func sourceHubObjectID(objectID string) string {
	docID, err := client.NewDocIDFromString(objectID)
	if err != nil {
		return objectID
	}
	return docID.UUID().String()
}
