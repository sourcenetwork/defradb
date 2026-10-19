// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package test_acp_dac

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// gatedSubscriptionPolicy is a policy whose `users` resource gates the test collection.
// The owner relation is implicit (granted to the document creator); read access is
// extended to any actor holding the `reader` relation.
const gatedSubscriptionPolicy = `
description: a test policy which marks a collection in a database as a resource
name: test
resources:
- name: users
  permissions:
  - name: delete
  - name: read
    expr: reader
  - name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: reader
    types:
    - actor
`

// TestACP_Subscription_AuthenticatedOwner_ReceivesEvents proves that a subscriber
// who authenticates as the document owner on a policy-gated collection receives
// the create event through the subscription channel.
//
// This exercises the full identity-to-subscription path: the identity supplied on
// the subscribe call is preserved into the per-event re-evaluation, so the ACP
// layer permits the owned document to flow back to the subscriber.
func TestACP_Subscription_AuthenticatedOwner_ReceivesEvents(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   gatedSubscriptionPolicy,
			},
			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},
			// Subscribe as identity 1 (the actor who will own the document).
			&action.SubscriptionRequest{
				Identity: testUtils.ClientIdentity(1),
				Request: `subscription {
					Users {
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"Users": []map[string]any{
							{
								"name": "Shahzad",
								"age":  int64(28),
							},
						},
					},
				},
			},
			// Owner creates a private document, which triggers the subscription event.
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestACP_Subscription_AuthenticatedReader_ReceivesEvents proves that a subscriber
// who is not the owner but has been granted the `reader` relation on the document
// receives the document's events through the subscription channel.
func TestACP_Subscription_AuthenticatedReader_ReceivesEvents(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   gatedSubscriptionPolicy,
			},
			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},
			// Owner creates the document first so it exists to be granted on.
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},
			// Owner grants identity 2 the reader relation on that document.
			testUtils.AddDACActorRelationship{
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
			},
			// Subscribe as the granted reader (identity 2).
			&action.SubscriptionRequest{
				Identity: testUtils.ClientIdentity(2),
				Request: `subscription {
					Users {
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"Users": []map[string]any{
							{
								"name": "Shahzad",
								"age":  int64(29),
							},
						},
					},
				},
			},
			// Owner updates the document, triggering an event the reader should see.
			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `
					{
						"age": 29
					}
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestACP_Subscription_AnonymousSubscriber_ReceivesNothing proves that a subscriber
// with no identity on a policy-gated collection receives no events for private
// documents. This is the correct ACP behavior (not an error, just an empty stream)
// and is the scenario an SSE client without an Authorization header hits.
func TestACP_Subscription_AnonymousSubscriber_ReceivesNothing(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   gatedSubscriptionPolicy,
			},
			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},
			// Subscribe with no identity (anonymous).
			&action.SubscriptionRequest{
				Request: `subscription {
					Users {
						name
						age
					}
				}`,
				// No results expected: the gated document must not reach an
				// anonymous subscriber.
				Results: []map[string]any{},
			},
			// Owner creates a private document. The anonymous subscriber must not
			// see it.
			&action.AddDoc{
				CollectionID: 0,
				Identity:     testUtils.ClientIdentity(1),
				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
