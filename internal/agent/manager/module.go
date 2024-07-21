package manager

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/andreepyro/dmap-zero/internal/common/constants"
	errs "github.com/andreepyro/dmap-zero/internal/common/errors"
	mm "github.com/andreepyro/dmap-zero/internal/common/manager"
	"github.com/andreepyro/dmap-zero/internal/common/utils"
	"github.com/andreepyro/dmap-zero/internal/common/wrapper"
	"github.com/docker/docker/api/types/container"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Module struct {
	id            string
	imageRef      string
	containerID   string
	configuration map[string]string
	givenPort     string

	mu sync.RWMutex
}

func NewModule(id, imageRef, containerID string, configuration map[string]string, givenPort string) *Module {
	if configuration == nil {
		configuration = map[string]string{}
	}

	return &Module{
		id:            id,
		imageRef:      imageRef,
		containerID:   containerID,
		configuration: configuration,
		givenPort:     givenPort,
	}
}

func (m *Module) GetID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.id
}

func (m *Module) GetImageReference() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.imageRef
}

func (m *Module) GetContainerID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.containerID
}

func (m *Module) GetConfiguration() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.configuration
}

func (m *Module) GetGivenPort() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.givenPort
}

type ModuleManager struct {
	mu            sync.RWMutex
	modules       map[string]*Module
	dockerWrapper *wrapper.DockerClientWrapper
	authStore     *mm.AuthStore

	apiBaseUrl          string
	moduleServerCertPEM []byte
	portCounter         int
}

func NewModuleManager(dockerWrapper *wrapper.DockerClientWrapper, authStore *mm.AuthStore, moduleServerCertPEM []byte, apiBaseUrl string) (*ModuleManager, error) {
	log.Debug().Msg("Creating new ModuleManager")

	if dockerWrapper == nil {
		return nil, errors.New("DockerClientWrapper must not be nil")
	}
	if authStore == nil {
		return nil, errors.New("AuthStore must not be nil")
	}
	if moduleServerCertPEM == nil {
		return nil, errors.New("moduleServerCertPEM must not be nil")
	}
	if apiBaseUrl == "" {
		return nil, errors.New("apiBaseUrl must be set")
	}

	return &ModuleManager{
		modules:             map[string]*Module{},
		dockerWrapper:       dockerWrapper,
		authStore:           authStore,
		portCounter:         constants.ModulePortRangeMin,
		apiBaseUrl:          apiBaseUrl,
		moduleServerCertPEM: moduleServerCertPEM,
	}, nil
}

func (mgr *ModuleManager) StartModule(id, imageRef string, configuration map[string]string) (*Module, error) {
	log.Info().Msgf("Starting module: %s", imageRef)

	// convert configuration map to variable list
	envCfg := []string{}
	for k, v := range configuration {
		envCfg = append(envCfg, fmt.Sprintf("%s=%s", k, v))
	}

	// choose allowed port for module
	givenPort := ""
	for {
		log.Info().Msgf("Trying port: %d...", mgr.portCounter)
		mgr.portCounter += 1
		if utils.TCPPortAvailable(mgr.portCounter) {
			log.Info().Msgf("using port: %d", mgr.portCounter)
			givenPort = strconv.Itoa(mgr.portCounter)
			break
		}
		if mgr.portCounter > constants.ModulePortRangeMax {
			mgr.portCounter = constants.ModulePortRangeMin
		}
	}

	// generate module api credentials
	moduleUsername := id
	modulePassword := uuid.New().String()
	base64Cert := base64.StdEncoding.EncodeToString(mgr.moduleServerCertPEM)
	envCfg = append(envCfg, fmt.Sprintf("%s=%s", constants.ModuleEnvAPIBaseUrl, mgr.apiBaseUrl))
	envCfg = append(envCfg, fmt.Sprintf("%s=%s", constants.ModuleEnvUsername, moduleUsername))
	envCfg = append(envCfg, fmt.Sprintf("%s=%s", constants.ModuleEnvPassword, modulePassword))
	envCfg = append(envCfg, fmt.Sprintf("%s=%s", constants.ModuleEnvCertificate, base64Cert))
	envCfg = append(envCfg, fmt.Sprintf("%s=%s", constants.ModuleEnvGivenPort, givenPort))
	mgr.authStore.Add(moduleUsername, modulePassword)

	info, err := mgr.dockerWrapper.InspectImage(context.Background(), imageRef)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect image, imageRef=%s, err: %v", imageRef, err)
	}

	containerName := fmt.Sprintf("module_%s_%s", id, givenPort)
	containerID, err := mgr.dockerWrapper.RunContainer(context.Background(), &container.Config{
		Image: imageRef,
		Env:   envCfg,
		Cmd:   info.Config.Cmd,
	}, &container.HostConfig{
		NetworkMode: "host",
	}, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to start module: %v", err)
	}

	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	module := NewModule(id, imageRef, containerID, configuration, givenPort)
	mgr.modules[id] = module
	return module, nil
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

func (mgr *ModuleManager) StopModule(moduleID string) error {
	log.Info().Msgf("Stopping module: %s", moduleID)

	mgr.mu.Lock()
	module, ok := mgr.modules[moduleID]
	mgr.mu.Unlock()

	if !ok {
		return errs.ErrNotFound
	}

	mgr.authStore.Remove(module.GetID())

	if err := mgr.dockerWrapper.StopContainer(context.Background(), module.GetContainerID()); err != nil {
		return fmt.Errorf("failed to stop container: %v", err)
	}
	if err := mgr.dockerWrapper.RemoveContainer(context.Background(), module.GetContainerID()); err != nil {
		return fmt.Errorf("failed to remove container: %v", err)
	}

	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	delete(mgr.modules, moduleID)

	return nil
}
