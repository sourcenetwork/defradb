# int64 support in JSON

Previously, the JSON interface inside client/json.go only had a singular numeric type: float64. It now supports int64, but this did break some existing test syntax. They had to be adjusted.