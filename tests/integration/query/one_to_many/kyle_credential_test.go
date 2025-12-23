package one_to_many

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

var userCredentialGQLSchema = (`
	type User {
		publicID: String
		isEnabled: Boolean
		credentials: [Credential]
	}

	type Credential {
		credentialID: String
		publicKeySPKI: String
		publicKeyAlgoCOSE: Int
		user: User
	}
`)

func TestQueryOneToMany_WithFilterOnParentAndChild_ReturnsExactlyOne(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: userCredentialGQLSchema,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-a312e5b0-1183-51e1-9579-6637942af8e4
				Doc: `{
					"publicID": "nUV6e3sbciL+KeWsdVEEsbuP",
					"isEnabled": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"credentialID": "r2fdP6Ydwds3s2VwTEfQeGgN+2oLFCa/DMWuo5ka+gc=",
					"publicKeySPKI": "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE+NAT861Mk+hKJodOV/kfDY+pdO2yDZJncJgZlTT2ZYK9AcVWDhsWL9oF/2EY/za8NL/deVEZ1Ylz4BpRi5Khcg==",
					"publicKeyAlgoCOSE": -7,
					"user_id": "bae-a312e5b0-1183-51e1-9579-6637942af8e4"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"credentialID": "AOWrPXX4jS504tsAT6/zEA==",
					"publicKeyAlgoCOSE": -7,
					"publicKeySPKI": "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEdY/y1sdUcDDiTvhEeigSIUbBtYTHpwUPVVDl9/Vs71bIyOdDfWoeduMM3EkCYDLoy1WAiA+oKwLe6hW53f3nNQ==",
					"user_id": "bae-a312e5b0-1183-51e1-9579-6637942af8e4"
				}`,
			}, testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"credentialID": "ATSd8SK6bIAggBdG1KEA1OButQrxjEidN1ijnfUjlxkPKLZPzUtZsuA3XlzrD32w70RaOKlRWwDZ/chy9/PmmbY=",
					"publicKeyAlgoCOSE": -7,
					"publicKeySPKI": "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE/kcvO18qMlDcfBxzgLzc6V7c7g7Q1rOGVxxFciedgSU7cXdKI8hY986cKDfetQI3wFgOz3wLSWKbOvuM3S4cdA==",
					"user_id": "bae-a312e5b0-1183-51e1-9579-6637942af8e4"
				}`,
			},
			testUtils.Request{
				Request: `
				query GetPasskeyCredential($credentialID:String,$userIsEnabled:Boolean,$userID:String) {
        			PasskeyCredential: Credential(
						filter:{
						credentialID:{_eq:$credentialID},
						user:{isEnabled:{_eq:$userIsEnabled},publicID:{_eq:$userID}}}
					) {
                		publicKeySPKI publicKeyAlgoCOSE credentialID user { publicID }
        			}
				}
				`,
				Variables: immutable.Some(map[string]any{
					"credentialID":  "AOWrPXX4jS504tsAT6/zEA==",
					"userIsEnabled": true,
					"userID":        "nUV6e3sbciL+KeWsdVEEsbuP",
				}),
				Results: map[string]any{
					"PasskeyCredential": []map[string]any{
						{
							"credentialID":      "AOWrPXX4jS504tsAT6/zEA==",
							"publicKeyAlgoCOSE": int64(-7),
							"publicKeySPKI":     "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEdY/y1sdUcDDiTvhEeigSIUbBtYTHpwUPVVDl9/Vs71bIyOdDfWoeduMM3EkCYDLoy1WAiA+oKwLe6hW53f3nNQ==",
							"user": map[string]any{
								"publicID": "nUV6e3sbciL+KeWsdVEEsbuP",
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
