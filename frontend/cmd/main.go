package main

import (
	"log"
	"net"
	"time"

	pb "mini-workflow/api"
	grpcadapter "mini-workflow/frontend/internal/adapters/grpc"
	"mini-workflow/frontend/internal/router"
	"mini-workflow/frontend/internal/service"

	"mini-workflow/config"
	"mini-workflow/pkg/interceptor"
	"mini-workflow/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	zapLog, err := logger.New(cfg.Log)
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer zapLog.Sync()

	historyClient, err := grpcadapter.NewHistoryClient(cfg.History)
	if err != nil {
		zapLog.Fatal("connect to history", zap.Error(err))
	}

	matchingClient, err := grpcadapter.NewMatchingClient(cfg.Matching)
	if err != nil {
		zapLog.Fatal("connect to matching", zap.Error(err))
	}

	r := router.New(historyClient)
	svc := service.New(r, matchingClient, zapLog)
	handler := grpcadapter.NewHandler(svc)

	srv := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: cfg.Matching.PollTimeout() + 5*time.Second,
			Time:              cfg.Matching.PollTimeout(),
			Timeout:           5 * time.Second,
		}),
		grpc.UnaryInterceptor(interceptor.UnaryServerLogger(zapLog)),
	)
	pb.RegisterFrontendServiceServer(srv, handler)

	lis, err := net.Listen("tcp", cfg.Frontend.ListenAddr())
	if err != nil {
		zapLog.Fatal("listen", zap.Error(err))
	}

	zapLog.Info("FrontendService starting", zap.String("addr", cfg.Frontend.ListenAddr()))
	if err := srv.Serve(lis); err != nil {
		zapLog.Fatal("serve", zap.Error(err))
	}
}
