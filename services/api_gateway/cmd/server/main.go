// @title           ML A/B Test Service — API Gateway
// @version         1.0
// @description     Gateway for A/B testing system for ML models.
// @description     Routes requests between User Service, Experiment Manager, Model Registry and Traffic Splitter.

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey    BearerAuth
// @scheme 						bearer
// @bearerFormat 				JWT
// @in                          header
// @name                        Authorization
// @description                 Enter the token in the format: Bearer {token}

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/clients/experiment"
	modelsservice "github.com/tahion/ml_ab_test_service/services/api_gateway/internal/clients/models_service"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/clients/splitter"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/clients/user"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/config"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/gateway"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/logger"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/transport/broker"
	httptransport "github.com/tahion/ml_ab_test_service/services/api_gateway/internal/transport/http_transport"
	mw "github.com/tahion/ml_ab_test_service/services/api_gateway/internal/transport/http_transport/middleware"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/tahion/ml_ab_test_service/services/api_gateway/docs"
)

func main() {
	lg, err := logger.NewLogger("DEBUG")
	if err != nil {
		panic(err)
	}
	defer lg.Sync()

	zapMiddleware := logger.ZapMiddleware(lg)

	cfg := config.Load()

	userClient := user.NewHTTPUserClient(user.UserConfig{
		BaseURL: cfg.UserServiceURL,
		Timeout: 5 * time.Second,
	})

	expClient := experiment.NewExperimentClient(experiment.Config{
		BaseURL: cfg.ExperimentManagerURL,
		Timeout: 5 * time.Second,
	})

	modelClient := modelsservice.NewHTTPModelClient(modelsservice.Config{
		BaseURL: cfg.ModelRegistryURL,
		Timeout: 5 * time.Second,
	})

	splitterClient, err := splitter.NewGRPCClient(splitter.SplitterConfig{
		Address: cfg.SplitterAddress,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Fatal("failed to connect to splitter", zap.Error(err))
	}
	defer splitterClient.Close()

	kafkaProducer := broker.NewProducer(cfg.KafkaBrokers, cfg.PredictionTopic, lg)
	defer kafkaProducer.Close()

	svc := gateway.NewService(userClient, expClient, modelClient, splitterClient, kafkaProducer, lg)

	authMiddleware := mw.Auth(svc)

	authHandler := httptransport.NewAuthHandler(svc, lg)
	expHandler := httptransport.NewExperimentHandler(svc, lg)
	modelHandler := httptransport.NewModelHandler(svc, lg)
	predictHandler := httptransport.NewPredictHandler(svc, lg)
	healthHandler := httptransport.NewHealthHandler(svc, lg)

	r := chi.NewRouter()

	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Recoverer,
		zapMiddleware,
		middleware.Timeout(30*time.Second),
		middleware.Throttle(100),
		cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			AllowCredentials: true,
		}).Handler,
	)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/api/v1", func(r chi.Router) {
		healthHandler.RegisterRoutes(r, authMiddleware)
		authHandler.RegisterRoutes(r, authMiddleware)
		expHandler.RegisterRoutes(r, authMiddleware)
		modelHandler.RegisterRoutes(r, authMiddleware)
		predictHandler.RegisterRoutes(r, authMiddleware)
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		lg.Info("starting gateway", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		lg.Error("shutdown error", zap.Error(err))
	}
	lg.Info("gateway stopped")
}
