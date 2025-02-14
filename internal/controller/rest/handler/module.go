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

type moduleHandler struct {
	service ModuleService
}

func NewModuleHandler(service ModuleService) *moduleHandler {
	return &moduleHandler{
		service: service,
	}
}

func (h *moduleHandler) CreateModule(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	req := &models.CreateModuleRequest{}
	if err := req.FromHttpRequest(r); err != nil {
		log.Error().Err(err).Msg("")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	module, err := h.service.CreateModule(r.Context(), &dto.CreateModuleRequest{
		Name:          req.Name,
		Image:         req.Image,
		Configuration: req.Configuration,
	})
	if err != nil {
		panic(err)
	}

	utils.WriteResponse(w, http.StatusOK, models.CreateModuleResponse{
		ID: module.ID,
	})
}

func (h *moduleHandler) GetModule(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	moduleID := chi.URLParam(r, "moduleID")
	if moduleID == "" {
		log.Info().Msg("moduleID is empty")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	module, err := h.service.GetModule(r.Context(), &dto.GetModuleRequest{
		ID: moduleID,
	})
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			log.Error().Err(err).Msg("")
			utils.WriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("module with id '%s' doesn't exists", moduleID))
			return
		}
		panic(err)
	}

	utils.WriteResponse(w, http.StatusOK, models.GetModuleResponse{
		Name:          module.Name,
		Image:         module.Image,
		Configuration: module.Configuration,
		IsRunning:     module.IsRunning,
	})
}

func (h *moduleHandler) ListModules(w http.ResponseWriter, r *http.Request) {
	modules, err := h.service.ListModules(r.Context(), &dto.ListModulesRequest{})
	if err != nil {
		panic(err)
	}

	moduleList := []models.ListModulesResponseModule{}
	for _, module := range modules.Modules {
		moduleList = append(moduleList, models.ListModulesResponseModule{
			ID:            module.ID,
			Name:          module.Name,
			Image:         module.Image,
			Configuration: module.Configuration,
			IsRunning:     module.IsRunning,
		})
	}
	utils.WriteResponse(w, http.StatusOK, models.ListModulesResponse{
		Modules: moduleList,
	})
}

func (h *moduleHandler) UpdateModule(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	moduleID := chi.URLParam(r, "moduleID")
	if moduleID == "" {
		log.Info().Msg("moduleID is empty")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	req := &models.UpdateModuleRequest{}
	if err := req.FromHttpRequest(r); err != nil {
		log.Error().Err(err).Msg("")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	if _, err := h.service.UpdateModule(r.Context(), &dto.UpdateModuleRequest{
		ID:            moduleID,
		Name:          req.Name,
		Image:         req.Image,
		Configuration: req.Configuration,
	}); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			log.Error().Err(err).Msg("")
			utils.WriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("module with id '%s' doesn't exists", moduleID))
			return
		}
		panic(err)
	}

	utils.WriteResponse(w, http.StatusOK, nil)
}

func (h *moduleHandler) DeleteModule(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	moduleID := chi.URLParam(r, "moduleID")
	if moduleID == "" {
		log.Info().Msg("moduleID is empty")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	if _, err := h.service.DeleteModule(r.Context(), &dto.DeleteModuleRequest{
		ID: moduleID,
	}); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			log.Error().Err(err).Msg("")
			utils.WriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("module with id '%s' doesn't exists", moduleID))
			return
		}
		panic(err)
	}

	utils.WriteResponse(w, http.StatusOK, nil)
}

func (h *moduleHandler) StartModule(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	moduleID := chi.URLParam(r, "moduleID")
	if moduleID == "" {
		log.Info().Msg("moduleID is empty")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	if _, err := h.service.StartModule(r.Context(), &dto.StartModuleRequest{
		ID: moduleID,
	}); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			log.Error().Err(err).Msg("")
			utils.WriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("module with id '%s' doesn't exists", moduleID))
			return
		}
		panic(err)
	}

	utils.WriteResponse(w, http.StatusOK, nil)
}

func (h *moduleHandler) StopModule(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	moduleID := chi.URLParam(r, "moduleID")
	if moduleID == "" {
		log.Info().Msg("moduleID is empty")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	if _, err := h.service.StopModule(r.Context(), &dto.StopModuleRequest{
		ID: moduleID,
	}); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			log.Error().Err(err).Msg("")
			utils.WriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("module with id '%s' doesn't exists", moduleID))
			return
		}
		panic(err)
	}

	utils.WriteResponse(w, http.StatusOK, nil)
}

func (h *moduleHandler) SendData(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	moduleID := chi.URLParam(r, "moduleID")
	if moduleID == "" {
		log.Info().Msg("moduleID is empty")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	req := &models.SendDataRequest{}
	if err := req.FromHttpRequest(r); err != nil {
		log.Error().Err(err).Msg("")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	if _, err := h.service.SendData(r.Context(), &dto.SendDataRequest{
		ModuleID: moduleID,
		Data:     req.Data,
	}); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			log.Error().Err(err).Msg("")
			utils.WriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("module with id '%s' doesn't exists", moduleID))
			return
		}
		panic(err)
	}

	utils.WriteResponse(w, http.StatusOK, nil)
}
