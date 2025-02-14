package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/pajtaand/dmap-zero/internal/agent/dto"
	"github.com/pajtaand/dmap-zero/internal/agent/manager"
	"github.com/pajtaand/dmap-zero/internal/common/constants"
	"github.com/pajtaand/dmap-zero/internal/common/utils"
	pb "github.com/pajtaand/dmap-zero/internal/proto"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
)

type shareService struct {
	pb.UnimplementedShareServiceServer

	webhookManager *manager.WebhookManager
}

func NewShareService(webhookManager *manager.WebhookManager) (pb.ShareServiceServer, error) {
	if webhookManager == nil {
		return nil, errors.New("WebhookManager must not be nil")
	}

	return &shareService{
		webhookManager: webhookManager,
	}, nil
}

func (svc *shareService) PushData(ctx context.Context, data *pb.ShareData) (*emptypb.Empty, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Push data request")

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

	var eventType dto.WebhookEvent
	if sourceIdentity == constants.OpenZitiIdentityController {
		eventType = dto.EventControllerData
	} else {
		eventType = dto.EventEndpointData
	}

	if err := svc.webhookManager.SendData(sourceIdentity, data.Receiver.Id, eventType, data.Data); err != nil {
		err := fmt.Errorf("failed to push data to module: %v", err)
		log.Error().Err(err).Msg("")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
