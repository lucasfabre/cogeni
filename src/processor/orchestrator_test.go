package processor

import (
	"testing"

	"github.com/lucasfabre/cogeni/src/config"
)

func TestNewCoordinator_Concurrency(t *testing.T) {
	cfg := &config.Config{
		Concurrency: 42,
	}

	coordinator := NewCoordinator(cfg)
	if coordinator.runtimePool.maxSize != 42 {
		t.Errorf("Expected coordinator to configure runtime pool with max size 42, got %d", coordinator.runtimePool.maxSize)
	}
}
