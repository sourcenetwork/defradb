// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package issues

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// This test documents: https://github.com/sourcenetwork/defradb/issues/4331
func TestQuerySchemaWithCyclicMutuallyReferentialRelations_Fails(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Player {
						name: String
						decks: [Deck] @relation(name: "deck_player")
						game: Game @relation(name: "game_players")
					}

					type Deck {
						name: String
						owner: Player @relation(name: "deck_player")
						game: Game @relation(name: "game_decks")
					}

					type Game {
						players: [Player] @relation(name: "game_players")
						decks: [Deck] @relation(name: "game_decks")
						winner: Player @relation(name: "game_winner")
					}
                `,
				ExpectedError: "no type found for given name",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
