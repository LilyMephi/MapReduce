package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	serverpb "mapreduce/proto"
)

func main() {
	conn, err := grpc.Dial("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	cli := serverpb.NewTaskServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Unary
	resp, err := cli.Process(ctx, &serverpb.TaskRequest{
		RequestId: "req-1",
		Operation: serverpb.Operation_OP_UPPER,
		Payload:   "hello grpc",
	})
	if err != nil {
		log.Fatalf("unary: %v", err)
	}
	fmt.Printf("unary → ok=%v result=%q ns=%d\n", resp.GetOk(), resp.GetResult(), resp.GetProcessingNs())

	// 2. Server streaming
	stream, err := cli.ProcessStream(ctx, &serverpb.TaskRequest{
		RequestId: "req-2",
		Operation: serverpb.Operation_OP_BASE64,
		Payload:   "streaming payload for chunks",
	})
	if err != nil {
		log.Fatalf("stream: %v", err)
	}
	fmt.Print("server-stream → ")
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("recv: %v", err)
		}
		fmt.Print(chunk.GetResult())
	}
	fmt.Println()

	// 3. Bidi streaming
	bidi, err := cli.ProcessBatch(ctx)
	if err != nil {
		log.Fatalf("bidi open: %v", err)
	}
	go func() {
		for i := 0; i < 5; i++ {
			_ = bidi.Send(&serverpb.TaskRequest{
				RequestId: fmt.Sprintf("b-%d", i),
				Operation: serverpb.Operation_OP_COUNT_WORDS,
				Payload:   fmt.Sprintf("hello world part %d", i),
			})
		}
		_ = bidi.CloseSend()
	}()
	for {
		r, err := bidi.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("bidi recv: %v", err)
		}
		fmt.Printf("bidi → %s ok=%v result=%s\n", r.GetRequestId(), r.GetOk(), r.GetResult())
	}
}
