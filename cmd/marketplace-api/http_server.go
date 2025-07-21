package main

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Snake1-1eyes/vk_task_marketplace/internal/adapter"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/config"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/logger"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/middleware"
	auth_pb "github.com/Snake1-1eyes/vk_task_marketplace/pkg/api/auth"
	listings_pb "github.com/Snake1-1eyes/vk_task_marketplace/pkg/api/listings"
)

// StartHTTPServer запускает HTTP Gateway и Swagger UI
func StartHTTPServer(ctx context.Context, cfg *config.Config, appLogger *logger.Logger) error {
	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(adapter.CustomHTTPError),
	)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	grpcEndpoint := cfg.Gateway.GRPCServerHost + ":" + cfg.Gateway.GRPCServerPort

	if err := auth_pb.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		appLogger.Error(ctx, "Не удалось зарегистрировать Auth сервис для gRPC-gateway",
			zap.String("endpoint", grpcEndpoint),
			zap.Error(err))
		return err
	}

	if err := listings_pb.RegisterListingsServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		appLogger.Error(ctx, "Не удалось зарегистрировать Listings сервис для gRPC-gateway",
			zap.String("endpoint", grpcEndpoint),
			zap.Error(err))
		return err
	}

	router := chi.NewRouter()
	router.Use(middleware.HTTPLoggerMiddleware(appLogger))

	router.Mount("/", mux)

	router.Get("/swagger/auth.json", func(w http.ResponseWriter, r *http.Request) {
		b, err := os.ReadFile(cfg.Swagger.AuthPath)
		if err != nil {
			appLogger.Error(r.Context(), "Ошибка чтения swagger.json",
				zap.String("path", cfg.Swagger.AuthPath),
				zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write(b)
		if writeErr != nil {
			appLogger.Error(r.Context(), "Ошибка записи ответа", zap.Error(writeErr))
		}
	})

	router.Get("/swagger/listings.json", func(w http.ResponseWriter, r *http.Request) {
		b, err := os.ReadFile(cfg.Swagger.ListingsPath)
		if err != nil {
			appLogger.Error(r.Context(), "Ошибка чтения swagger.json",
				zap.String("path", cfg.Swagger.ListingsPath),
				zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write(b)
		if writeErr != nil {
			appLogger.Error(r.Context(), "Ошибка записи ответа", zap.Error(writeErr))
		}
	})

	router.Get("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		b, err := os.ReadFile(cfg.Swagger.AuthPath)
		if err != nil {
			appLogger.Error(r.Context(), "Ошибка чтения swagger.json",
				zap.String("path", cfg.Swagger.AuthPath),
				zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write(b)
		if writeErr != nil {
			appLogger.Error(r.Context(), "Ошибка записи ответа", zap.Error(writeErr))
		}
	})

	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger.json"),
	))

	router.Get("/swagger/auth/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/auth.json"),
	))

	router.Get("/swagger/listings/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/listings.json"),
	))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})

	httpServer := &http.Server{
		Addr:    cfg.Gateway.Host + ":" + cfg.Gateway.Port,
		Handler: router,
	}

	appLogger.Info(ctx, "HTTP Gateway и Swagger запущены",
		zap.String("address", cfg.Gateway.Host+":"+cfg.Gateway.Port),
		zap.String("grpc_endpoint", grpcEndpoint))

	serverErrors := make(chan error, 1)
	serverStopped := make(chan struct{})

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error(ctx, "Ошибка работы HTTP сервера", zap.Error(err))
			serverErrors <- err
		}
		close(serverErrors)
	}()

	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Gateway.Timeout)
		defer shutdownCancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			appLogger.Error(ctx, "Ошибка при остановке HTTP сервера", zap.Error(err))
		}
		appLogger.Info(ctx, "HTTP сервер остановлен")
		close(serverStopped)
	}()

	select {
	case err := <-serverErrors:
		return err
	case <-ctx.Done():
		<-serverStopped
		return nil
	}
}
