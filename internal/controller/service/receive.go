package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/andreepyro/dmap-zero/internal/common/utils"
	"github.com/andreepyro/dmap-zero/internal/controller/manager"
	pb "github.com/andreepyro/dmap-zero/internal/proto"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
)

type receiveService struct {
	pb.UnimplementedReceiveServiceServer

	webhookManager *manager.WebhookManager
}

func NewReceiveService(webhookManager *manager.WebhookManager) (pb.ReceiveServiceServer, error) {
	if webhookManager == nil {
		return nil, errors.New("WebhookManager must not be nil")
	}

	return &receiveService{
		webhookManager: webhookManager,
	}, nil
}

func (svc *receiveService) PushData(ctx context.Context, data *pb.ModuleControllerData) (*emptypb.Empty, error) {
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

	log.Info().Msgf("Received module message: agentID=%s, moduleID=%s, Receiver=%s", sourceIdentity, data.Sender, data.Receiver)

	if err := svc.webhookManager.SendData(data.Sender.Id, data.Receiver, data.Data); err != nil {
		err := fmt.Errorf("failed to push data to module: %v", err)
		log.Error().Err(err).Msg("")
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
