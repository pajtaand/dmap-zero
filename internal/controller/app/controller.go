package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/pajtaand/dmap-zero/internal/common/constants"
	"github.com/pajtaand/dmap-zero/internal/common/database"
	mm "github.com/pajtaand/dmap-zero/internal/common/manager"
	"github.com/pajtaand/dmap-zero/internal/common/wrapper"
	ctrl_grpc "github.com/pajtaand/dmap-zero/internal/controller/grpc"
	"github.com/pajtaand/dmap-zero/internal/controller/manager"
	"github.com/pajtaand/dmap-zero/internal/controller/metrics"
	"github.com/pajtaand/dmap-zero/internal/controller/rest"
	"github.com/pajtaand/dmap-zero/internal/controller/service"

	"github.com/openziti/sdk-golang/ziti"
	"github.com/rs/zerolog/log"
)

type ControllerAppConfig struct {
	ApiCredentials map[string]string
	RESTapi        struct {
		Address  string
		CertFile string
		KeyFile  string
	}
	MetricsApi struct {
		Address  string
		CertFile string
		KeyFile  string
	}
	OpenZiti struct {
		KeyAlg          ziti.KeyAlgVar
		EnrollmentToken string
	}
}

type ControllerApp struct {
	cfg           *ControllerAppConfig
	agentServer   *ctrl_grpc.RPCServer
	clientServer  *rest.RESTServer
	metricsServer *metrics.MetricsServer

	agentManager   *manager.AgentManager
	moduleManager  *manager.ModuleManager
	imageManager   *manager.ImageManager
	webhookManager *manager.WebhookManager
	userAuthStore  *mm.AuthStore
}

func NewControllerApp(cfg *ControllerAppConfig) (*ControllerApp, error) {
	log.Debug().Msg("Initializing new Controller")

	if cfg.ApiCredentials == nil {
		return nil, errors.New("value ApiCredentials is nill")
	}
	if cfg.RESTapi.Address == "" {
		return nil, errors.New("value Address for RESTapi not set")
	}
	if cfg.RESTapi.CertFile == "" {
		return nil, errors.New("value CertFile for RESTapi not set")
	}
	if cfg.RESTapi.KeyFile == "" {
		return nil, errors.New("value KeyFile for RESTapi not set")
	}
	if cfg.MetricsApi.Address == "" {
		return nil, errors.New("value Address for MetricsApi not set")
	}
	if cfg.MetricsApi.CertFile == "" {
		return nil, errors.New("value CertFile for MetricsApi not set")
	}
	if cfg.MetricsApi.KeyFile == "" {
		return nil, errors.New("value KeyFile for MetricsApi not set")
	}
	if cfg.OpenZiti.KeyAlg == "" {
		return nil, errors.New("value KeyAlg for OpenZiti not set")
	}
	if cfg.OpenZiti.EnrollmentToken == "" {
		return nil, errors.New("value EnrollmentToken for OpenZiti not set")
	}

	log.Info().Msg("Controller initialization was successful")
	return &ControllerApp{
		cfg: cfg,
	}, nil
}

func (app *ControllerApp) Setup(ctx context.Context) error {
	log.Info().Msg("Setting up controller")

	if err := app.Clean(ctx); err != nil {
		return err
	}

	// OpenZiti Identity
	openZitiClient, err := wrapper.NewOpenZitiClientWrapperFromToken(
		&wrapper.OpenZitiClientWrapperConfig{
			KeyAlg: app.cfg.OpenZiti.KeyAlg,
		},
		app.cfg.OpenZiti.EnrollmentToken,
	)
	if err != nil {
		return fmt.Errorf("failed to register to OpenZiti: %v", err)
	}

	// OpenZiti ManagementAPI
	openZitiWrapper, err := wrapper.NewOpenZitiManagementWrapper(openZitiClient.GetOpenZitiConfig())
	if err != nil {
		return err
	}

	log.Debug().Msg("Connecting to database")
	imageDatabase := database.NewKVStore()

	log.Debug().Msg("Creating managers")
	agentManager, err := manager.NewAgentManager(&manager.AgentManagerConfig{
		AgentServiceName: constants.OpenZitiServiceAgent,
	}, openZitiClient)
	if err != nil {
		return fmt.Errorf("failed to create AgentManager: %v", err)
	}
	moduleManager, err := manager.NewModuleManager()
	if err != nil {
		return fmt.Errorf("failed to create ModuleManager: %v", err)
	}
	imageManager, err := manager.NewImageManager(imageDatabase)
	if err != nil {
		return fmt.Errorf("failed to create ImageManager: %v", err)
	}
	webhookManager, err := manager.NewWebhookManager()
	if err != nil {
		return fmt.Errorf("failed to create WebhookManager: %v", err)
	}
	userAuthStore := mm.NewAuthStore()
	for username, password := range app.cfg.ApiCredentials {
		userAuthStore.Add(username, password)
	}
	app.agentManager = agentManager
	app.moduleManager = moduleManager
	app.imageManager = imageManager
	app.webhookManager = webhookManager
	app.userAuthStore = userAuthStore

	log.Debug().Msg("Creating services")
	agentService, err := service.NewAgentService(agentManager, openZitiWrapper)
	if err != nil {
		return fmt.Errorf("failed to create AgentService: %v", err)
	}
	moduleService, err := service.NewModuleService(moduleManager, imageManager, agentManager)
	if err != nil {
		return fmt.Errorf("failed to create ModuleService: %v", err)
	}
	imageService, err := service.NewImageService(imageManager, agentManager, moduleManager)
	if err != nil {
		return fmt.Errorf("failed to create ImageService: %v", err)
	}
	webhookService, err := service.NewWebhookService(webhookManager, moduleManager)
	if err != nil {
		return fmt.Errorf("failed to create WebhookService: %v", err)
	}
	enrollmentService, err := service.NewEnrollmentService(agentManager, openZitiWrapper)
	if err != nil {
		return fmt.Errorf("failed to create EnrollmentService: %v", err)
	}
	phonehomeService, err := service.NewPhonehomeService(agentManager)
	if err != nil {
		return fmt.Errorf("failed to create HealthService: %v", err)
	}
	setupService, err := service.NewSetupService(agentManager, imageManager, moduleManager)
	if err != nil {
		return fmt.Errorf("failed to create SetupService: %v", err)
	}
	receiveService, err := service.NewReceiveService(webhookManager)
	if err != nil {
		return fmt.Errorf("failed to create ReceiveService: %v", err)
	}

	log.Debug().Msg("Preparing servers")
	app.clientServer = rest.NewRESTServer(
		app.userAuthStore,
		agentService,
		moduleService,
		imageService,
		webhookService,
		enrollmentService,
	)

	listener, err := openZitiClient.Listen(constants.OpenZitiServiceController)
	if err != nil {
		return fmt.Errorf("failed to get listener: %v", err)
	}

	app.agentServer = ctrl_grpc.NewRPCServer(
		phonehomeService,
		setupService,
		receiveService,
		listener,
	)

	app.metricsServer = metrics.NewMetricsServer()

	log.Info().Msg("Controller successfully set up")

	return nil
}

func (app *ControllerApp) Run(ctx context.Context) error {
	log.Info().Msg("Starting controller")
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.agentServer.Run(); err != nil {
			panic(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.clientServer.Run(
			app.cfg.RESTapi.Address,
			app.cfg.RESTapi.CertFile,
			app.cfg.RESTapi.KeyFile,
		); err != nil {
			if err != http.ErrServerClosed {
				panic(err)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.metricsServer.Run(
			app.cfg.MetricsApi.Address,
			app.cfg.MetricsApi.CertFile,
			app.cfg.MetricsApi.KeyFile,
		); err != nil {
			if err != http.ErrServerClosed {
				panic(err)
			}
		}
	}()

	log.Info().Msg("Controller successfully started")
	wg.Wait()

	return nil
}

func (app *ControllerApp) Stop(ctx context.Context) error {
	log.Info().Msg("Stopping Controller")

	if err := app.agentServer.Stop(); err != nil {
		return fmt.Errorf("failed to stop agent server: %v", err)
	}
	if err := app.clientServer.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop client server: %v", err)
	}
	return nil
}

func (app *ControllerApp) Clean(ctx context.Context) error {
	log.Debug().Msg("Cleaning up")

	if app.agentManager != nil {
		for _, agent := range app.agentManager.ListAgents() {
			if err := app.agentManager.RemoveAgent(agent.GetID()); err != nil {
				return err
			}
		}
	}
	if app.imageManager != nil {
		for _, image := range app.imageManager.ListImages() {
			if err := app.imageManager.RemoveImage(image.GetID()); err != nil {
				return err
			}
		}
	}
	return nil
}
