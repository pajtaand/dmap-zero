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

type agentHandler struct {
	service AgentService
}

func NewAgentHandler(service AgentService) *agentHandler {
	return &agentHandler{
		service: service,
	}
}

func (h *agentHandler) CreateAgent(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	req := &models.CreateAgentRequest{}
	if err := req.FromHttpRequest(r); err != nil {
		log.Error().Err(err).Msg("")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	agent, err := h.service.CreateAgent(r.Context(), &dto.CreateAgentRequest{
		Name:          req.Name,
		Configuration: req.Configuration,
	})
	if err != nil {
		panic(err)
	}

	utils.WriteResponse(w, http.StatusOK, models.CreateAgentResponse{
		ID: agent.ID,
	})
}

func (h *agentHandler) GetAgent(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		log.Info().Msg("agentID is empty")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	agent, err := h.service.GetAgent(r.Context(), &dto.GetAgentRequest{
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

	utils.WriteResponse(w, http.StatusOK, models.GetAgentResponse{
		Name:           agent.Name,
		Configuration:  agent.Configuration,
		IsEnrolled:     agent.IsEnrolled,
		IsOnline:       agent.IsOnline,
		PresentImages:  agent.PresentImages,
		PresentModules: agent.PresentModules,
	})
}

func (h *agentHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := h.service.ListAgents(r.Context(), &dto.ListAgentsRequest{})
	if err != nil {
		panic(err)
	}

	agentList := []models.ListAgentsResponseAgent{}
	for _, agent := range agents.Agents {
		agentList = append(agentList, models.ListAgentsResponseAgent{
			ID:             agent.ID,
			Name:           agent.Name,
			Configuration:  agent.Configuration,
			IsEnrolled:     agent.IsEnrolled,
			IsOnline:       agent.IsOnline,
			PresentImages:  agent.PresentImages,
			PresentModules: agent.PresentModules,
		})
	}
	utils.WriteResponse(w, http.StatusOK, models.ListAgentsResponse{
		Agents: agentList,
	})
}

func (h *agentHandler) UpdateAgent(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		log.Info().Msg("agentID is empty")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	req := &models.UpdateAgentRequest{}
	if err := req.FromHttpRequest(r); err != nil {
		log.Error().Err(err).Msg("")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	if _, err := h.service.UpdateAgent(r.Context(), &dto.UpdateAgentRequest{
		ID:            agentID,
		Name:          req.Name,
		Configuration: req.Configuration,
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

func (h *agentHandler) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	agentID := chi.URLParam(r, "agentID")
	if agentID == "" {
		log.Info().Msg("agentID is empty")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	if _, err := h.service.DeleteAgent(r.Context(), &dto.DeleteAgentRequest{
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
