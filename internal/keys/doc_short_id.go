// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keys

import "github.com/sourcenetwork/defradb/internal/encoding"

// EncodeDocShortID returns the sortable path encoding for a local document storage ID.
func EncodeDocShortID(docShortID uint64) []byte {
	if docShortID == 0 {
		return nil
	}
	return encoding.EncodeUvarintAscending(nil, docShortID)
}

// DecodeDocShortID decodes a local document storage ID from a single encoded key segment.
func DecodeDocShortID(data []byte) (uint64, error) {
	if len(data) == 0 {
		return 0, ErrInvalidKey
	}
	rest, docShortID, err := encoding.DecodeUvarintAscending(data)
	if err != nil {
		return 0, err
	}
	if len(rest) > 0 || docShortID == 0 {
		return 0, ErrInvalidKey
	}
	return docShortID, nil
}
