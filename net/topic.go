// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package net

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcenetwork/corelog"
	rpc "github.com/sourcenetwork/go-libp2p-pubsub-rpc"

	"github.com/sourcenetwork/defradb/errors"
)

// pubsubTopic is a wrapper of rpc.Topic to be able to track if the topic has
// been subscribed to.
type pubsubTopic struct {
	*rpc.Topic
	subscribed bool
}

// addPubSubTopic subscribes to a topic on the pubsub network
// A custom message handler can be provided to handle incoming messages. If not provided,
// the default message handler will be used.
func (p *Peer) addPubSubTopic(topic string, subscribe bool, handler rpc.MessageHandler) (pubsubTopic, error) {
	if p.ps == nil {
		return pubsubTopic{}, nil
	}

	if handler == nil {
		return pubsubTopic{}, fmt.Errorf("handler cannot be nil")
	}

	log.InfoContext(p.ctx, "Adding pubsub topic",
		corelog.String("PeerID", p.host.ID().String()),
		corelog.String("Topic", topic))

	p.topicMu.Lock()
	defer p.topicMu.Unlock()
	if t, ok := p.topics[topic]; ok {
		// When the topic was previously set to publish only and we now want to subscribe,
		// we need to close the existing topic and create a new one.
		if !t.subscribed && subscribe {
			if err := t.Close(); err != nil {
				return pubsubTopic{}, err
			}
		} else {
			return t, nil
		}
	}

	t, err := rpc.NewTopic(p.ctx, p.ps, p.host.ID(), topic, subscribe)
	if err != nil {
		return pubsubTopic{}, err
	}

	t.SetMessageHandler(handler)
	pst := pubsubTopic{
		Topic:      t,
		subscribed: subscribe,
	}
	p.topics[topic] = pst
	return pst, nil
}

// removePubSubTopic unsubscribes to a topic
func (p *Peer) removePubSubTopic(topic string) error {
	if p.ps == nil {
		return nil
	}

	log.Info("Removing pubsub topic",
		corelog.String("PeerID", p.host.ID().String()),
		corelog.String("Topic", topic))

	p.topicMu.Lock()
	defer p.topicMu.Unlock()
	if t, ok := p.topics[topic]; ok {
		delete(p.topics, topic)
		return t.Close()
	}
	return nil
}

func (p *Peer) removeAllPubsubTopics() error {
	if p.ps == nil {
		return nil
	}

	log.Info("Removing all pubsub topics",
		corelog.String("PeerID", p.host.ID().String()))

	p.topicMu.Lock()
	defer p.topicMu.Unlock()
	for id, t := range p.topics {
		delete(p.topics, id)
		if err := t.Close(); err != nil {
			return err
		}
	}
	return nil
}

// publishDirectToTopic temporarily joins a pubsub topic to publish data and immediately closes it.
//
// This is useful to publish messages without incurring the cost of a full pubsub rpc topic.
func (p *Peer) publishDirectToTopic(ctx context.Context, topic string, data []byte, isRetry bool) error {
	psTopic, err := p.ps.Join(topic)
	if err != nil {
		if strings.Contains(err.Error(), "topic already exists") && !isRetry {
			// Reaching this is really rare and probably only possible
			// through from tests. We can handle this by simply trying again a single time.
			return p.publishDirectToTopic(ctx, topic, data, true)
		}
		return NewErrPushLog(err, errors.NewKV("Topic", topic))
	}
	err = psTopic.Publish(ctx, data)
	if err != nil {
		return NewErrPushLog(err, errors.NewKV("Topic", topic))
	}
	return psTopic.Close()
}
