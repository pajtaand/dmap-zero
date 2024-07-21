package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/andreepyro/dmap-zero/internal/common/utils"
	"github.com/andreepyro/dmap-zero/internal/controller/manager"
	"github.com/andreepyro/dmap-zero/internal/controller/metrics"
	pb "github.com/andreepyro/dmap-zero/internal/proto"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
)

type phonehomeService struct {
	pb.UnimplementedPhonehomeServiceServer

	agentManager *manager.AgentManager
}

func NewPhonehomeService(agentManager *manager.AgentManager) (pb.PhonehomeServiceServer, error) {
	if agentManager == nil {
		return nil, errors.New("AgentManager must not be nil")
	}

	return &phonehomeService{
		agentManager: agentManager,
	}, nil
}

func (svc *phonehomeService) Phonehome(ctx context.Context, data *pb.PhonehomeData) (*emptypb.Empty, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Phonehome request")

	if data == nil {
		return nil, errors.New("data must not be nil")
	}

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
		err := fmt.Errorf("failed to get agent by identity: %v", err)
		log.Error().Err(err).Msg("")
		return nil, err
	}

	presentImage := map[string]string{}
	for key, value := range data.Images {
		presentImage[key] = value.Id
	}
	metrics.AgentPresentImagesGauge.WithLabelValues(agent.GetID()).Set(float64(len(data.Images)))

	presentModules := map[string]string{}
	for key, value := range data.Modules {
		presentModules[key] = value.Id
	}
	metrics.AgentRunningModulesGauge.WithLabelValues(agent.GetID()).Set(float64(len(data.Modules)))

	if err := svc.agentManager.ReceiveAgentDiagnostics(sourceIdentity, &manager.Diagnostics{
		PresentImages:  presentImage,
		PresentModules: presentModules,
	}); err != nil {
		err := fmt.Errorf("failed to push agent diagnostics: %v", err)
		log.Error().Err(err).Msg("")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
