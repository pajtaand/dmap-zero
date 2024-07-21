package manager

import (
	"sync"

	"github.com/rs/zerolog/log"
)

type ConfigManager struct {
	mu            sync.RWMutex
	configuration map[string]string
}

func NewConfigManager() (*ConfigManager, error) {
	log.Debug().Msg("Creating new ConfigManager")

	return &ConfigManager{
		configuration: map[string]string{},
	}, nil
}

func (mgr *ConfigManager) GetValue(key string) (string, bool) {
	log.Debug().Msgf("Getting config value for key: %s", key)

	mgr.mu.RLock()
	defer mgr.mu.RUnlock()
	value, ok := mgr.configuration[key]
	return value, ok
}

func (mgr *ConfigManager) GetConfiguration() map[string]string {
	log.Debug().Msg("Getting configuration")

	mgr.mu.RLock()
	defer mgr.mu.RUnlock()
	return mgr.configuration
}

func (mgr *ConfigManager) ReplaceConfiguration(configuration map[string]string) {
	log.Debug().Msg("Replacing configuration")

	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	if configuration == nil {
		configuration = map[string]string{}
	}
	mgr.configuration = configuration
}
