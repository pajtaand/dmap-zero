package grpc

import (
	"net"

	"github.com/rs/zerolog/log"

	pb "github.com/andreepyro/dmap-zero/internal/proto"
	"google.golang.org/grpc"
)

type AgentServer struct {
	s   *grpc.Server
	lis net.Listener
}

func NewAgentServer(
	configurationService pb.ConfigurationServiceServer,
	imageService pb.ImageServiceServer,
	moduleService pb.ModuleServiceServer,
	shareService pb.ShareServiceServer,
	listener net.Listener,
) *AgentServer {
	s := grpc.NewServer()
	pb.RegisterConfigurationServiceServer(s, configurationService)
	pb.RegisterImageServiceServer(s, imageService)
	pb.RegisterModuleServiceServer(s, moduleService)
	pb.RegisterShareServiceServer(s, shareService)
	return &AgentServer{
		s:   s,
		lis: listener,
	}
}

func (srv *AgentServer) Run() error {
	log.Info().Msgf("Listening on %v", srv.lis.Addr())
	return srv.s.Serve(srv.lis)
}

func (srv *AgentServer) Stop() error {
	log.Info().Msg("Stopping server")
	srv.s.Stop()
	return nil
}
