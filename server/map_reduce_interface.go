package main

import (
	"context"
	"hash/fnv"
	"log"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"logcounter/grpc/api"
)

// Posts is a single stored post.
type Posts struct {
	author  string
	title   string
	content string
	id      uint
}

// posts is the global in-memory store keyed by the content hash.
var (
	fileMu sync.RWMutex
	file   = make(map[uint32]Posts)
)


var (
	host []api.HostInfo
)

// ApiServiceServer implements api.ApiServiceServer.
type MapServiceServer struct {
	api.UnimplementedMapServiceServer
}

// CreatePost hashes the request content, stores the post in global memory and
// returns the hash together with the status of the operation.
func (server *MapServiceServer) Init(ctx context.Context, req *api.InitRequest) (*api.InitResponse, error) {
	host := req.GetHost()


	return &api.InitResponse{
		Status: api.StatusResponse_Ok,
	}, nil
}

func Slice_Tasks() (*api.TaskResponse) {

}
