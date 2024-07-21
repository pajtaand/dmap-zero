package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"

	"github.com/andreepyro/dmap-zero/internal/common/wrapper"
	"github.com/andreepyro/dmap-zero/internal/controller/dto"
	"github.com/andreepyro/dmap-zero/internal/controller/manager"
	pb "github.com/andreepyro/dmap-zero/internal/proto"
	"github.com/openziti/edge-api/rest_model"
)

type agentService struct {
	agentManager    *manager.AgentManager
	openZitiWrapper *wrapper.OpenZitiManagementWrapper
}

func NewAgentService(agentManager *manager.AgentManager, openZitiWrapper *wrapper.OpenZitiManagementWrapper) (*agentService, error) {
	if agentManager == nil {
		return nil, errors.New("AgentManager must not be nil")
	}

	if openZitiWrapper == nil {
		return nil, errors.New("OpenZitiManagementWrapper wrapper must not be nil")
	}

	return &agentService{
		agentManager:    agentManager,
		openZitiWrapper: openZitiWrapper,
	}, nil
}

func (svc *agentService) CreateAgent(ctx context.Context, request *dto.CreateAgentRequest) (*dto.CreateAgentResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Create agent request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}
	agentID := svc.agentManager.AddAgent(request.Name, request.Configuration)

	return &dto.CreateAgentResponse{
		ID: agentID,
	}, nil
}

func (svc *agentService) GetAgent(ctx context.Context, request *dto.GetAgentRequest) (*dto.GetAgentResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Get agent request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	agent, err := svc.agentManager.GetAgent(request.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %v", err)
	}

	isEnrolled := false
	if identityID := agent.GetIdentityID(); identityID != "" {
		detail, err := svc.openZitiWrapper.GetIdentityDetail(identityID)
		if err != nil {
			return nil, fmt.Errorf("failed to get identity detail: %v", err)
		}
		isEnrolled = *detail.HasAPISession
	}

	isOnline := false
	presentImages := []string{}
	presentModules := []string{}

	if diag := agent.GetDiagnostics(); diag != nil {
		isOnline = true
		for img := range diag.PresentImages {
			presentImages = append(presentImages, img)
		}
		for mod := range diag.PresentModules {
			presentModules = append(presentModules, mod)
		}
	}

	return &dto.GetAgentResponse{
		Name:           agent.GetName(),
		Configuration:  agent.GetConfiguration(),
		IsEnrolled:     isEnrolled,
		IsOnline:       isOnline,
		PresentImages:  presentImages,
		PresentModules: presentModules,
	}, nil
}

func (svc *agentService) ListAgents(ctx context.Context, request *dto.ListAgentsRequest) (*dto.ListAgentsResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("List agents request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	details, err := svc.openZitiWrapper.ListIdentityDetails()
	if err != nil {
		return nil, fmt.Errorf("failed to get identity details: %v", err)
	}
	tmpIdentityMap := map[string]*rest_model.IdentityDetail{}
	for _, detail := range details {
		if detail == nil || detail.ID == nil {
			log.Warn().Msgf("Either identity detail or its ID is nil, identityDetail=%v+", detail)
			continue
		}
		if _, ok := tmpIdentityMap[*detail.ID]; ok {
			log.Warn().Msgf("Replacing identity in identity map, identityID=%s", *detail.ID)
		}
		tmpIdentityMap[*detail.ID] = detail
	}

	agents := make([]*dto.ListAgentsResponseAgent, 0)
	for _, agent := range svc.agentManager.ListAgents() {
		isEnrolled := false
		if identityID := agent.GetIdentityID(); identityID != "" {
			if identity, ok := tmpIdentityMap[identityID]; ok {
				isEnrolled = *identity.HasAPISession
			} else {
				log.Error().Msgf("Identity is not in identity map, identityID=%s", identityID)
			}
		}

		isOnline := false
		presentImages := []string{}
		presentModules := []string{}

		if diag := agent.GetDiagnostics(); diag != nil {
			isOnline = true
			for img := range diag.PresentImages {
				presentImages = append(presentImages, img)
			}
			for mod := range diag.PresentModules {
				presentModules = append(presentModules, mod)
			}
		}

		agents = append(agents, &dto.ListAgentsResponseAgent{
			ID:             agent.GetID(),
			Name:           agent.GetName(),
			Configuration:  agent.GetConfiguration(),
			IsEnrolled:     isEnrolled,
			IsOnline:       isOnline,
			PresentImages:  presentImages,
			PresentModules: presentModules,
		})
	}
	return &dto.ListAgentsResponse{
		Agents: agents,
	}, nil
}

func (svc *agentService) UpdateAgent(ctx context.Context, request *dto.UpdateAgentRequest) (*dto.UpdateAgentResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Update agent request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	agent, err := svc.agentManager.GetAgent(request.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %v", err)
	}

	if c := agent.GetConfigurationServiceClient(); c != nil {
		log.Info().Msgf("Sending update configuration request: agentID=%s", agent.GetID())

		if _, err := c.UpdateConfiguration(ctx, &pb.AgentConfiguration{
			Env: request.Configuration,
		}); err != nil {
			log.Info().Msgf("Failed to send updated configuration to agentID=%s: %v", agent.GetID(), err)
		}
	}

	agent.SetName(request.Name)
	agent.SetConfiguration(request.Configuration)
	return &dto.UpdateAgentResponse{}, nil
}

func (svc *agentService) DeleteAgent(ctx context.Context, request *dto.DeleteAgentRequest) (*dto.DeleteAgentResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Delete agent request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	agent, err := svc.agentManager.GetAgent(request.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %v", err)
	}

	if identityID := agent.GetIdentityID(); identityID != "" {
		if err := svc.openZitiWrapper.DeleteIdentity(identityID); err != nil {
			return nil, fmt.Errorf("failed to remove identity for agent: %v", err)
		}
	}

	if err := svc.agentManager.RemoveAgent(request.ID); err != nil {
		return nil, fmt.Errorf("failed to remove agent: %v", err)
	}
	return &dto.DeleteAgentResponse{}, nil
}
