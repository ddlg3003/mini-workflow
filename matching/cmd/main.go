package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/signal"
	"syscall"

	pb "mini-workflow/api"
	grpcadapter "mini-workflow/matching/internal/adapters/grpc"
	redisadapter "mini-workflow/matching/internal/adapters/redis"
	"mini-workflow/matching/internal/service"

	"mini-workflow/config"
	"mini-workflow/pkg/interceptor"
	"mini-workflow/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	zapLog, err := logger.New(cfg.Log)
	if err != nil {
		fmt.Printf("init logger: %v\n", err)
		return
	}
	defer zapLog.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		zapLog.Fatal("redis connection failed", zap.Error(err))
	}
	defer redisClient.Close()

	taskQueue := redisadapter.NewTaskQueue(redisClient)
	svc := service.New(taskQueue, cfg.Matching)
	handler := grpcadapter.NewHandler(svc)

	lis, err := net.Listen("tcp", cfg.Matching.ListenAddr())
	if err != nil {
		zapLog.Fatal("failed to listen", zap.Error(err))
	}

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.UnaryServerLogger(zapLog)),
	)
	pb.RegisterMatchingServiceServer(srv, handler)

	go func() {
		<-ctx.Done()
		zapLog.Info("shutting down MatchingService")
		srv.GracefulStop()
	}()

	zapLog.Info("MatchingService starting", zap.String("addr", cfg.Matching.ListenAddr()))
	if err := srv.Serve(lis); err != nil {
		zapLog.Fatal("failed to serve", zap.Error(err))
	}
}
