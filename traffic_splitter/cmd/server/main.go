package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/yourorg/ml_ab_test_service/traffic_splitter/internal/config"
	"github.com/yourorg/ml_ab_test_service/traffic_splitter/internal/domain"
	"github.com/yourorg/ml_ab_test_service/traffic_splitter/internal/domain/allocation"
	"github.com/yourorg/ml_ab_test_service/traffic_splitter/internal/domain/cache"
	experimentclient "github.com/yourorg/ml_ab_test_service/traffic_splitter/internal/domain/experiment_client"
	modelclient "github.com/yourorg/ml_ab_test_service/traffic_splitter/internal/domain/model_client"
	"github.com/yourorg/ml_ab_test_service/traffic_splitter/internal/domain/updater"
	"github.com/yourorg/ml_ab_test_service/traffic_splitter/internal/hasher"
	"github.com/yourorg/ml_ab_test_service/traffic_splitter/internal/logger"
	"github.com/yourorg/ml_ab_test_service/traffic_splitter/internal/service/splitting"
	grpctransport "github.com/yourorg/ml_ab_test_service/traffic_splitter/internal/transport/grpc_transport"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	lg, err := logger.NewLogger("DEBUG")
	if err != nil {
		panic(err)
	}

	cfg := config.Load()

	store := cache.NewMemoryStore()

	expClient := experimentclient.NewHTTPExperimentClient(cfg.ExperimentManagerURL)
	modelClient := modelclient.NewHTTPModelClient(cfg.ModelRegistryURL)

	loader := updater.NewCacheLoader(expClient, modelClient, store, lg)

	if err := loader.LoadAll(ctx); err != nil {
		if errors.Is(err, domain.ErrNoActiveExperiments) {
			lg.Info("no active experiments at startup, continuing with empty cache")
		} else {
			lg.Warn("initial cache load failed, continuing", zap.Error(err))
		}
	} else {
		lg.Info("cache loaded")
	}

	poller := updater.NewPoller(loader, 30*time.Second, lg)
	go poller.Run(ctx)

	hasher := hasher.NewHasher()
	allocator := allocation.NewAllocator()
	splitter := splitting.NewSplitter(store, hasher, allocator)

	handler := grpctransport.NewHandler(splitter)

	grpcPort, err := strconv.Atoi(cfg.GRPCPort)
	if err != nil {
		panic(err)
	}

	srv := grpctransport.NewServer(grpcPort, handler)

	go func() {
		lg.Info("gRPC server starting", zap.Int("port", grpcPort))
		if err := srv.Run(); err != nil {
			lg.Fatal("gRPC server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()

	lg.Info("shutting down gracefully...")
	srv.Stop()
	lg.Info("server stopped")
}
