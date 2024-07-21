package manager

import (
	"sync"

	errs "github.com/andreepyro/dmap-zero/internal/common/errors"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Module struct {
	id            string
	name          string
	image         string
	configuration map[string]string
	isRunning     bool

	mu sync.RWMutex
}

func NewModule(id, name, image string, configuration map[string]string) *Module {
	if configuration == nil {
		configuration = map[string]string{}
	}

	return &Module{
		id:            id,
		name:          name,
		image:         image,
		configuration: configuration,
		isRunning:     false,
	}
}

func (m *Module) GetID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.id
}

func (m *Module) GetName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.name
}

func (m *Module) SetName(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.name = name
}

func (m *Module) GetImage() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.image
}

func (m *Module) SetImage(image string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.image = image
}

func (m *Module) GetConfiguration() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.configuration
}

func (m *Module) SetConfiguration(configuration map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if configuration == nil {
		configuration = map[string]string{}
	}
	m.configuration = configuration
}

func (m *Module) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isRunning
}

func (m *Module) SetRunning(isRunning bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isRunning = isRunning
}

type ModuleManager struct {
	mu      sync.RWMutex
	modules map[string]*Module
}

func NewModuleManager() (*ModuleManager, error) {
	log.Debug().Msg("Creating new ModuleManager")

	return &ModuleManager{
		modules: map[string]*Module{},
	}, nil
}

func (mgr *ModuleManager) AddModule(name, image string, configuration map[string]string) string {
	log.Info().Msgf("Adding new module: %s", name)

	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	moduleID := uuid.New().String()
	mgr.modules[moduleID] = NewModule(moduleID, name, image, configuration)
	return moduleID
}

func (mgr *ModuleManager) GetModule(moduleID string) (*Module, error) {
	log.Info().Msgf("Getting module: %s", moduleID)

	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	module, ok := mgr.modules[moduleID]
	if !ok {
		return nil, errs.ErrNotFound
	}

	return module, nil
}

func (mgr *ModuleManager) ListModules() []*Module {
	log.Info().Msg("Listing all modules")

	mgr.mu.RLock()
	defer mgr.mu.RUnlock()
	modules := []*Module{}
	for _, module := range mgr.modules {
		modules = append(modules, module)
	}
	return modules
}

func (mgr *ModuleManager) RemoveModule(moduleID string) error {
	log.Info().Msgf("Removing module: %s", moduleID)

	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	_, ok := mgr.modules[moduleID]
	if !ok {
		return errs.ErrNotFound
	}

	delete(mgr.modules, moduleID)
	return nil
}

func (mgr *ModuleManager) ModuleExists(moduleID string) bool {
	log.Info().Msgf("Checking if module exists: %s", moduleID)
	_, ok := mgr.modules[moduleID]
	return ok
}
