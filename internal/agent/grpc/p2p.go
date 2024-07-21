package grpc

import (
	"net"

	"github.com/rs/zerolog/log"

	pb "github.com/andreepyro/dmap-zero/internal/proto"
	"google.golang.org/grpc"
)

type P2PServer struct {
	s   *grpc.Server
	lis net.Listener
}

func NewP2PServer(
	pingService pb.PingServiceServer,
	shareService pb.ShareServiceServer,
	listener net.Listener,
) *P2PServer {
	s := grpc.NewServer()
	pb.RegisterPingServiceServer(s, pingService)
	pb.RegisterShareServiceServer(s, shareService)
	return &P2PServer{
		s:   s,
		lis: listener,
	}
}

func (srv *P2PServer) Run() error {
	log.Info().Msgf("Listening on %v", srv.lis.Addr())
	return srv.s.Serve(srv.lis)
}

func (srv *P2PServer) Stop() error {
	log.Info().Msg("Stopping server")
	srv.s.Stop()
	return nil
}
