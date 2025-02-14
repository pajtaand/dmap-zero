package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/pajtaand/dmap-zero/internal/agent/dto"
	pb "github.com/pajtaand/dmap-zero/internal/proto"
	"github.com/rs/zerolog"
)

type controllerService struct {
	receiveServiceClient pb.ReceiveServiceClient
}

func NewControllerService(receiveServiceClient pb.ReceiveServiceClient) (*controllerService, error) {
	if receiveServiceClient == nil {
		return nil, errors.New("ReceiveServiceClient must not be nil")
	}

	return &controllerService{
		receiveServiceClient: receiveServiceClient,
	}, nil
}

func (svc *controllerService) PushBlob(ctx context.Context, request *dto.ControllerPushBlobRequest) (*dto.ControllerPushBlobResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Push blob request")

	if _, err := svc.receiveServiceClient.PushData(ctx, &pb.ModuleControllerData{
		Receiver: request.ReceiverModuleID,
		Sender: &pb.ModuleIdentifier{
			Id: request.SourceModuleID,
		},
		Data: request.Blob,
	}); err != nil {
		log.Error().Err(err).Msg("failed to send data to controller")
		return nil, fmt.Errorf("failed to send data to controller: %v", err)
	}
	return &dto.ControllerPushBlobResponse{}, nil
}
