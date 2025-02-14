package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	errs "github.com/pajtaand/dmap-zero/internal/common/errors"
	"github.com/pajtaand/dmap-zero/internal/common/utils"
	"github.com/pajtaand/dmap-zero/internal/controller/dto"
	"github.com/pajtaand/dmap-zero/internal/controller/rest/models"
	"github.com/rs/zerolog"
)

type enrollmentHandler struct {
	service EnrollmentService
}

func NewEnrollmentHandler(service EnrollmentService) *enrollmentHandler {
	return &enrollmentHandler{
		service: service,
	}
}

func (h *enrollmentHandler) CreateEnrollment(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		log.Info().Msg("agentID is empty")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	enroll, err := h.service.CreateEnrollment(r.Context(), &dto.CreateEnrollmentRequest{
		ID: agentID,
	})
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			log.Error().Err(err).Msg("")
			utils.WriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("agent with id '%s' doesn't exists", agentID))
			return
		}
		panic(err)
	}

	utils.WriteResponse(w, http.StatusOK, models.CreateEnrollmentResponse{
		JWT:       enroll.JWT,
		ExpiresAt: enroll.ExpiresAt,
	})
}

func (h *enrollmentHandler) GetEnrollment(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		log.Info().Msg("agentID is empty")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	enroll, err := h.service.GetEnrollment(r.Context(), &dto.GetEnrollmentRequest{
		ID: agentID,
	})
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			log.Error().Err(err).Msg("")
			utils.WriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("enrollment not found"))
			return
		}
		panic(err)
	}

	utils.WriteResponse(w, http.StatusOK, models.GetEnrollmentResponse{
		JWT:       enroll.JWT,
		ExpiresAt: enroll.ExpiresAt,
	})
}

func (h *enrollmentHandler) DeleteEnrollment(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		log.Info().Msg("agentID is empty")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	if _, err := h.service.DeleteEnrollment(r.Context(), &dto.DeleteEnrollmentRequest{
		ID: agentID,
	}); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			log.Error().Err(err).Msg("")
			utils.WriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("agent with id '%s' doesn't exists", agentID))
			return
		}
		panic(err)
	}

	utils.WriteResponse(w, http.StatusOK, nil)
}
