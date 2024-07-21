package rest

import (
	"context"
	"fmt"
	"net/http"

	m "github.com/andreepyro/dmap-zero/internal/common/middleware"
	"github.com/andreepyro/dmap-zero/internal/controller/rest/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/docgen"
	"github.com/rs/zerolog/log"
)

type RESTServer struct {
	r      chi.Router
	server *http.Server
}

func NewRESTServer(
	authenticator m.Authenticator,
	agentService handler.AgentService,
	moduleService handler.ModuleService,
	imageService handler.ImageService,
	webhookService handler.WebhookService,
	enrollmentService handler.EnrollmentService,
) *RESTServer {
	baseAuthMiddleware := m.BasicAuth("api", authenticator)
	webAppHandler := handler.NewWebAppHandler()
	agentHandler := handler.NewAgentHandler(agentService)
	moduleHandler := handler.NewModuleHandler(moduleService)
	imageHandler := handler.NewImageHandler(imageService)
	webhookHandler := handler.NewWebhookHandler(webhookService)
	enrollmentHandler := handler.NewEnrollmentHandler(enrollmentService)

	r := chi.NewRouter()
	srv := &RESTServer{
		r: r,
	}
	srv.addHandlers(
		webAppHandler,
		agentHandler,
		moduleHandler,
		imageHandler,
		webhookHandler,
		enrollmentHandler,
		baseAuthMiddleware,
	)
	return srv
}

func (srv *RESTServer) Run(addr, certFile, keyFile string) error {
	log.Info().Msgf("Listening on https://%s/", addr)
	srv.server = &http.Server{Addr: addr, Handler: srv.r}
	return srv.server.ListenAndServeTLS(certFile, keyFile)
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

func (srv *RESTServer) GetDocs() string {
	// run with GOPATH=$(go env GOPATH)
	return docgen.JSONRoutesDoc(srv.r)
}

func (srv *RESTServer) addHandlers(
	webAppHandler WebAppHandler,
	agentHandler AgentHandler,
	moduleHandler ModuleHandler,
	imageHandler ImageHandler,
	webhookHandler WebhookHandler,
	enrollmentHandler EnrollmentHandler,
	authMiddleware func(next http.Handler) http.Handler,
) {
	srv.r.Use(middleware.RequestID)
	srv.r.Use(m.Logger)
	srv.r.Use(middleware.Recoverer)
	srv.r.Use(middleware.URLFormat)
	srv.r.Use(m.Metrics)

	srv.r.Get("/", webAppHandler.GetWebApplication)
	srv.r.Route("/src", func(r chi.Router) {
		r.Get("/favicon", webAppHandler.GetFavicon)
		r.Get("/icon", webAppHandler.GetIcon)
		r.Get("/css", webAppHandler.GetCSS)
		r.Get("/js", webAppHandler.GetJS)
	})
	srv.r.Route("/api/v1", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Route("/agent", func(r chi.Router) {
			r.Post("/", agentHandler.CreateAgent)
			r.Get("/", agentHandler.ListAgents)
			r.Route("/{agentID}", func(r chi.Router) {
				r.Get("/", agentHandler.GetAgent)
				r.Patch("/", agentHandler.UpdateAgent)
				r.Delete("/", agentHandler.DeleteAgent)
				r.Route("/enrollment", func(r chi.Router) {
					r.Get("/", enrollmentHandler.GetEnrollment)
					r.Post("/", enrollmentHandler.CreateEnrollment)
					r.Delete("/", enrollmentHandler.DeleteEnrollment)
				})
			})
		})
		r.Route("/module", func(r chi.Router) {
			r.Post("/", moduleHandler.CreateModule)
			r.Get("/", moduleHandler.ListModules)
			r.Route("/{moduleID}", func(r chi.Router) {
				r.Get("/", moduleHandler.GetModule)
				r.Patch("/", moduleHandler.UpdateModule)
				r.Delete("/", moduleHandler.DeleteModule)
				r.Post("/start", moduleHandler.StartModule)
				r.Post("/stop", moduleHandler.StopModule)
				r.Post("/send", moduleHandler.SendData)
			})
		})
		r.Route("/image", func(r chi.Router) {
			r.Post("/", imageHandler.UploadImage)
			r.Get("/", imageHandler.ListImages)
			r.Route("/{imageID}", func(r chi.Router) {
				r.Get("/", imageHandler.GetImage)
				r.Delete("/", imageHandler.DeleteImage)
			})
		})
		r.Route("/webhook", func(r chi.Router) {
			r.Get("/", webhookHandler.ListWebhooks)
			r.Post("/", webhookHandler.RegisterWebhook)
			r.Delete("/", webhookHandler.DeleteWebhook)
		})
	})
}
