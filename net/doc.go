/*
Package net provides p2p network functions for the core DefraDB
instance.

Notable design descision. All DocKeys (Documents) have their own
respective PubSub topics.

@todo: Needs review/scrutiny.

Its structured as follows.

We define a Peer object, which encapsulates an instanciated DB
objects, libp2p host object, libp2p DAGService.
 - Peer is responsible for storing all network related meta-data,
   maintaining open connections, pubsub mechanics, etc.

 - Peer object also contains a Server instance

type Peer struct {
	config

	DAGService
	libp2pHost

	db client.DB

	context???
}

Server object is responsible for all underlying gRPC related
functions and as it relates to the pubsub network.

Credit: Some of the base structure of this net package and its
types is inspired/inherited from Textile Threads
(github.com/textileio/go-threads). As such, we are omitting
copyright on this "net" package and will release this folder
under the Apache 2.0 license as per the header of each file.
*/

package net
