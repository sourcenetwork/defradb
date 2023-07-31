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

/*
Package net provides P2P network functions for the core DefraDB instance.

Notable design descision: all DocKeys (Documents) have their own respective PubSub topics.

The Peer object encapsulates an instanciated DB objects, libp2p host object, libp2p DAGService.
Peer is responsible for storing all network related meta-data, maintaining open connections, pubsub mechanics, etc.
The Peer object also contains a Server instance.

The Server object is responsible for all underlying gRPC related functions and as it relates to the pubsub network.

Credit: Some of the base structure of this net package and its types is inspired/inherited from
Textile Threads (github.com/textileio/go-threads). As such, we are omitting copyright on this "net" package
and will release this folder under the Apache 2.0 license as per the header of each file.

@todo: Needs review/scrutiny.
*/
package net
