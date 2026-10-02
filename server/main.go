package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	"logcounter/grpc/api"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	srv := grpc.NewServer()
	api.RegisterApiServiceServer(srv, &ApiServiceServer{})

	log.Printf("gRPC server listening on %s", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
