package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"

	"github.com/pajtaand/dmap-zero/internal/agent/manager"
	pb "github.com/pajtaand/dmap-zero/internal/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type moduleService struct {
	pb.UnimplementedModuleServiceServer

	moduleManager  *manager.ModuleManager
	imageManager   *manager.ImageManager
	configManager  *manager.ConfigManager
	webhookManager *manager.WebhookManager
}

func NewModuleService(moduleManager *manager.ModuleManager, imageManager *manager.ImageManager, configManager *manager.ConfigManager, webhookManager *manager.WebhookManager) (pb.ModuleServiceServer, error) {
	if moduleManager == nil {
		return nil, errors.New("ModuleManager must not be nil")
	}
	if imageManager == nil {
		return nil, errors.New("ImageManager must not be nil")
	}
	if configManager == nil {
		return nil, errors.New("ConfigManager must not be nil")
	}
	if webhookManager == nil {
		return nil, errors.New("WebhookManager must not be nil")
	}

	return &moduleService{
		moduleManager:  moduleManager,
		imageManager:   imageManager,
		configManager:  configManager,
		webhookManager: webhookManager,
	}, nil
}

func (svc *moduleService) StartModule(ctx context.Context, cfg *pb.ModuleConfiguration) (*emptypb.Empty, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Start module request")

	moduleID := cfg.Module.Id
	imageID := cfg.Image.Id
	moduleCfg := svc.configManager.GetConfiguration()

	// extend agent's configuration with module configuration
	for k, v := range cfg.Env {
		moduleCfg[k] = v
	}

	image, err := svc.imageManager.GetImage(imageID)
	if err != nil {
		err := fmt.Errorf("failed to get image, imageID=%s, err: %v", imageID, err)
		log.Error().Err(err).Msg("")
		return nil, err
	}
	imageRef := image.GetReference()

	log.Info().Msgf("Starting module moduleID=%s, imageID=%s, moduleCfg=%v", moduleID, imageID, moduleCfg)

	if err := svc.webhookManager.AddModule(moduleID); err != nil {
		err := fmt.Errorf("failed to add module to webhook manager: %v", err)
		log.Error().Err(err).Msg("")
		return nil, err
	}

	if _, err := svc.moduleManager.StartModule(moduleID, imageRef, moduleCfg); err != nil {
		err := fmt.Errorf("failed to start module: %v", err)
		log.Error().Err(err).Msg("")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (svc *moduleService) StopModule(ctx context.Context, module *pb.ModuleIdentifier) (*emptypb.Empty, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Stop module request")

	if err := svc.moduleManager.StopModule(module.Id); err != nil {
		err := fmt.Errorf("failed to stop module: %v", err)
		log.Error().Err(err).Msg("")
		return nil, err
	}

	if err := svc.webhookManager.RemoveModule(module.Id); err != nil {
		err := fmt.Errorf("failed to remove module from webhook manager: %v", err)
		log.Error().Err(err).Msg("")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
