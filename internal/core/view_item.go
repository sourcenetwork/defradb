// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package core

import (
	"encoding/json"

	"github.com/sourcenetwork/defradb/client"
)

// MarshalViewItem marshals the given doc ready for storage.
//
// It first trims the doc leaving only an array of field values (including
// relations), and then marshals that into json.
//
// Note: MarshalViewItem and UnmarshalViewItem rely on the Doc (and DocumentMapping)
// being consistent at write and read time.
func MarshalViewItem(doc Doc) ([]byte, error) {
	trimmedDoc := trimDoc(doc)
	return json.Marshal(trimmedDoc)
}

func trimDoc(doc Doc) []any {
	fields := make([]any, 0, len(doc.Fields))
	for _, field := range doc.Fields {
		switch typedField := field.(type) {
		case []Doc:
			trimmedField := make([]any, 0, len(typedField))
			for _, innerDoc := range typedField {
				trimmedField = append(
					trimmedField,
					trimDoc(innerDoc),
				)
			}
			fields = append(fields, trimmedField)

		case Doc:
			fields = append(
				fields,
				trimDoc(typedField),
			)

		default:
			fields = append(fields, typedField)
		}
	}

	return fields
}

// UnmarshalViewItem unmarshals the given byte array into a [Doc] using the given
// mapping.
//
// It assumes that `bytes` is in the appropriate format (see MarshalViewItem) and
// will only error if the json unmarshalling fails.
//
// Note: MarshalViewItem and UnmarshalViewItem rely on the Doc (and DocumentMapping)
// being consistent at write and read time.
func UnmarshalViewItem(documentMap *DocumentMapping, bytes []byte) (Doc, error) {
	var trimmedDoc []any
	err := json.Unmarshal(bytes, &trimmedDoc)
	if err != nil {
		return Doc{}, err
	}

	return expandViewItem(documentMap, trimmedDoc), nil
}

func expandViewItem(documentMap *DocumentMapping, trimmed []any) Doc {
	fields := make(DocFields, len(trimmed))

	for _, indexes := range documentMap.IndexesByName {
		for _, index := range indexes {
			fieldValue := trimmed[index]
			var childMapping *DocumentMapping
			if index < len(documentMap.ChildMappings) {
				childMapping = documentMap.ChildMappings[index]
			}

			if childMapping == nil {
				// If the childMapping is nil, this property must not be a relation and we can
				// set the value and continue.
				fields[index] = fieldValue
				continue
			}

			if untypedArray, ok := fieldValue.([]any); ok {
				isArrayOfDocs := true
				for _, inner := range untypedArray {
					if _, ok := inner.([]any); !ok {
						// To know if this is an array of documents we need to check the inner values to see if
						// this is esentially an `[][]any`
						isArrayOfDocs = false
						break
					}
				}

				if isArrayOfDocs {
					innerDocs := make([]Doc, 0, len(untypedArray))
					for _, inner := range untypedArray {
						innerDocs = append(innerDocs, expandViewItem(childMapping, inner.([]any)))
					}
					fields[index] = innerDocs
				} else {
					fields[index] = expandViewItem(childMapping, untypedArray)
				}
			}
			// else: no-op
			//
			// The relation must be either an empty array (many side of one-many), or
			// nil (one side of either a one-many or one-one).  Either way the value is nil/default
			// and we can continue
		}
	}

	return Doc{
		Hidden: false,
		Fields: fields,
		Status: client.Active,
	}
}
