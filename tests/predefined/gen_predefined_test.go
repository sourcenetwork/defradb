// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package predefined

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client/request"
)

func TestGeneratePredefinedFromSchema_Simple(t *testing.T) {
	schema := `
		type User {
			name: String
			age: Int
		}`

	docsList := DocsList{
		ColName: "User",
		Docs: []map[string]any{
			{"name": "John", "age": 30},
			{"name": "Fred", "age": 25},
		},
	}
	docs, err := CreateFromSDL(schema, docsList)
	assert.NoError(t, err)

	colDefMap, err := parseSDL(schema)
	require.NoError(t, err)

	errorMsg := assertDocs(mustAddDocIDsToDocs(docsList.Docs, colDefMap["User"].Schema), docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

func TestGeneratePredefinedFromSchema_StripExcessiveFields(t *testing.T) {
	schema := `
		type User {
			name: String
		}`

	docs, err := CreateFromSDL(schema, DocsList{
		ColName: "User",
		Docs: []map[string]any{
			{"name": "John", "age": 30},
			{"name": "Fred", "age": 25},
		},
	})
	assert.NoError(t, err)

	colDefMap, err := parseSDL(schema)
	require.NoError(t, err)

	errorMsg := assertDocs(mustAddDocIDsToDocs([]map[string]any{
		{"name": "John"},
		{"name": "Fred"},
	}, colDefMap["User"].Schema), docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

func TestGeneratePredefinedFromSchema_OneToOne(t *testing.T) {
	schema := `
		type User {
			name: String
			device: Device
		}
		type Device {
			model: String
			owner: User
		}`

	docs, err := CreateFromSDL(schema, DocsList{
		ColName: "User",
		Docs: []map[string]any{
			{
				"name": "John",
				"device": map[string]any{
					"model": "iPhone",
				},
			},
			{
				"name": "Fred",
				"device": map[string]any{
					"model": "MacBook",
				},
			},
		},
	})
	assert.NoError(t, err)

	colDefMap, err := parseSDL(schema)
	require.NoError(t, err)

	userDocs := mustAddDocIDsToDocs([]map[string]any{
		{"name": "John"},
		{"name": "Fred"},
	}, colDefMap["User"].Schema)

	deviceDocs := mustAddDocIDsToDocs([]map[string]any{
		{
			"model":    "iPhone",
			"owner_id": mustGetDocIDFromDocMap(map[string]any{"name": "John"}, colDefMap["User"].Schema),
		},
		{
			"model":    "MacBook",
			"owner_id": mustGetDocIDFromDocMap(map[string]any{"name": "Fred"}, colDefMap["User"].Schema),
		},
	}, colDefMap["Device"].Schema)

	errorMsg := assertDocs(append(userDocs, deviceDocs...), docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

func TestGeneratePredefinedFromSchema_OneToOnePrimary(t *testing.T) {
	schema := `
		type User {
			name: String
			device: Device @primary
		}
		type Device {
			model: String
			owner: User
		}`

	docs, err := CreateFromSDL(schema, DocsList{
		ColName: "User",
		Docs: []map[string]any{
			{
				"name": "John",
				"device": map[string]any{
					"model": "iPhone",
				},
			},
			{
				"name": "Fred",
				"device": map[string]any{
					"model": "MacBook",
				},
			},
		},
	})
	assert.NoError(t, err)

	colDefMap, err := parseSDL(schema)
	require.NoError(t, err)

	userDocs := mustAddDocIDsToDocs([]map[string]any{
		{
			"name":      "John",
			"device_id": mustGetDocIDFromDocMap(map[string]any{"model": "iPhone"}, colDefMap["Device"].Schema),
		},
		{
			"name":      "Fred",
			"device_id": mustGetDocIDFromDocMap(map[string]any{"model": "MacBook"}, colDefMap["Device"].Schema),
		},
	}, colDefMap["User"].Schema)
	deviceDocs := mustAddDocIDsToDocs([]map[string]any{
		{"model": "iPhone"},
		{"model": "MacBook"},
	}, colDefMap["Device"].Schema)

	errorMsg := assertDocs(append(userDocs, deviceDocs...), docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

func TestGeneratePredefinedFromSchema_OneToOneToOnePrimary(t *testing.T) {
	schema := `
		type User {
			name: String
			device: Device @primary
		}
		type Device {
			model: String
			owner: User
			specs: Specs @primary
		}
		type Specs {
			OS: String
			device: Device
		}`

	docs, err := CreateFromSDL(schema, DocsList{
		ColName: "User",
		Docs: []map[string]any{
			{
				"name": "John",
				"device": map[string]any{
					"model": "iPhone",
					"specs": map[string]any{
						"OS": "iOS",
					},
				},
			},
		},
	})
	assert.NoError(t, err)

	colDefMap, err := parseSDL(schema)
	require.NoError(t, err)

	specsDoc := mustAddDocIDToDoc(map[string]any{"OS": "iOS"}, colDefMap["Specs"].Schema)
	deviceDoc := mustAddDocIDToDoc(map[string]any{
		"model":    "iPhone",
		"specs_id": specsDoc[request.DocIDFieldName],
	}, colDefMap["Device"].Schema)
	userDoc := mustAddDocIDToDoc(map[string]any{
		"name":      "John",
		"device_id": deviceDoc[request.DocIDFieldName],
	}, colDefMap["User"].Schema)

	errorMsg := assertDocs([]map[string]any{userDoc, deviceDoc, specsDoc}, docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

func TestGeneratePredefinedFromSchema_TwoPrimaryToOneMiddle(t *testing.T) {
	schema := `
		type User {
			name: String
			device: Device 
		}
		type Device {
			model: String
			owner: User @primary
			specs: Specs @primary
		}
		type Specs {
			OS: String
			device: Device
		}`

	docs, err := CreateFromSDL(schema, DocsList{
		ColName: "User",
		Docs: []map[string]any{
			{
				"name": "John",
				"device": map[string]any{
					"model": "iPhone",
					"specs": map[string]any{
						"OS": "iOS",
					},
				},
			},
		},
	})
	assert.NoError(t, err)

	colDefMap, err := parseSDL(schema)
	require.NoError(t, err)

	specsDoc := mustAddDocIDToDoc(map[string]any{"OS": "iOS"}, colDefMap["Specs"].Schema)
	userDoc := mustAddDocIDToDoc(map[string]any{"name": "John"}, colDefMap["User"].Schema)
	deviceDoc := mustAddDocIDToDoc(map[string]any{
		"model":    "iPhone",
		"specs_id": specsDoc[request.DocIDFieldName],
		"owner_id": userDoc[request.DocIDFieldName],
	}, colDefMap["Device"].Schema)

	errorMsg := assertDocs([]map[string]any{userDoc, deviceDoc, specsDoc}, docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

func TestGeneratePredefinedFromSchema_OneToTwoPrimary(t *testing.T) {
	schema := `
		type User {
			name: String
			device: Device @primary
		}
		type Device {
			model: String
			owner: User
			specs: Specs
		}
		type Specs {
			OS: String
			device: Device @primary
		}`

	docs, err := CreateFromSDL(schema, DocsList{
		ColName: "User",
		Docs: []map[string]any{
			{
				"name": "John",
				"device": map[string]any{
					"model": "iPhone",
					"specs": map[string]any{
						"OS": "iOS",
					},
				},
			},
		},
	})
	assert.NoError(t, err)

	colDefMap, err := parseSDL(schema)
	require.NoError(t, err)

	deviceDoc := mustAddDocIDToDoc(map[string]any{"model": "iPhone"}, colDefMap["Device"].Schema)
	specsDoc := mustAddDocIDToDoc(map[string]any{
		"OS":        "iOS",
		"device_id": deviceDoc[request.DocIDFieldName],
	}, colDefMap["Specs"].Schema)
	userDoc := mustAddDocIDToDoc(map[string]any{
		"name":      "John",
		"device_id": deviceDoc[request.DocIDFieldName],
	}, colDefMap["User"].Schema)

	errorMsg := assertDocs([]map[string]any{userDoc, deviceDoc, specsDoc}, docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

func TestGeneratePredefinedFromSchema_TwoPrimaryToOneRoot(t *testing.T) {
	schema := `
		type User {
			name: String
			device: Device @primary
			address: Address @primary
		}
		type Device {
			model: String
			owner: User
		}
		type Address {
			street: String
			user: User 
		}`

	docs, err := CreateFromSDL(schema, DocsList{
		ColName: "User",
		Docs: []map[string]any{
			{
				"name": "John",
				"device": map[string]any{
					"model": "iPhone",
				},
				"address": map[string]any{
					"street": "Backer",
				},
			},
		},
	})
	assert.NoError(t, err)

	colDefMap, err := parseSDL(schema)
	require.NoError(t, err)

	deviceDoc := mustAddDocIDToDoc(map[string]any{"model": "iPhone"}, colDefMap["Device"].Schema)
	addressDoc := mustAddDocIDToDoc(map[string]any{"street": "Backer"}, colDefMap["Address"].Schema)
	userDoc := mustAddDocIDToDoc(map[string]any{
		"name":       "John",
		"device_id":  deviceDoc[request.DocIDFieldName],
		"address_id": addressDoc[request.DocIDFieldName],
	}, colDefMap["User"].Schema)

	errorMsg := assertDocs([]map[string]any{userDoc, deviceDoc, addressDoc}, docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

// func TestGeneratePredefinedFromSchema_OneToMany(t *testing.T) {
// 	schema := `
// 		type User {
// 			name: String
// 			devices: [Device]
// 		}
// 		type Device {
// 			model: String
// 			owner: User
// 		}`

// 	docs, err := CreateFromSDL(schema, DocsList{
// 		ColName: "User",
// 		Docs: []map[string]any{
// 			{
// 				"name": "John",
// 				"devices": []map[string]any{
// 					{"model": "iPhone"},
// 					{"model": "PlayStation"},
// 				},
// 			},
// 			{
// 				"name": "Fred",
// 				"devices": []map[string]any{
// 					{"model": "Surface"},
// 					{"model": "Pixel"},
// 				},
// 			},
// 		},
// 	})
// 	assert.NoError(t, err)

// 	colDefMap, err := parseSDL(schema)
// 	require.NoError(t, err)

// 	johnDocID := mustGetDocIDFromDocMap(map[string]any{"name": "John"}, colDefMap["User"].Schema)
// 	fredDocID := mustGetDocIDFromDocMap(map[string]any{"name": "Fred"}, colDefMap["User"].Schema)
// 	errorMsg := assertDocs(mustAddDocIDsToDocs([]map[string]any{
// 		{"name": "John"},
// 		{"name": "Fred"},
// 		{"model": "iPhone", "owner_id": johnDocID},
// 		{"model": "PlayStation", "owner_id": johnDocID},
// 		{"model": "Surface", "owner_id": fredDocID},
// 		{"model": "Pixel", "owner_id": fredDocID},
// 	}, col), docs)
// 	if errorMsg != "" {
// 		t.Error(errorMsg)
// 	}
// }

// func TestGeneratePredefinedFromSchema_OneToManyToOne(t *testing.T) {
// 	schema := `
// 		type User {
// 			name: String
// 			devices: [Device]
// 		}
// 		type Device {
// 			model: String
// 			owner: User
// 			specs: Specs
// 		}
// 		type Specs {
// 			CPU: String
// 			device: Device @primary
// 		}`

// 	docs, err := CreateFromSDL(schema, DocsList{
// 		ColName: "User",
// 		Docs: []map[string]any{
// 			{
// 				"name": "John",
// 				"devices": []map[string]any{
// 					{
// 						"model": "iPhone",
// 						"specs": map[string]any{
// 							"CPU": "A13",
// 						},
// 					},
// 					{
// 						"model": "MacBook",
// 						"specs": map[string]any{
// 							"CPU": "M2",
// 						},
// 					},
// 				},
// 			},
// 		},
// 	})
// 	assert.NoError(t, err)

// 	colDefMap, err := parseSDL(schema)
// 	require.NoError(t, err)

// 	johnDocID := mustGetDocIDFromDocMap(map[string]any{"name": "John"}, colDefMap["User"].Schema)
// 	errorMsg := assertDocs(mustAddDocIDsToDocs([]map[string]any{
// 		{"name": "John"},
// 		{"model": "iPhone", "owner_id": johnDocID},
// 		{"model": "MacBook", "owner_id": johnDocID},
// 		{
// 			"CPU": "A13",
// 			"device_id": mustGetDocIDFromDocMap(map[string]any{
// 				"model":    "iPhone",
// 				"owner_id": johnDocID,
// 			}, colDefMap["Device"].Schema),
// 		},
// 		{
// 			"CPU": "M2",
// 			"device_id": mustGetDocIDFromDocMap(map[string]any{
// 				"model":    "MacBook",
// 				"owner_id": johnDocID,
// 			}, colDefMap["Device"].Schema),
// 		},
// 	}), docs)
// 	if errorMsg != "" {
// 		t.Error(errorMsg)
// 	}
// }

// func TestGeneratePredefined_OneToMany(t *testing.T) {
// 	defs := []client.CollectionDefinition{
// 		{
// 			Description: client.CollectionDescription{
// 				Name: "User",
// 				ID:   0,
// 			},
// 			Schema: client.SchemaDescription{
// 				Name: "User",
// 				Fields: []client.FieldDescription{
// 					{
// 						Name: "name",
// 						Kind: client.FieldKind_STRING,
// 					},
// 					{
// 						Name:         "devices",
// 						Kind:         client.FieldKind_FOREIGN_OBJECT_ARRAY,
// 						Schema:       "Device",
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Description: client.CollectionDescription{
// 				Name: "Device",
// 				ID:   1,
// 			},
// 			Schema: client.SchemaDescription{
// 				Name: "Device",
// 				Fields: []client.FieldDescription{
// 					{
// 						Name: "model",
// 						Kind: client.FieldKind_STRING,
// 					},
// 					{
// 						Name:   "owner",
// 						Kind:   client.FieldKind_FOREIGN_OBJECT,
// 						Schema: "User",
// 						IsPrimary: true,
// 					},
// 				},
// 			},
// 		},
// 	}
// 	docs, err := Create(defs, DocsList{
// 		ColName: "User",
// 		Docs: []map[string]any{
// 			{
// 				"name": "John",
// 				"devices": []map[string]any{
// 					{"model": "iPhone"},
// 					{"model": "PlayStation"},
// 				},
// 			},
// 			{
// 				"name": "Fred",
// 				"devices": []map[string]any{
// 					{"model": "Surface"},
// 					{"model": "Pixel"},
// 				},
// 			},
// 		},
// 	})
// 	assert.NoError(t, err)

// 	johnDocID := mustGetDocIDFromDocMap(map[string]any{"name": "John"}, defs[0].Schema)
// 	fredDocID := mustGetDocIDFromDocMap(map[string]any{"name": "Fred"}, defs[0].Schema)
// 	errorMsg := assertDocs(mustAddDocIDsToDocs([]map[string]any{
// 		{"name": "John"},
// 		{"name": "Fred"},
// 		{"model": "iPhone", "owner_id": johnDocID},
// 		{"model": "PlayStation", "owner_id": johnDocID},
// 		{"model": "Surface", "owner_id": fredDocID},
// 		{"model": "Pixel", "owner_id": fredDocID},
// 	}), docs)
// 	if errorMsg != "" {
// 		t.Error(errorMsg)
// 	}
// }
