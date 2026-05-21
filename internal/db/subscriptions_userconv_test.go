// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Join collection with two foreign-key relations.
const userConversationSchema = `
type User {
	displayName: String
}

type Conversation {
	name: String
}

type UserConversation {
	joinedAt: String
	user: User @primary
	conversation: Conversation @primary
}
`

// A local write to a two-FK join collection while a subscription on
// that collection is open must deliver the event to the subscriber.
func TestSubscription_UserConversationLocalWrite_DeliversEvent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, userConversationSchema)
	require.NoError(t, err)

	userRes := db.ExecRequest(ctx, `mutation {
		add_User(input: {displayName: "Alice"}) {
			_docID
		}
	}`)
	require.Empty(t, userRes.GQL.Errors)

	convRes := db.ExecRequest(ctx, `mutation {
		add_Conversation(input: {name: "alice-bob"}) {
			_docID
		}
	}`)
	require.Empty(t, convRes.GQL.Errors)

	userDocID := userRes.GQL.Data.(map[string]any)["add_User"].([]map[string]any)[0]["_docID"].(string)
	convDocID := convRes.GQL.Data.(map[string]any)["add_Conversation"].([]map[string]any)[0]["_docID"].(string)

	sub := db.ExecRequest(ctx, `subscription {
		UserConversation {
			_docID
			_userID
			_conversationID
		}
	}`)
	require.Empty(t, sub.GQL.Errors)
	require.NotNil(t, sub.Subscription)

	joinRes := db.ExecRequest(ctx, `mutation {
		add_UserConversation(input: {
			joinedAt: "2026-05-18T00:00:00Z",
			user: "`+userDocID+`",
			conversation: "`+convDocID+`"
		}) {
			_docID
		}
	}`)
	require.Empty(t, joinRes.GQL.Errors)

	select {
	case result, ok := <-sub.Subscription:
		require.True(t, ok, "subscription stream must stay open after the write")
		require.NotNil(t, result.Data, "event for the just-written doc must be delivered")
	case <-time.After(1 * time.Second):
		t.Fatal("no event delivered within 1s")
	}
}
