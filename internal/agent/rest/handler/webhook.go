package handler

import (
	"net/http"

	"github.com/pajtaand/dmap-zero/internal/agent/dto"
	"github.com/pajtaand/dmap-zero/internal/agent/rest/models"
	"github.com/pajtaand/dmap-zero/internal/common/utils"
	"github.com/rs/zerolog"
)

type webhookHandler struct {
	service WebhookService
}

func NewWebhookHandler(service WebhookService) *webhookHandler {
	return &webhookHandler{
		service: service,
	}
}

func (h *webhookHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	ctx := r.Context()
	user, ok := utils.GetUser(ctx)
	if !ok {
		panic("user not present in context")
	}

	webhooks, err := h.service.ListWebhooks(r.Context(), &dto.ListWebhooksRequest{
		SourceModuleID: user,
	})
	if err != nil {
		log.Error().Err(err).Msg("")
		utils.WriteErrorResponse(w, http.StatusInternalServerError, nil)
		return
	}

	resp := []*models.Webhook{}
	for _, webhook := range webhooks.Webhooks {
		resp = append(resp, &models.Webhook{
			ID:      webhook.ID,
			URLPath: webhook.URLPath,
			Event:   string(webhook.Event),
		})
	}
	utils.WriteResponse(w, http.StatusOK, &resp)
}

func (h *webhookHandler) RegisterWebhook(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	ctx := r.Context()
	user, ok := utils.GetUser(ctx)
	if !ok {
		panic("user not present in context")
	}

	req := &models.WebhookRegistrationRequest{}
	if err := req.FromHttpRequest(r); err != nil {
		log.Error().Err(err).Msg("")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	event, err := dto.ParseWebhookEvent(req.Event)
	if err != nil {
		log.Error().Err(err).Msg("")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	resp, err := h.service.RegisterWebhook(r.Context(), &dto.RegisterWebhookRequest{
		SourceModuleID: user,
		URLPath:        req.URLPath,
		Event:          event,
	})
	if err != nil {
		log.Error().Err(err).Msg("")
		utils.WriteErrorResponse(w, http.StatusInternalServerError, nil)
		return
	}

	utils.WriteResponse(w, http.StatusCreated, &models.WebhookRegistrationResponse{
		ID: resp.ID,
	})
}

func (h *webhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	ctx := r.Context()
	user, ok := utils.GetUser(ctx)
	if !ok {
		panic("user not present in context")
	}

	ID := r.URL.Query().Get("id")
	if ID == "" {
		log.Info().Msg("Missing query parameter: ID")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	if _, err := h.service.DeleteWebhook(r.Context(), &dto.DeleteWebhookRequest{
		SourceModuleID: user,
		ID:             ID,
	}); err != nil {
		log.Error().Err(err).Msg("")
		utils.WriteErrorResponse(w, http.StatusInternalServerError, nil)
		return
	}

	utils.WriteResponse(w, http.StatusNoContent, nil)
}
