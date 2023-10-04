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

type docsCollection struct {
	colName string
	docs    []map[string]any
}

func getUserDocs() docsCollection {
	return docsCollection{
		colName: "User",
		docs: []map[string]any{
			{
				"name":     "Shahzad",
				"age":      20,
				"verified": false,
				"email":    "shahzad@gmail.com",
				"devices": docsCollection{
					colName: "Device",
					docs: []map[string]any{
						{
							"model": "iPhone Xs",
							"year":  2022,
							"type":  "phone",
						},
						{
							"model": "MacBook Pro",
							"year":  2020,
							"type":  "laptop",
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
				"name":     "Fred",
				"age":      28,
				"verified": false,
				"email":    "fred@gmail.com",
				"devices": docsCollection{
					colName: "Device",
					docs: []map[string]any{
						{
							"model": "Samsung Galaxy S20",
							"year":  2022,
							"type":  "phone",
						},
						{
							"model": "Lenovo ThinkPad",
							"year":  2020,
							"type":  "laptop",
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
				"devices": docsCollection{
					colName: "Device",
					docs: []map[string]any{
						{
							"model": "Google Pixel 5",
							"year":  2022,
							"type":  "phone",
						},
						{
							"model": "Asus Vivobook",
							"year":  2022,
							"type":  "laptop",
						},
						{
							"model": "Commodore 64",
							"year":  1982,
							"type":  "computer",
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
				"devices": docsCollection{
					colName: "Device",
					docs: []map[string]any{
						{
							"model": "iPhone 12s",
							"year":  2018,
							"type":  "phone",
						},
						{
							"model": "MacBook Pro",
							"year":  2023,
							"type":  "laptop",
						},
						{
							"model": "iPad Pro",
							"year":  2020,
							"type":  "tablet",
						},
						{
							"model": "Playstation 5",
							"year":  2022,
							"type":  "game_console",
						},
						{
							"model": "Nokia 7610",
							"year":  2003,
							"type":  "phone",
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
				"devices": docsCollection{
					colName: "Device",
					docs: []map[string]any{
						{
							"model": "Xiaomi Phone",
							"year":  2022,
							"type":  "phone",
						},
						{
							"model": "Alienware x16",
							"year":  2018,
							"type":  "laptop",
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
				"devices": docsCollection{
					colName: "Device",
					docs: []map[string]any{
						{
							"model": "iPhone 10",
							"year":  2021,
							"type":  "phone",
						},
						{
							"model": "Acer Aspire 5",
							"year":  2020,
							"type":  "laptop",
						},
						{
							"model": "HyperX Headset",
							"year":  2014,
							"type":  "headset",
						},
						{
							"model": "Playstation 5",
							"year":  2021,
							"type":  "game_console",
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
				"devices": docsCollection{
					colName: "Device",
					docs: []map[string]any{
						{
							"model": "iPhone 13",
							"year":  2022,
							"type":  "phone",
						},
						{
							"model": "MacBook Pro",
							"year":  2017,
							"type":  "laptop",
						},
						{
							"model": "iPad Mini",
							"year":  2015,
							"type":  "tablet",
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
				"devices": docsCollection{
					colName: "Device",
					docs: []map[string]any{
						{
							"model": "Walkman",
							"year":  2000,
							"type":  "phone",
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
