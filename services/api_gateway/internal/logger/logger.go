package logger

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/common/contextkeys"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(level string) (*zap.Logger, error) {
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = zapcore.InfoLevel
	}

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(lvl),
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return config.Build()
}

func ZapMiddleware(lg *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			reqID := r.Header.Get("X-Request-Id")
			if reqID == "" {
				reqID = middleware.GetReqID(r.Context())
			}

			ctx := context.WithValue(r.Context(), contextkeys.ReqId, reqID)
			r = r.WithContext(ctx)

			defer func() {
				status := ww.Status()
				duration := time.Since(start)
				lg.Info("http.request",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("remote_addr", r.RemoteAddr),
					zap.Int("status", status),
					zap.Duration("duration", duration),
					zap.String("request_id", reqID),
					zap.Int("bytes", ww.BytesWritten()),
				)

				if status >= 500 {
					lg.Error("http.server_error", zap.Int("status", status))
				} else if status >= 400 {
					lg.Warn("http.client_error", zap.Int("status", status))
				}
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
