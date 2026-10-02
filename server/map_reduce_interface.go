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

type Host struct {
	info    *api.HostInfo
	connect bool
}

func NewHost(info *api.HostInfo) (Host){
	return Host{info, false}

}

func NewHosts(infos []api.HostInfo) ([]Host){
	hosts := make(Host, len(infos))
	for i, info := range infos{
		hosts[i].info = info
		hosts[i].connect = false
	}
	return hosts
}

var (
	host []api.HostInfo
)

// ApiServiceServer implements api.ApiServiceServer.
type MapServiceServer struct {
	api.UnimplementedMapServiceServer
}

type badConnectionError struct {
	msg string
}

func (e *badConnectionError) Error() string {
	return e.msg
}

// CreatePost hashes the request content, stores the post in global memory and
// returns the hash together with the status of the operation.
func (server *MapServiceServer) Init(ctx context.Context, req *api.InitRequest) (*api.InitResponse, error) {
	host = req.GetHost() // correct here
	if try_to_connect() {
		return &api.InitResponse{
			Status: api.StatusResponse_Ok,
		}, nil
	} else {
		return &api.InitResponse{
			Status: api.StatusResponse_Error,
		}, badConnectionError{"Wro"}
	}
}

func Slice_Tasks() *api.TaskResponse {

}
