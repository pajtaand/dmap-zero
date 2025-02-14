package rest

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pajtaand/dmap-zero/internal/agent/rest/handler"
	m "github.com/pajtaand/dmap-zero/internal/common/middleware"
)

type RESTServer struct {
	r      chi.Router
	server *http.Server
}

func NewRESTServer(
	authenticator m.Authenticator,
	endpointService handler.EndpointService,
	controllerService handler.ControllerService,
	webhookService handler.WebhookService,
) *RESTServer {
	baseAuthMiddleware := m.BasicAuth("api", authenticator)
	endpointHandler := handler.NewEndpointHandler(endpointService)
	controllerHandler := handler.NewControllerHandler(controllerService)
	webhookHandler := handler.NewWebhookHandler(webhookService)

	r := chi.NewRouter()
	srv := &RESTServer{
		r: r,
	}
	srv.addHandlers(
		endpointHandler,
		controllerHandler,
		webhookHandler,
		baseAuthMiddleware,
	)
	return srv
}

func (srv *RESTServer) Run(addr string, cert *tls.Certificate) error {
	log.Info().Msgf("Listening on https://%s/", addr)
	srv.server = &http.Server{
		Addr:    addr,
		Handler: srv.r,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*cert},
		},
	}
	return srv.server.ListenAndServeTLS("", "")
}

func (srv *RESTServer) Stop(ctx context.Context) error {
	if srv.server == nil {
		return fmt.Errorf("failed to stop server: not running")
	}

	log.Info().Msg("Stopping server")
	if err := srv.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to stop server: %v", err)
	}

	return nil
}

func (srv *RESTServer) addHandlers(
	endpointHandler EndpointHandler,
	controllerHandler ControllerHandler,
	webhookHandler WebhookHandler,
	authMiddleware func(next http.Handler) http.Handler,
) {
	srv.r.Use(middleware.RequestID)
	srv.r.Use(m.Logger)
	srv.r.Use(middleware.Recoverer)
	srv.r.Use(middleware.URLFormat)

	srv.r.Route("/api/v1", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Route("/endpoint", func(r chi.Router) {
			r.Get("/", endpointHandler.ListEndpoints)
			r.Post("/push", endpointHandler.PushBlobToEndpoint)
		})
		r.Route("/controller", func(r chi.Router) {
			r.Post("/push", controllerHandler.PushBlobToController)
		})
		r.Route("/webhook", func(r chi.Router) {
			r.Get("/", webhookHandler.ListWebhooks)
			r.Post("/", webhookHandler.RegisterWebhook)
			r.Delete("/", webhookHandler.DeleteWebhook)
		})
	})
}
