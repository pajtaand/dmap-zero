package handler

import (
	"io"
	"net/http"

	"github.com/andreepyro/dmap-zero/internal/agent/dto"
	"github.com/andreepyro/dmap-zero/internal/agent/rest/models"
	"github.com/andreepyro/dmap-zero/internal/common/utils"
	"github.com/rs/zerolog"
)

type endpointHandler struct {
	service EndpointService
}

func NewEndpointHandler(service EndpointService) *endpointHandler {
	return &endpointHandler{service: service}
}

func (h *endpointHandler) ListEndpoints(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	ctx := r.Context()
	user, ok := utils.GetUser(ctx)
	if !ok {
		panic("user not present in context")
	}

	endpoints, err := h.service.ListEndpoints(r.Context(), &dto.ListEndpointsRequest{
		SourceModuleID: user,
	})
	if err != nil {
		log.Error().Err(err).Msg("")
		utils.WriteErrorResponse(w, http.StatusInternalServerError, nil)
		return
	}

	resp := []*models.Endpoint{}
	for _, endpoint := range endpoints.Endpoints {
		resp = append(resp, &models.Endpoint{
			ID: endpoint.ID,
		})
	}
	utils.WriteResponse(w, http.StatusOK, &resp)
}

func (h *endpointHandler) PushBlobToEndpoint(w http.ResponseWriter, r *http.Request) {
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

	blob, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	if _, err := h.service.PushBlob(r.Context(), &dto.EndpointPushBlobRequest{
		SourceModuleID:     user,
		ReceiverIdentityID: ID,
		ReceiverModuleID:   user,
		Blob:               blob,
	}); err != nil {
		log.Error().Err(err).Msg("")
		utils.WriteErrorResponse(w, http.StatusInternalServerError, nil)
		return
	}

	utils.WriteResponse(w, http.StatusOK, nil)
}
