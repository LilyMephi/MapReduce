package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	serverpb "mapreduce/proto"
)

type taskServer struct {
	serverpb.UnimplementedTaskServiceServer // forward-compat: требуется новыми версиями grpc-go
}

// ---------- 1. Unary ----------

func (s *taskServer) Process(ctx context.Context, req *serverpb.TaskRequest) (*serverpb.TaskResponse, error) {
	log.Printf("[unary] req=%s op=%s len=%d", req.GetRequestId(), req.GetOperation(), len(req.GetPayload()))

	resp := runTask(req)
	if !resp.GetOk() {
		// Прикладная ошибка → INVALID_ARGUMENT + возвращаем resp с деталями
		return resp, status.Errorf(codes.InvalidArgument, "%s", resp.GetError())
	}
	return resp, nil
}

// ---------- 2. Server streaming ----------

func (s *taskServer) ProcessStream(req *serverpb.TaskRequest, stream serverpb.TaskService_ProcessStreamServer) error {
	log.Printf("[server-stream] req=%s op=%s", req.GetRequestId(), req.GetOperation())

	resp := runTask(req)
	if !resp.GetOk() {
		return status.Errorf(codes.InvalidArgument, "%s", resp.GetError())
	}

	// Отдаём результат чанками по 16 байт — эмулируем «большой ответ».
	const chunk = 16
	data := resp.GetResult()
	for i := 0; i < len(data); i += chunk {
		end := i + chunk
		if end > len(data) {
			end = len(data)
		}
		part := &serverpb.TaskResponse{
			RequestId: resp.GetRequestId(),
			Ok:        true,
			Result:    data[i:end],
		}
		if err := stream.Send(part); err != nil {
			return status.Errorf(codes.Internal, "send chunk: %v", err)
		}
		// Уважаем отмену клиента
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		default:
		}
	}
	return nil
}

// ---------- 3. Bidirectional streaming ----------

func (s *taskServer) ProcessBatch(stream serverpb.TaskService_ProcessBatchServer) error {
	log.Println("[bidi] client connected")

	// Пул воркеров — обрабатываем запросы конкурентно и шлём ответы.
	// Но gRPC требует, чтобы Send вызывался из одной горутины — поэтому
	// собираем ответы через канал.
	type result struct{ resp *serverpb.TaskResponse }

	reqCh := make(chan *serverpb.TaskRequest, 16)
	resCh := make(chan *serverpb.TaskResponse, 16)

	// Reader: читает из стрима в канал
	go func() {
		defer close(reqCh)
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Printf("[bidi] recv err: %v", err)
				return
			}
			reqCh <- req
		}
	}()

	// Worker pool
	const workers = 4
	done := make(chan struct{}, workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for req := range reqCh {
				resCh <- runTask(req)
			}
		}()
	}
	go func() {
		for i := 0; i < workers; i++ {
			<-done
		}
		close(resCh)
	}()

	// Writer: единственная горутина, которая пишет в stream
	for resp := range resCh {
		if err := stream.Send(resp); err != nil {
			return status.Errorf(codes.Internal, "send: %v", err)
		}
	}
	log.Println("[bidi] done")
	return nil
}

// ---------- main ----------

func main() {
	addr := "0.0.0.0:50051"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		// Лимиты — практические значения по умолчанию:
		grpc.MaxRecvMsgSize(16<<20), // 16 MiB
		grpc.MaxSendMsgSize(16<<20),
		grpc.ConnectionTimeout(10*time.Second),
	)
	serverpb.RegisterTaskServiceServer(grpcServer, &taskServer{})

	// Graceful shutdown по SIGINT/SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		log.Println("shutting down...")
		grpcServer.GracefulStop()
	}()

	log.Printf("task gRPC server listening on %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
	log.Println("bye")
}

// заглушка, чтобы strings не ругался в некоторых конфигурациях
var _ = strings.TrimSpace
var _ = fmt.Sprintf
