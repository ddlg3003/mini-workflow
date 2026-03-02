package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/signal"
	"syscall"

	pb "mini-workflow/api"
	grpcadapter "mini-workflow/history/internal/adapters/grpc"
	pgadapter "mini-workflow/history/internal/adapters/postgres"
	redisadapter "mini-workflow/history/internal/adapters/redis"
	"mini-workflow/history/internal/service"

	"mini-workflow/config"
	"mini-workflow/pkg/interceptor"
	"mini-workflow/pkg/logger"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
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

	db, err := sqlx.Open("postgres", cfg.Database.DSN())
	if err != nil {
		zapLog.Fatal("open database", zap.Error(err))
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		zapLog.Fatal("ping database", zap.Error(err))
	}

	redisClient := redisadapter.NewRedisClient(cfg.Redis.Addr)
	defer redisClient.Close()

	matchingClient, err := grpcadapter.NewMatchingClient(cfg.Matching)
	if err != nil {
		zapLog.Fatal("create matching client", zap.Error(err))
	}

	repo := pgadapter.NewExecutionRepository(db)
	timerStore := redisadapter.NewTimerStore(redisClient)

	svc := service.New(repo, matchingClient, timerStore, zapLog)
	tp := service.NewTimerProcessor(repo, timerStore, matchingClient, zapLog)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go tp.Run(ctx)

	lis, err := net.Listen("tcp", cfg.History.ListenAddr())
	if err != nil {
		zapLog.Fatal("failed to listen", zap.Error(err))
	}

	handler := grpcadapter.NewHandler(svc)
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.UnaryServerLogger(zapLog)),
	)
	pb.RegisterHistoryServiceServer(srv, handler)

	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()

	zapLog.Info("HistoryService starting", zap.String("addr", cfg.History.ListenAddr()))
	if err := srv.Serve(lis); err != nil {
		zapLog.Fatal("failed to serve", zap.Error(err))
	}
}
