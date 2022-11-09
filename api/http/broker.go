// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

// import (
// 	"context"

// 	"github.com/sourcenetwork/defradb/client"
// 	"github.com/sourcenetwork/defradb/logging"
// )

// type broker struct {
// 	notifier    chan client.GQLResult
// 	subscribe   chan chan client.GQLResult
// 	unsubscribe chan chan client.GQLResult
// }

// func newBroker() *broker {
// 	return &broker{
// 		notifier:    make(chan client.GQLResult, 1),
// 		subscribe:   make(chan chan client.GQLResult),
// 		unsubscribe: make(chan chan client.GQLResult),
// 	}
// }

// func (b *broker) listen(ctx context.Context) {
// 	clients := make(map[chan client.GQLResult]struct{})

// 	for {
// 		select {
// 		case subCh := <-b.subscribe:
// 			clients[subCh] = struct{}{}
// 			log.Info(ctx, "GraphQL client added to broker", logging.NewKV("clients", len(clients)))
// 		case unsubCh := <-b.unsubscribe:
// 			delete(clients, unsubCh)
// 			unsubCh = nil
// 			log.Info(ctx, "GraphQL client removed from broker", logging.NewKV("clients", len(clients)))
// 		case msg := <-b.notifier:
// 			for sub := range clients {
// 				// To protect against unresponsive clients, we use a non-blocking send.
// 				select {
// 				case sub <- msg:
// 				default:
// 				}
// 			}
// 		}
// 	}
// }
