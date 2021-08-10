/*
Package net provides p2p network functions for the core DefraDB instance.

Its structured as follows.

We define a Peer object, which encapsulates an instanciated DB objects, libp2p host object, libp2p DAGService.
 - Peer is responsible for storing all network related meta-data, maintaining open connections, pubsub mechanics, etc.
 - Peer object also contains a Server instance

type Peer struct {
	config

	DAGService
	libp2pHost

	db client.DB

	context???
}

Server object

*/
package net
