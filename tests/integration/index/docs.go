// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import "github.com/sourcenetwork/defradb/tests/predefined"

func makeExplainQuery(req string) string {
	return "query @explain(type: execute) " + req[6:]
}

func getUserDocs() predefined.DocsList {
	return predefined.DocsList{
		ColName: "User",
		Docs: []map[string]any{
			{
				"name":     "Shahzad",
				"age":      20,
				"verified": false,
				"email":    "shahzad@gmail.com",
				"devices": []map[string]any{
					{
						"model": "iPhone Xs",
						"year":  2022,
						"type":  "phone",
						"specs": map[string]any{
							"CPU":     2.2,
							"Chip":    "Intel i3",
							"RAM":     8,
							"Storage": 512,
							"OS":      "iOS 12",
						},
					},
					{
						"model": "MacBook Pro",
						"year":  2020,
						"type":  "laptop",
						"specs": map[string]any{
							"CPU":     2.4,
							"Chip":    "Intel i5",
							"RAM":     16,
							"Storage": 2048,
							"OS":      "Yosemite",
						},
					},
				},
				"address": map[string]any{
					"postalCode": 4635,
					"city":       "Montreal",
					"country":    "Canada",
					"street":     "Queen Mary Rd",
				},
			},
			{
				"name":     "Bruno",
				"age":      23,
				"verified": true,
				"email":    "bruno@gmail.com",
				"devices":  []map[string]any{},
				"address": map[string]any{
					"postalCode": 10001,
					"city":       "New York",
					"country":    "USA",
					"street":     "5th Ave",
				},
			},
			{
				"name":     "Roy",
				"age":      44,
				"verified": true,
				"email":    "roy@gmail.com",
				"devices":  []map[string]any{},
				"address": map[string]any{
					"postalCode": 90028,
					"city":       "Los Angeles",
					"country":    "USA",
					"street":     "Hollywood Blvd",
				},
			},
			{
				"name":     "Fred",
				"age":      28,
				"verified": false,
				"email":    "fred@gmail.com",
				"devices": []map[string]any{
					{
						"model": "Samsung Galaxy S20",
						"year":  2022,
						"type":  "phone",
						"specs": map[string]any{
							"CPU":     2.0,
							"Chip":    "AMD Athlon",
							"RAM":     8,
							"Storage": 256,
							"OS":      "Android 11",
						},
					},
					{
						"model": "Lenovo ThinkPad",
						"year":  2020,
						"type":  "laptop",
						"specs": map[string]any{
							"CPU":     1.9,
							"Chip":    "AMD Ryzen",
							"RAM":     8,
							"Storage": 1024,
							"OS":      "Windows 10",
						},
					},
				},
				"address": map[string]any{
					"postalCode": 6512,
					"city":       "Montreal",
					"country":    "Canada",
					"street":     "Park Ave",
				},
			},
			{
				"name":     "John",
				"age":      30,
				"verified": false,
				"email":    "john@gmail.com",
				"devices": []map[string]any{
					{
						"model": "Google Pixel 5",
						"year":  2022,
						"type":  "phone",
						"specs": map[string]any{
							"CPU":     2.4,
							"Chip":    "Octa-core",
							"RAM":     16,
							"Storage": 512,
							"OS":      "Android 11",
						},
					},
					{
						"model": "Asus Vivobook",
						"year":  2022,
						"type":  "laptop",
						"specs": map[string]any{
							"CPU":     2.9,
							"Chip":    "Intel i7",
							"RAM":     64,
							"Storage": 2048,
							"OS":      "Windows 10",
						},
					},
					{
						"model": "Commodore 64",
						"year":  1982,
						"type":  "computer",
						"specs": map[string]any{
							"CPU":     0.1,
							"Chip":    "MOS 6510",
							"RAM":     1,
							"Storage": 1,
							"OS":      "Commodore BASIC 2.0",
						},
					},
				},
				"address": map[string]any{
					"postalCode": 690,
					"city":       "Montreal",
					"country":    "Canada",
					"street":     "Notre-Dame St W",
				},
			},
			{
				"name":     "Islam",
				"age":      32,
				"verified": false,
				"email":    "islam@gmail.com",
				"devices": []map[string]any{
					{
						"model": "iPhone 12s",
						"year":  2018,
						"type":  "phone",
						"specs": map[string]any{
							"CPU":     2.1,
							"Chip":    "A11 Bionic",
							"RAM":     8,
							"Storage": 1024,
							"OS":      "iOS 14",
						},
					},
					{
						"model": "MacBook Pro",
						"year":  2023,
						"type":  "laptop",
						"specs": map[string]any{
							"CPU":     2.6,
							"Chip":    "Apple M2 Max",
							"RAM":     32,
							"Storage": 1024,
							"OS":      "Sonoma 14",
						},
					},
					{
						"model": "iPad Pro",
						"year":  2020,
						"type":  "tablet",
						"specs": map[string]any{
							"CPU":     2.1,
							"Chip":    "Intel i5",
							"RAM":     8,
							"Storage": 512,
							"OS":      "iOS 14",
						},
					},
					{
						"model": "Playstation 5",
						"year":  2022,
						"type":  "game_console",
						"specs": map[string]any{
							"CPU":     3.5,
							"Chip":    "AMD Zen 2",
							"RAM":     16,
							"Storage": 825,
							"OS":      "FreeBSD",
						},
					},
					{
						"model": "Nokia 7610",
						"year":  2003,
						"type":  "phone",
						"specs": map[string]any{
							"CPU":     1.8,
							"Chip":    "Cortex A710",
							"RAM":     12,
							"Storage": 2,
							"OS":      "Symbian 7.0",
						},
					},
				},
				"address": map[string]any{
					"postalCode": 80804,
					"city":       "Munich",
					"country":    "Germany",
					"street":     "Leopold Str",
				},
			},
			{
				"name":     "Andy",
				"age":      33,
				"verified": true,
				"email":    "andy@gmail.com",
				"devices": []map[string]any{
					{
						"model": "Xiaomi Phone",
						"year":  2022,
						"type":  "phone",
						"specs": map[string]any{
							"CPU":     1.6,
							"Chip":    "AMD Octen",
							"RAM":     8,
							"Storage": 512,
							"OS":      "Android 11",
						},
					},
					{
						"model": "Alienware x16",
						"year":  2018,
						"type":  "laptop",
						"specs": map[string]any{
							"CPU":     3.2,
							"Chip":    "Intel i7",
							"RAM":     64,
							"Storage": 2048,
							"OS":      "Windows 9",
						},
					},
				},
				"address": map[string]any{
					"postalCode": 101103,
					"city":       "London",
					"country":    "UK",
					"street":     "Baker St",
				},
			},
			{
				"name":     "Addo",
				"age":      42,
				"verified": true,
				"email":    "addo@gmail.com",
				"devices": []map[string]any{
					{
						"model": "iPhone 10",
						"year":  2021,
						"type":  "phone",
						"specs": map[string]any{
							"CPU":     1.8,
							"Chip":    "Intel i3",
							"RAM":     8,
							"Storage": 256,
							"OS":      "iOS 12",
						},
					},
					{
						"model": "Acer Aspire 5",
						"year":  2020,
						"type":  "laptop",
						"specs": map[string]any{
							"CPU":     2.0,
							"Chip":    "Intel i5",
							"RAM":     16,
							"Storage": 512,
							"OS":      "Windows 10",
						},
					},
					{
						"model": "HyperX Headset",
						"year":  2014,
						"type":  "headset",
						"specs": map[string]any{
							"CPU":     "N/A",
							"Chip":    "N/A",
							"RAM":     "N/A",
							"Storage": "N/A",
							"OS":      "N/A",
						},
					},
					{
						"model": "Playstation 5",
						"year":  2021,
						"type":  "game_console",
						"specs": map[string]any{
							"CPU":     3.5,
							"Chip":    "AMD Zen 2",
							"RAM":     16,
							"Storage": 825,
							"OS":      "FreeBSD",
						},
					},
				},
				"address": map[string]any{
					"postalCode": 403,
					"city":       "Ottawa",
					"country":    "Canada",
					"street":     "Bank St",
				},
			},
			{
				"name":     "Keenan",
				"age":      48,
				"verified": true,
				"email":    "keenan@gmail.com",
				"devices": []map[string]any{
					{
						"model": "iPhone 13",
						"year":  2022,
						"type":  "phone",
						"specs": map[string]any{
							"CPU":     2.3,
							"Chip":    "M1",
							"RAM":     8,
							"Storage": 1024,
							"OS":      "iOS 14",
						},
					},
					{
						"model": "MacBook Pro",
						"year":  2017,
						"type":  "laptop",
						"specs": map[string]any{
							"CPU":     2.0,
							"Chip":    "A11 Bionic",
							"RAM":     16,
							"Storage": 512,
							"OS":      "Ventura",
						},
					},
					{
						"model": "iPad Mini",
						"year":  2015,
						"type":  "tablet",
						"specs": map[string]any{
							"CPU":     1.9,
							"Chip":    "Intel i3",
							"RAM":     8,
							"Storage": 1024,
							"OS":      "iOS 12",
						},
					},
				},
				"address": map[string]any{
					"postalCode": 1600,
					"city":       "San Francisco",
					"country":    "USA",
					"street":     "Market St",
				},
			},
			{
				"name":     "Chris",
				"age":      55,
				"verified": true,
				"email":    "chris@gmail.com",
				"devices": []map[string]any{
					{
						"model": "Walkman",
						"year":  2000,
						"type":  "phone",
						"specs": map[string]any{
							"CPU":     1.8,
							"Chip":    "Cortex-A53 ",
							"RAM":     8,
							"Storage": 256,
							"OS":      "Android 11",
						},
					},
				},
				"address": map[string]any{
					"postalCode": 11680,
					"city":       "Toronto",
					"country":    "Canada",
					"street":     "Yonge St",
				},
			},
		},
	}
}
