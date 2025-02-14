package grpc

import (
	"net"

	"github.com/rs/zerolog/log"

	pb "github.com/pajtaand/dmap-zero/internal/proto"
	"google.golang.org/grpc"
)

type RPCServer struct {
	s   *grpc.Server
	lis net.Listener
}

func NewRPCServer(
	phonehomeService pb.PhonehomeServiceServer,
	setupService pb.SetupServiceServer,
	receiveService pb.ReceiveServiceServer,
	listener net.Listener,
) *RPCServer {
	s := grpc.NewServer()
	pb.RegisterPhonehomeServiceServer(s, phonehomeService)
	pb.RegisterSetupServiceServer(s, setupService)
	pb.RegisterReceiveServiceServer(s, receiveService)
	return &RPCServer{
		s:   s,
		lis: listener,
	}
}

func (srv *RPCServer) Run() error {
	log.Info().Msgf("Listening on %v", srv.lis.Addr())
	return srv.s.Serve(srv.lis)
}

func (srv *RPCServer) Stop() error {
	log.Info().Msg("Stopping server")
	srv.s.Stop()
	return nil
}
