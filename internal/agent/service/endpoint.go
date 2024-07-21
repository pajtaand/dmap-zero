package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/andreepyro/dmap-zero/internal/agent/dto"
	"github.com/andreepyro/dmap-zero/internal/agent/manager"
	"github.com/rs/zerolog"
)

type endpointService struct {
	endpointManager *manager.EndpointManager
}

func NewEndpointService(endpointManager *manager.EndpointManager) (*endpointService, error) {
	if endpointManager == nil {
		return nil, errors.New("EndpointManager must not be nil")
	}
	return &endpointService{
		endpointManager: endpointManager,
	}, nil
}

func (svc *endpointService) ListEndpoints(ctx context.Context, request *dto.ListEndpointsRequest) (*dto.ListEndpointsResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("List endpoints request")

	endpoints, err := svc.endpointManager.ListEndpoints()
	if err != nil {
		return nil, fmt.Errorf("failed to list endpoints: %v", err)
	}

	endpointList := make([]*dto.ListEndpointsResponseEndpoint, 0, len(endpoints))
	for _, endpoint := range endpoints {
		endpointList = append(endpointList, &dto.ListEndpointsResponseEndpoint{
			ID: endpoint,
		})
	}

	return &dto.ListEndpointsResponse{
		Endpoints: endpointList,
	}, nil
}

func (svc *endpointService) PushBlob(ctx context.Context, request *dto.EndpointPushBlobRequest) (*dto.EndpointPushBlobResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Push blob request")

	if err := svc.endpointManager.SendData(ctx, request.ReceiverIdentityID, request.ReceiverModuleID, request.Blob); err != nil {
		return nil, fmt.Errorf("failed to send data to IdentityID=%s, ModuleID=%s, reason: %v", request.ReceiverIdentityID, request.ReceiverModuleID, err)
	}
	return &dto.EndpointPushBlobResponse{}, nil
}
