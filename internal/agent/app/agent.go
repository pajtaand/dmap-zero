package app

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	agent_grpc "github.com/andreepyro/dmap-zero/internal/agent/grpc"
	"github.com/andreepyro/dmap-zero/internal/agent/manager"
	"github.com/andreepyro/dmap-zero/internal/agent/rest"
	"github.com/andreepyro/dmap-zero/internal/agent/service"
	"github.com/andreepyro/dmap-zero/internal/common/constants"
	"github.com/andreepyro/dmap-zero/internal/common/database"
	mm "github.com/andreepyro/dmap-zero/internal/common/manager"
	"github.com/andreepyro/dmap-zero/internal/common/utils"
	"github.com/andreepyro/dmap-zero/internal/common/wrapper"
	pb "github.com/andreepyro/dmap-zero/internal/proto"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AgentAppConfig struct {
	JWT    string
	KeyAlg string
}

type AgentApp struct {
	cfg AgentAppConfig

	openZitiWrapper        *wrapper.OpenZitiClientWrapper
	dockerWrapper          *wrapper.DockerClientWrapper
	imageManager           *manager.ImageManager
	moduleManager          *manager.ModuleManager
	webhookManager         *manager.WebhookManager
	configManager          *manager.ConfigManager
	endpointManager        *manager.EndpointManager
	receiveServiceClient   pb.ReceiveServiceClient
	setupServiceClient     pb.SetupServiceClient
	phonehomeServiceClient pb.PhonehomeServiceClient
	moduleAuthStore        *mm.AuthStore
	agentServer            *agent_grpc.AgentServer
	p2pServer              *agent_grpc.P2PServer
	moduleServer           *rest.RESTServer
	moduleServerCert       *tls.Certificate
	controllerConn         *grpc.ClientConn

	identityName           string
	moduleServerChosenPort int
	database               *database.KVStore
}

func NewAgentApp(ctx context.Context, cfg AgentAppConfig) (*AgentApp, error) {
	log.Debug().Msg("Initializing new Agent")

	agent := &AgentApp{
		cfg:                    cfg,
		moduleAuthStore:        mm.NewAuthStore(),
		moduleServerChosenPort: constants.AgentModuleServerDefaultPort,
	}

	log.Debug().Msg("Validating configuration")
	if agent.cfg.JWT == "" {
		return nil, errors.New("JWT token is required to create Agent")
	}

	agent.moduleServerChosenPort = utils.FirstAvailablePort(constants.AgentModuleServerDefaultPort)

	log.Debug().Msg("Generating certificates for module REST API")
	certExpiration := time.Now().Add(constants.AgentModuleServerCertificateValidity)
	cert, certPEM, err := utils.GenerateCertificate(constants.AgentDockerHostAddress, certExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key and certificate: %v", err)
	}
	agent.moduleServerCert = cert

	log.Debug().Msg("Initializing wrapper clients")
	dockerWrapper, err := wrapper.NewDockerClientWrapper(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize docker wrapper: %v", err)
	}
	agent.dockerWrapper = dockerWrapper

	openZitiWrapper, err := wrapper.NewOpenZitiClientWrapperFromToken(
		&wrapper.OpenZitiClientWrapperConfig{
			KeyAlg: ziti.KeyAlgVar(cfg.KeyAlg),
		},
		agent.cfg.JWT,
	)
	if err != nil {
		return nil, fmt.Errorf("failed create OpenZitiWrapper: %v", err)
	}
	agent.openZitiWrapper = openZitiWrapper

	identity, err := agent.openZitiWrapper.GetIdentity()
	if err != nil {
		return nil, fmt.Errorf("failed to get agent's identity: %v", err)
	}
	agent.identityName = identity
	log.Info().Msgf("Agent OpenZiti identity: %s", identity)

	log.Debug().Msg("Connecting to database")
	agent.database = database.NewKVStore()

	log.Debug().Msg("Creating managers")
	imageManager, err := manager.NewImageManager(agent.dockerWrapper, agent.database)
	if err != nil {
		return nil, fmt.Errorf("failed to create ImageManager: %v", err)
	}
	agent.imageManager = imageManager

	agentAPIBaseUrl := fmt.Sprintf("https://%s:%d/api/v1", constants.AgentDockerHostAddress, agent.moduleServerChosenPort)
	moduleManager, err := manager.NewModuleManager(agent.dockerWrapper, agent.moduleAuthStore, certPEM, agentAPIBaseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create ModuleManager: %v", err)
	}
	agent.moduleManager = moduleManager

	configManager, err := manager.NewConfigManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create ConfigManager: %v", err)
	}
	agent.configManager = configManager

	endpointManager, err := manager.NewEndpointManager(agent.openZitiWrapper)
	if err != nil {
		return nil, fmt.Errorf("failed to create EndpointManager: %v", err)
	}
	agent.endpointManager = endpointManager

	webhookManager, err := manager.NewWebhookManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create WebhookManager: %v", err)
	}
	agent.webhookManager = webhookManager

	log.Debug().Msg("Creating grpc clients")
	controllerConn, err := grpc.NewClient(
		fmt.Sprintf("passthrough:///%s", constants.OpenZitiServiceController),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(agent.openZitiWrapper.GetContextDialer()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize controller connection: %v", err)
	}
	agent.controllerConn = controllerConn

	agent.receiveServiceClient = pb.NewReceiveServiceClient(agent.controllerConn)
	agent.setupServiceClient = pb.NewSetupServiceClient(agent.controllerConn)
	agent.phonehomeServiceClient = pb.NewPhonehomeServiceClient(agent.controllerConn)

	log.Debug().Msg("Creating agent services")
	pingService, err := service.NewPingService()
	if err != nil {
		return nil, fmt.Errorf("failed to create new PingService: %v", err)
	}
	configurationService, err := service.NewConfigurationService(configManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create new ConfigurationService: %v", err)
	}
	imageService, err := service.NewImageService(agent.imageManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create new ImageService: %v", err)
	}
	moduleService, err := service.NewModuleService(agent.moduleManager, agent.imageManager, agent.configManager, agent.webhookManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create new ModuleService: %v", err)
	}
	shareService, err := service.NewShareService(agent.webhookManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create new ShareService: %v", err)
	}

	log.Debug().Msg("Creating module services")
	controllerService, err := service.NewControllerService(agent.receiveServiceClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create ControllerService: %v", err)
	}
	endpointService, err := service.NewEndpointService(agent.endpointManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create EndpointService: %v", err)
	}
	webhookService, err := service.NewWebhookService(webhookManager, agent.moduleManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebhookService: %v", err)
	}

	log.Debug().Msg("Preparing servers")
	agentListener, err := agent.openZitiWrapper.ListenWithOptions(constants.OpenZitiServiceAgent, &ziti.ListenOptions{
		BindUsingEdgeIdentity: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get listener: %v", err)
	}

	agent.agentServer = agent_grpc.NewAgentServer(
		configurationService,
		imageService,
		moduleService,
		shareService,
		agentListener,
	)

	p2pListener, err := agent.openZitiWrapper.ListenWithOptions(constants.OpenZitiServiceP2P, &ziti.ListenOptions{
		BindUsingEdgeIdentity: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get listener: %v", err)
	}

	agent.p2pServer = agent_grpc.NewP2PServer(
		pingService,
		shareService,
		p2pListener,
	)

	agent.moduleServer = rest.NewRESTServer(
		agent.moduleAuthStore,
		endpointService,
		controllerService,
		webhookService,
	)

	log.Info().Msg("Agent initialization was successful")
	return agent, nil
}

func (a *AgentApp) DownloadConfiguration() error {
	log.Info().Msg("Requesting agent configuration")
	resp, err := a.setupServiceClient.ConfigurationRequest(context.Background(), &emptypb.Empty{})
	if err != nil {
		return fmt.Errorf("failed to get response: %v", err)
	}
	a.configManager.ReplaceConfiguration(resp.Env)
	log.Info().Msgf("Agent configuration received: %+v", resp.Env)
	return nil
}

func (a *AgentApp) DownloadImagesAndStartModules() error {
	log.Info().Msg("Requesting available images")

	stream, err := a.setupServiceClient.ImageRequest(context.Background(), &emptypb.Empty{})
	if err != nil {
		return fmt.Errorf("failed to start stream: %v", err)
	}

	var imageID, imageName string
	var imageData []byte
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				if len(imageData) == 0 {
					log.Info().Msg("No images received from controller")
					return nil
				}
				log.Info().Msgf("Image successfully received: imageID=%s, imageName=%s", imageID, imageName)

				if _, err := a.imageManager.AddImageWithID(imageID, imageName, imageData); err != nil {
					log.Error().Err(err).Msgf("Failed to add image to the ImageManager: imageID=%s, imageName=%s", imageID, imageName)
				}
				break // continue with modules setup
			}
			return fmt.Errorf("failed to receive data: %v", err)
		}
		if imageID == "" {
			imageID = chunk.Id
			imageName = chunk.Name
		} else if imageID != chunk.Id {
			log.Info().Msgf("Image successfully received: imageID=%s, imageName=%s", imageID, imageName)

			if _, err := a.imageManager.AddImageWithID(imageID, imageName, imageData); err != nil {
				log.Error().Err(err).Msgf("Failed to add image to the ImageManager: imageID=%s, imageName=%s", imageID, imageName)
			}

			imageID = chunk.Id
			imageName = chunk.Name
			imageData = []byte{}
		}
		imageData = append(imageData, chunk.Content...)
	}

	log.Info().Msg("Requesting running modules")
	resp, err := a.setupServiceClient.ModuleRequest(context.Background(), &emptypb.Empty{})
	if err != nil {
		return fmt.Errorf("failed to reveice module configuration: %v", err)
	}

	log.Debug().Msgf("Received configuration for %d modules", len(resp.Configs))
	for ind, cfg := range resp.Configs {
		log.Debug().Msgf("[%d] Module data: %v", ind, cfg)

		moduleID := cfg.Module.Id
		imageID := cfg.Image.Id
		moduleCfg := a.configManager.GetConfiguration()

		// extend agent's configuration with module configuration
		for k, v := range cfg.Env {
			moduleCfg[k] = v
		}

		image, err := a.imageManager.GetImage(imageID)
		if err != nil {
			log.Error().Err(err).Msgf("failed to get image, imageID=%s", imageID)
			continue
		}
		imageRef := image.GetReference()

		log.Info().Msgf("Starting module moduleID=%s, imageID=%s, moduleCfg=%v", moduleID, imageID, moduleCfg)

		if err := a.webhookManager.AddModule(moduleID); err != nil {
			log.Error().Err(err).Msg("failed to add module to webhook manager")
			continue
		}

		if _, err := a.moduleManager.StartModule(moduleID, imageRef, moduleCfg); err != nil {
			log.Error().Err(err).Msg("failed to start module")
			continue
		}
	}

	return nil
}

func (a *AgentApp) pingAgents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// context cancelled
			return
		default:
			func() {
				log.Debug().Msg("Looking for other agents...")
				identityIDs, err := a.openZitiWrapper.GetServiceTerminators(constants.OpenZitiServiceP2P)
				if err != nil {
					log.Error().Err(err).Msg("Failed to find any agents")
					return
				}
				for _, identityID := range identityIDs {
					if identityID == a.identityName {
						continue // do not ping itself
					}

					conn, err := grpc.NewClient(
						fmt.Sprintf("passthrough:///%s", constants.OpenZitiServiceP2P),
						grpc.WithTransportCredentials(insecure.NewCredentials()),
						grpc.WithContextDialer(a.openZitiWrapper.GetContextDialerWithOptions(&ziti.DialOptions{
							Identity: identityID,
						})),
					)
					if err != nil {
						log.Error().Err(err).Msg("failed to initialize agent connection")
						continue
					}
					defer conn.Close()
					c := pb.NewPingServiceClient(conn)

					ctx, cancel := context.WithTimeout(context.Background(), time.Second)
					defer cancel()

					log.Debug().Msgf("Pinging other agent: %s", identityID)
					_, err = c.Ping(ctx, &emptypb.Empty{})
					if err != nil {
						log.Error().Err(err).Msgf("Failed to ping agent: %s", identityID)
					}
				}
			}()
		}
		time.Sleep(constants.AgentPingInterval)
	}
}

func (a *AgentApp) repeatPhonehome(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// context cancelled
			return
		default:
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				phonehomeData := &pb.PhonehomeData{
					Images:  map[string]*pb.ImageInfo{},
					Modules: map[string]*pb.ModuleInfo{},
				}
				for _, image := range a.imageManager.ListImages() {
					phonehomeData.Images[image.GetID()] = &pb.ImageInfo{
						Id:   image.GetID(),
						Name: image.GetName(),
						Size: int64(image.GetSize()),
					}
				}
				for _, module := range a.moduleManager.ListModules() {
					phonehomeData.Modules[module.GetID()] = &pb.ModuleInfo{
						Id:     module.GetID(),
						Status: pb.ModuleStatus_UNKNOWN,
					}
				}

				log.Debug().Msg("Phoning home...")
				_, err := a.phonehomeServiceClient.Phonehome(ctx, phonehomeData)
				if err != nil {
					log.Error().Err(err).Msg("Failed to phone home")
				}
			}()
		}
		time.Sleep(constants.AgentPhonehomeInterval)
	}
}

func (a *AgentApp) Run(ctx context.Context) error {
	log.Info().Msg("Starting agent")
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.agentServer.Run(); err != nil {
			panic(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.p2pServer.Run(); err != nil {
			panic(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := a.moduleServer.Run(
			fmt.Sprintf("%s:%d", constants.AgentModuleServerIP, a.moduleServerChosenPort),
			a.moduleServerCert,
		); err != nil {
			panic(err)
		}
	}()

	log.Debug().Msg("Waiting for everything to setup")
	time.Sleep(2 * time.Second)

	if err := a.DownloadConfiguration(); err != nil {
		return fmt.Errorf("failed to download configuration: %v", err)
	}
	if err := a.DownloadImagesAndStartModules(); err != nil {
		return fmt.Errorf("failed to download images: %v", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	go a.repeatPhonehome(ctx)
	go a.pingAgents(ctx)

	log.Info().Msg("Agent successfully started")
	wg.Wait()
	cancel()
	return nil
}

func (a *AgentApp) Stop(ctx context.Context) error {
	if err := a.agentServer.Stop(); err != nil {
		return fmt.Errorf("failed to stop agent server: %v", err)
	}
	if err := a.p2pServer.Stop(); err != nil {
		return fmt.Errorf("failed to stop p2p server: %v", err)
	}
	if err := a.moduleServer.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop p2p server: %v", err)
	}
	return nil
}

func (a *AgentApp) Clean(ctx context.Context) error {
	for _, module := range a.moduleManager.ListModules() {
		a.moduleManager.StopModule(module.GetID())
	}
	for _, image := range a.imageManager.ListImages() {
		a.imageManager.RemoveImage(image.GetID())
	}
	if err := a.dockerWrapper.Close(); err != nil {
		return fmt.Errorf("failed to close docker wrapper: %v", err)
	}
	if err := a.controllerConn.Close(); err != nil {
		return fmt.Errorf("failed to close controller connection: %v", err)
	}
	return nil
}
