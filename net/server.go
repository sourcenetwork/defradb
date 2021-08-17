package net

// Server is the request/response instance for all P2P RPC communication.
// Implements gRPC server. See net/pb/net.proto for corresponding service definitions.
//
// Specifically, server handles the push/get request/response aspects of the RPC service
// but not the API calls.
type server struct {
}

// GetDocGraph recieves a get graph request
// func (s *server) GetDocGraph( GetDocGraphRequest ) (GetDocGraphResponse, error)

// PushDocGraph recieves a push graph request
// func (s *server) PushDocGraph( PushDocGraphRequest ) (PushDocGraphResponse, error)
