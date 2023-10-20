// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package gen

import (
	"testing"
)

func TestGeneratePredefinedDocs_Simple(t *testing.T) {
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
	docs := GenerateDocs(schema, docsList)

	errorMsg := assertDocs(docsList.Docs, docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

func TestGeneratePredefinedDocs_StripExcessiveFields(t *testing.T) {
	schema := `
		type User {
			name: String
		}`

	docs := GenerateDocs(schema, DocsList{
		ColName: "User",
		Docs: []map[string]any{
			{"name": "John", "age": 30},
			{"name": "Fred", "age": 25},
		},
	})

	errorMsg := assertDocs([]map[string]any{
		{"name": "John"},
		{"name": "Fred"},
	}, docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

func TestGeneratePredefinedDocs_OneToOne(t *testing.T) {
	schema := `
		type User {
			name: String
			device: Device
		}
		type Device {
			model: String
			owner: User
		}`

	docs := GenerateDocs(schema, DocsList{
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

	errorMsg := assertDocs([]map[string]any{
		{"name": "John"},
		{"name": "Fred"},
		{"model": "iPhone", "owner_id": getDocKeyFromDocMap(map[string]any{"name": "John"})},
		{"model": "MacBook", "owner_id": getDocKeyFromDocMap(map[string]any{"name": "Fred"})},
	}, docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

func TestGeneratePredefinedDocs_OneToOnePrimary(t *testing.T) {
	schema := `
		type User {
			name: String
			device: Device @primary
		}
		type Device {
			model: String
			owner: User
		}`

	docs := GenerateDocs(schema, DocsList{
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

	errorMsg := assertDocs([]map[string]any{
		{"name": "John", "device_id": getDocKeyFromDocMap(map[string]any{"model": "iPhone"})},
		{"name": "Fred", "device_id": getDocKeyFromDocMap(map[string]any{"model": "MacBook"})},
		{"model": "iPhone"},
		{"model": "MacBook"},
	}, docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

func TestGeneratePredefinedDocs_OneToMany(t *testing.T) {
	schema := `
		type User {
			name: String
			device: [Device]
		}
		type Device {
			model: String
			owner: User
		}`

	docs := GenerateDocs(schema, DocsList{
		ColName: "User",
		Docs: []map[string]any{
			{
				"name": "John",
				"device": []map[string]any{
					{"model": "iPhone"},
					{"model": "PlayStation"},
				},
			},
			{
				"name": "Fred",
				"device": []map[string]any{
					{"model": "Surface"},
					{"model": "Pixel"},
				},
			},
		},
	})

	johnDocKey := getDocKeyFromDocMap(map[string]any{"name": "John"})
	fredDocKey := getDocKeyFromDocMap(map[string]any{"name": "Fred"})
	errorMsg := assertDocs([]map[string]any{
		{"name": "John"},
		{"name": "Fred"},
		{"model": "iPhone", "owner_id": johnDocKey},
		{"model": "PlayStation", "owner_id": johnDocKey},
		{"model": "Surface", "owner_id": fredDocKey},
		{"model": "Pixel", "owner_id": fredDocKey},
	}, docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

func TestGeneratePredefinedDocs_OneToManyToOne(t *testing.T) {
	schema := `
		type User {
			name: String
			device: [Device]
		}
		type Device {
			model: String
			owner: User
			specs: Specs
		}
		type Specs {
			CPU: String
			device: Device
		}`

	docs := GenerateDocs(schema, DocsList{
		ColName: "User",
		Docs: []map[string]any{
			{
				"name": "John",
				"device": []map[string]any{
					{
						"model": "iPhone",
						"specs": map[string]any{
							"CPU": "A13",
						},
					},
					{
						"model": "MacBook",
						"specs": map[string]any{
							"CPU": "M2",
						},
					},
				},
			},
		},
	})

	johnDocKey := getDocKeyFromDocMap(map[string]any{"name": "John"})
	errorMsg := assertDocs([]map[string]any{
		{"name": "John"},
		{"model": "iPhone", "owner_id": johnDocKey},
		{"model": "MacBook", "owner_id": johnDocKey},
		{"CPU": "A13", "device_id": getDocKeyFromDocMap(map[string]any{"model": "iPhone", "owner_id": johnDocKey})},
		{"CPU": "M2", "device_id": getDocKeyFromDocMap(map[string]any{"model": "MacBook", "owner_id": johnDocKey})},
	}, docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}
