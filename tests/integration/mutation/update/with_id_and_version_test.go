package update

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_WithIdAndVersion_ReturnResults(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						points: Float
					}
				`,
			},
			testUtils.CreateDoc{
				// bae-9466cfe3-c011-5d44-b1cd-f0c5a46d9202
				Doc: `{
					"name": "John",
					"points": 42.1
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Bob",
					"points": 66.6
				}`,
			},
			testUtils.Request{
				Request: `mutation {
					update_Users(docID: "bae-9466cfe3-c011-5d44-b1cd-f0c5a46d9202", input: {points: 59}) {
						name
						_version {
							cid
						}
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{
						{
							"name": "John",
							"_version": []map[string]any{
								{
									"cid": "bafyreickdavrgglg4cksxq74vvkg2tiuu56d6rltpavlfzzhnv73bmstxa",
								},
								{
									"cid": "bafyreib52mddt7kbaj3ewdzbvktcq5jkzmvj24op5eapjhhm5stzxkaxb4",
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
