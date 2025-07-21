package main

import (
	"context"
	"errors"
	"net"
	"sync"

	authHandler "github.com/Snake1-1eyes/vk_task_marketplace/internal/auth/delivery/grpc"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/bootstrap"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/config"
	listingHandler "github.com/Snake1-1eyes/vk_task_marketplace/internal/listing/delivery/grpc"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/logger"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/middleware"
	auth_pb "github.com/Snake1-1eyes/vk_task_marketplace/pkg/api/auth"
	listings_pb "github.com/Snake1-1eyes/vk_task_marketplace/pkg/api/listings"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// StartGRPCServer запускает gRPC сервер
func StartGRPCServer(ctx context.Context, cfg *config.Config, services *bootstrap.Services, appLogger *logger.Logger) error {
	keepAliveParams := keepalive.ServerParameters{
		MaxConnectionAge: cfg.GRPC.MaxConnAge,
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.ValidationUnaryInterceptor(appLogger),
			middleware.UnaryLoggerInterceptor(appLogger),
			middleware.AuthInterceptor(services.AuthUseCase, appLogger),
		),
		grpc.KeepaliveParams(keepAliveParams),
	)

	authServer := authHandler.New(services.AuthUseCase, appLogger)
	auth_pb.RegisterAuthServiceServer(grpcServer, authServer)

	listingServer := listingHandler.New(services.ListingsUseCase, appLogger)
	listings_pb.RegisterListingsServiceServer(grpcServer, listingServer)

	reflection.Register(grpcServer)

	addr := cfg.GRPC.Host + ":" + cfg.GRPC.Port
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		appLogger.Error(ctx, "Не удалось открыть TCP-порт для gRPC",
			zap.String("address", addr),
			zap.Error(err))
		return err
	}

	appLogger.Info(ctx, "gRPC сервер запущен",
		zap.String("address", addr),
		zap.String("environment", cfg.Environment))

	serverErrors := make(chan error, 1)
	var wg sync.WaitGroup

	go func() {
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			appLogger.Error(ctx, "Ошибка работы gRPC сервера", zap.Error(err))
			serverErrors <- err
		}
		close(serverErrors)
	}()

	go func() {
		<-ctx.Done()
		appLogger.Info(ctx, "Начинаем graceful shutdown gRPC сервера...")
		wg.Add(1)
		defer wg.Done()
		grpcServer.GracefulStop()
		appLogger.Info(ctx, "gRPC сервер остановлен")
	}()

	select {
	case err := <-serverErrors:
		return err
	case <-ctx.Done():
		wg.Wait()
		return nil
	}
}
