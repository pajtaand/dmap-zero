package service

import (
	"context"
	"errors"

	"github.com/pajtaand/dmap-zero/internal/agent/manager"
	pb "github.com/pajtaand/dmap-zero/internal/proto"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/emptypb"
)

type configurationService struct {
	pb.UnimplementedConfigurationServiceServer

	configManager *manager.ConfigManager
}

func NewConfigurationService(configManager *manager.ConfigManager) (pb.ConfigurationServiceServer, error) {
	if configManager == nil {
		return nil, errors.New("ConfigManager must not be nil")
	}

	return &configurationService{
		configManager: configManager,
	}, nil
}

func (svc *configurationService) UpdateConfiguration(ctx context.Context, config *pb.AgentConfiguration) (*emptypb.Empty, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Update configuration request")

	log.Info().Msg("Configuration received: " + config.String())
	svc.configManager.ReplaceConfiguration(config.Env)
	return &emptypb.Empty{}, nil
}
