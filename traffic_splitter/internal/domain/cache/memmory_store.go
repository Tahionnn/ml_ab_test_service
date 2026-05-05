package cache

import (
	"context"
	"sync"

	"github.com/yourorg/ml_ab_test_service/traffic_splitter/internal/domain"
)

type MemoryStore struct {
	mu      sync.RWMutex
	configs map[string]domain.ExperimentConfig
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		configs: make(map[string]domain.ExperimentConfig),
	}
}

func (s *MemoryStore) Get(_ context.Context, experimentID string) (domain.ExperimentConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg, ok := s.configs[experimentID]
	return cfg, ok
}

func (s *MemoryStore) Set(experimentID string, config domain.ExperimentConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.configs[experimentID] = config
}

func (s *MemoryStore) SetAll(configs map[string]domain.ExperimentConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.configs = configs
}

func (s *MemoryStore) UpdateModelEndpoint(modelID string, endpoint string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for expID, cfg := range s.configs {
		for i, v := range cfg.Variants {
			if v.ModelId == modelID {
				cfg.Variants[i].ServingEndpoint = endpoint
				s.configs[expID] = cfg
			}
		}
	}
}
