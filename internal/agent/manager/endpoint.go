package manager

import (
	"context"
	"errors"
	"fmt"

	"github.com/andreepyro/dmap-zero/internal/common/constants"
	"github.com/andreepyro/dmap-zero/internal/common/wrapper"
	pb "github.com/andreepyro/dmap-zero/internal/proto"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type EndpointManager struct {
	openZitiWrapper *wrapper.OpenZitiClientWrapper
}

func NewEndpointManager(openZitiWrapper *wrapper.OpenZitiClientWrapper) (*EndpointManager, error) {
	log.Debug().Msg("Creating new EndpointManager")

	if openZitiWrapper == nil {
		return nil, errors.New("OpenZitiClientWrapper must not be nil")
	}
	return &EndpointManager{
		openZitiWrapper: openZitiWrapper,
	}, nil
}

func (mgr *EndpointManager) ListEndpoints() ([]string, error) {
	log.Info().Msg("Listing all endpoints")

	identityIDs, err := mgr.openZitiWrapper.GetServiceTerminators(constants.OpenZitiServiceP2P)
	if err != nil {
		return nil, fmt.Errorf("failed to list endpoints: %v", err)
	}
	return identityIDs, nil
}

func (mgr *EndpointManager) SendData(ctx context.Context, identityID, moduleID string, data []byte) error {
	log.Info().Msgf("Sending data to endpoint: identityID=%s, moduleID=%s", identityID, moduleID)

	conn, err := grpc.NewClient(
		fmt.Sprintf("passthrough:///%s", constants.OpenZitiServiceP2P),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(mgr.openZitiWrapper.GetContextDialerWithOptions(&ziti.DialOptions{
			Identity: identityID,
		})),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to other agent: %v", err)
	}
	defer conn.Close()

	c := pb.NewShareServiceClient(conn)
	if _, err = c.PushData(ctx, &pb.ShareData{
		Receiver: &pb.ModuleIdentifier{
			Id: moduleID,
		},
		Data: data,
	}); err != nil {
		return fmt.Errorf("failed to send data to other agent: %v", err)
	}

	return nil
}
