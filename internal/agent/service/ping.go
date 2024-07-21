package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/andreepyro/dmap-zero/internal/common/utils"
	pb "github.com/andreepyro/dmap-zero/internal/proto"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
)

type pingService struct {
	pb.UnimplementedPingServiceServer
}

func NewPingService() (pb.PingServiceServer, error) {
	return &pingService{}, nil
}

func (svc *pingService) Ping(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Ping request")

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

	return &emptypb.Empty{}, nil
}
