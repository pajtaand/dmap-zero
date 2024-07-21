package handler

import (
	"net/http"

	"github.com/andreepyro/dmap-zero/internal/agent/dto"
	"github.com/andreepyro/dmap-zero/internal/agent/rest/models"
	"github.com/andreepyro/dmap-zero/internal/common/utils"
	"github.com/rs/zerolog"
)

type controllerHandler struct {
	service ControllerService
}

func NewControllerHandler(service ControllerService) *controllerHandler {
	return &controllerHandler{
		service: service,
	}
}

func (h *controllerHandler) PushBlobToController(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	ctx := r.Context()
	user, ok := utils.GetUser(ctx)
	if !ok {
		panic("user not present in context")
	}

	req := &models.ControllerPushRequest{}
	if err := req.FromHttpRequest(r); err != nil {
		log.Error().Err(err).Msg("")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	if _, err := h.service.PushBlob(ctx, &dto.ControllerPushBlobRequest{
		SourceModuleID:   user,
		ReceiverModuleID: req.ReceiverID,
		Blob:             req.Blob,
	}); err != nil {
		log.Error().Err(err).Msg("")
		utils.WriteErrorResponse(w, http.StatusInternalServerError, nil)
		return
	}

	utils.WriteResponse(w, http.StatusOK, nil)
}
