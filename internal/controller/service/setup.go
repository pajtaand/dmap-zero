package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"

	"github.com/pajtaand/dmap-zero/internal/common/constants"
	"github.com/pajtaand/dmap-zero/internal/common/utils"
	"github.com/pajtaand/dmap-zero/internal/controller/manager"
	pb "github.com/pajtaand/dmap-zero/internal/proto"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
)

type setupService struct {
	pb.UnimplementedSetupServiceServer

	agentManager  *manager.AgentManager
	imageManager  *manager.ImageManager
	moduleManager *manager.ModuleManager
}

func NewSetupService(agentManager *manager.AgentManager, imageManager *manager.ImageManager, moduleManager *manager.ModuleManager) (pb.SetupServiceServer, error) {
	if agentManager == nil {
		return nil, errors.New("AgentManager must not be nil")
	}
	if imageManager == nil {
		return nil, errors.New("ImageManager must not be nil")
	}
	if moduleManager == nil {
		return nil, errors.New("ModuleManager must not be nil")
	}

	return &setupService{
		agentManager:  agentManager,
		imageManager:  imageManager,
		moduleManager: moduleManager,
	}, nil
}

func (svc *setupService) ConfigurationRequest(ctx context.Context, request *emptypb.Empty) (*pb.AgentConfiguration, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Configuration request request")

	p, ok := peer.FromContext(ctx)
	if !ok {
		err := errors.New("failed to get peer from request context")
		log.Error().Err(err).Msg("")
		return nil, err
	}

	_, _, sourceIdentity, err := utils.ParseOpenZitiAddress(p.LocalAddr.String())
	if err != nil {
		err := fmt.Errorf("failed to parse source address: %v", err)
		log.Error().Err(err).Msg("")
		return nil, err
	}

	log.Info().Msgf("Caller identity: %s", sourceIdentity)

	agent, err := svc.agentManager.GetAgent(sourceIdentity)
	if err != nil {
		err := fmt.Errorf("failed to get agent: %v", err)
		log.Error().Err(err).Msg("")
		return nil, err
	}

	return &pb.AgentConfiguration{
		Env: agent.GetConfiguration(),
	}, nil
}

func (svc *setupService) ImageRequest(request *emptypb.Empty, stream pb.SetupService_ImageRequestServer) error {
	log := zerolog.Ctx(stream.Context())
	log.Info().Msg("Image request request")

	p, ok := peer.FromContext(stream.Context())
	if !ok {
		err := errors.New("failed to get peer from request context")
		log.Error().Err(err).Msg("")
		return err
	}

	_, _, sourceIdentity, err := utils.ParseOpenZitiAddress(p.LocalAddr.String())
	if err != nil {
		err := fmt.Errorf("failed to parse source address: %v", err)
		log.Error().Err(err).Msg("")
		return err
	}

	log.Info().Msgf("Caller identity: %s", sourceIdentity)

	images := svc.imageManager.ListImages()
	for _, image := range images {
		imageID := image.GetID()

		data, err := image.GetData()
		if err != nil {
			err := fmt.Errorf("failed to get image data: imageID=%s, agentID=%s: %v", imageID, sourceIdentity, err)
			log.Error().Err(err).Msg("")
			return err
		}

		log.Info().Msgf("Streaming image to agent: imageID=%s, agentID=%s: %v", imageID, sourceIdentity, err)
		for start := 0; start < len(data); start += constants.AgentImageStreamChunkSize {
			end := start + constants.AgentImageStreamChunkSize
			if end > len(data) {
				end = len(data)
			}
			if err := stream.Send(&pb.ImageStreamData{
				Id:      imageID,
				Name:    image.GetName(),
				Content: data[start:end],
			}); err != nil {
				err := fmt.Errorf("failed to stream image to agent: imageID=%s, agentID=%s: %v", imageID, sourceIdentity, err)
				log.Error().Err(err).Msg("")
				return err
			}
		}
	}
	return nil
}

func (svc *setupService) ModuleRequest(ctx context.Context, request *emptypb.Empty) (*pb.ModuleConfigurations, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Module request request")

	p, ok := peer.FromContext(ctx)
	if !ok {
		err := errors.New("failed to get peer from request context")
		log.Error().Err(err).Msg("")
		return nil, err
	}

	_, _, sourceIdentity, err := utils.ParseOpenZitiAddress(p.LocalAddr.String())
	if err != nil {
		err := fmt.Errorf("failed to parse source address: %v", err)
		log.Error().Err(err).Msg("")
		return nil, err
	}

	log.Info().Msgf("Caller identity: %s", sourceIdentity)

	if _, err := svc.agentManager.GetAgent(sourceIdentity); err != nil {
		err := fmt.Errorf("failed to get agent: %v", err)
		log.Error().Err(err).Msg("")
		return nil, err
	}

	configs := []*pb.ModuleConfiguration{}
	for _, module := range svc.moduleManager.ListModules() {
		if module.IsRunning() {
			configs = append(configs, &pb.ModuleConfiguration{
				Module: &pb.ModuleIdentifier{
					Id: module.GetID(),
				},
				Image: &pb.ImageIdentifier{
					Id: module.GetImage(),
				},
				Env: module.GetConfiguration(),
			})
		}
	}

	return &pb.ModuleConfigurations{
		Configs: configs,
	}, nil
}
