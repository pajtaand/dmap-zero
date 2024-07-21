package metrics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

type MetricsServer struct {
	server *http.Server
}

func NewMetricsServer() *MetricsServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Handler: mux,
	}

	return &MetricsServer{
		server: srv,
	}
}

func (srv *MetricsServer) Run(addr, certFile, keyFile string) error {
	srv.server.Addr = addr
	log.Info().Msgf("Listening on https://%s/", addr)
	return srv.server.ListenAndServeTLS(certFile, keyFile)
}

func (srv *MetricsServer) Stop(ctx context.Context) error {
	if srv.server == nil {
		return fmt.Errorf("failed to stop server: not running")
	}

	log.Info().Msg("Stopping server")
	if err := srv.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to stop server: %v", err)
	}

	return nil
}
