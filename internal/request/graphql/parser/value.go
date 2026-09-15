// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.

package parser

import "github.com/sourcenetwork/defradb/client"

func parseStringList(name string, value any) ([]string, error) {
	values, ok := value.([]any)
	if !ok {
		if strings, ok := value.([]string); ok {
			return strings, nil
		}
		return nil, client.NewErrUnexpectedType[[]string](name, value)
	}

	strings := make([]string, len(values))
	for i, value := range values {
		stringValue, ok := value.(string)
		if !ok {
			return nil, client.NewErrUnexpectedType[string](name, value)
		}
		strings[i] = stringValue
	}
	return strings, nil
}
