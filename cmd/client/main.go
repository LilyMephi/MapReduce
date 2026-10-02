// Command client is a small gRPC client for ApiService.
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"logcounter/grpc/api"
)

func main() {
	addr := flag.String("addr", "localhost:50051", "gRPC server address")
	title := flag.String("title", "", "post title")
	author := flag.String("author", "", "post author id")
	content := flag.String("content", "", "post content")
	timeout := flag.Duration("timeout", 5*time.Second, "request timeout")
	flag.Parse()

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("client: dial %s: %v", *addr, err)
	}
	defer conn.Close()

	client := api.NewApiServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	resp, err := client.CreatePost(ctx, &api.CreatePostRequest{
		Title:    *title,
		AuthorId: *author,
		Content:  *content,
	})
	if err != nil {
		log.Fatalf("client: CreatePost: %v", err)
	}

	log.Printf("CreatePost: id=%d log=%s", resp.GetId(), resp.GetLog())

	got, err := client.GetPost(ctx, &api.GetPostRequest{Id: resp.GetId()})
	if err != nil {
		log.Fatalf("client: GetPost: %v", err)
	}

	log.Printf("GetPost: id=%d title=%q author=%q content=%q log=%s",
		resp.GetId(), got.GetTitle(), got.GetAuthorId(), got.GetContext(), got.GetLog())
}
