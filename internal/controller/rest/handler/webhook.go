package handler

import (
	"net/http"

	"github.com/andreepyro/dmap-zero/internal/common/utils"
	"github.com/andreepyro/dmap-zero/internal/controller/dto"
	"github.com/andreepyro/dmap-zero/internal/controller/rest/models"
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

	webhooks, err := h.service.ListWebhooks(r.Context(), &dto.ListWebhooksRequest{})
	if err != nil {
		log.Error().Err(err).Msg("")
		utils.WriteErrorResponse(w, http.StatusInternalServerError, nil)
		return
	}

	resp := []*models.Webhook{}
	for _, webhook := range webhooks.Webhooks {
		resp = append(resp, &models.Webhook{
			ID:       webhook.ID,
			ModuleID: webhook.ModuleID,
			URL:      webhook.URL,
		})
	}
	utils.WriteResponse(w, http.StatusOK, &resp)
}

func (h *webhookHandler) RegisterWebhook(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	req := &models.WebhookRegistrationRequest{}
	if err := req.FromHttpRequest(r); err != nil {
		log.Error().Err(err).Msg("")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	resp, err := h.service.RegisterWebhook(r.Context(), &dto.RegisterWebhookRequest{
		ModuleID: req.ModuleID,
		URL:      req.URL,
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

	ID := r.URL.Query().Get("id")
	if ID == "" {
		log.Info().Msg("Missing query parameter: ID")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	if _, err := h.service.DeleteWebhook(r.Context(), &dto.DeleteWebhookRequest{
		ID: ID,
	}); err != nil {
		log.Error().Err(err).Msg("")
		utils.WriteErrorResponse(w, http.StatusInternalServerError, nil)
		return
	}

	utils.WriteResponse(w, http.StatusNoContent, nil)
}
