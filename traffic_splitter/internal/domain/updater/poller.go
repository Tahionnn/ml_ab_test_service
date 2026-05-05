package updater

import (
	"context"
	"errors"
	"time"

	"github.com/yourorg/ml_ab_test_service/traffic_splitter/internal/domain"
	"go.uber.org/zap"
)

type Poller struct {
	loader   *CacheLoader
	interval time.Duration
	logger   *zap.Logger
}

func NewPoller(loader *CacheLoader, interval time.Duration, logger *zap.Logger) *Poller {
	return &Poller{loader: loader, interval: interval, logger: logger}
}

func (p *Poller) Run(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := p.loader.LoadAll(ctx); err != nil {
				if errors.Is(err, domain.ErrNoActiveExperiments) {
					p.logger.Info("no active experiments, cache empty")
					continue
				}

				p.logger.Error("failed to load cache",
					zap.Error(err),
					zap.String("component", "poller"),
				)
			} else {
				p.logger.Info("cache successfully updated")
			}
		case <-ctx.Done():
			return
		}
	}
}
