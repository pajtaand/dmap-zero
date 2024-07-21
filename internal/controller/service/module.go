package service

import (
	"context"
	"errors"
	"fmt"

	errs "github.com/andreepyro/dmap-zero/internal/common/errors"
	"github.com/andreepyro/dmap-zero/internal/controller/dto"
	"github.com/andreepyro/dmap-zero/internal/controller/manager"
	pb "github.com/andreepyro/dmap-zero/internal/proto"
	"github.com/rs/zerolog"
)

type moduleService struct {
	moduleManager *manager.ModuleManager
	imageManager  *manager.ImageManager
	agentManager  *manager.AgentManager
}

func NewModuleService(moduleManager *manager.ModuleManager, imageManager *manager.ImageManager, agentManager *manager.AgentManager) (*moduleService, error) {
	if moduleManager == nil {
		return nil, errors.New("ModuleManager must not be nil")
	}
	if imageManager == nil {
		return nil, errors.New("ImageManager must not be nil")
	}
	if agentManager == nil {
		return nil, errors.New("AgentManager must not be nil")
	}

	return &moduleService{
		moduleManager: moduleManager,
		imageManager:  imageManager,
		agentManager:  agentManager,
	}, nil
}

func (svc *moduleService) CreateModule(ctx context.Context, request *dto.CreateModuleRequest) (*dto.CreateModuleResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Create module request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	if !svc.imageManager.ImageExists(request.Image) {
		return nil, fmt.Errorf("failed to find image: %s", request.Image)
	}

	moduleID := svc.moduleManager.AddModule(request.Name, request.Image, request.Configuration)
	return &dto.CreateModuleResponse{
		ID: moduleID,
	}, nil
}

func (svc *moduleService) GetModule(ctx context.Context, request *dto.GetModuleRequest) (*dto.GetModuleResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Get module request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	module, err := svc.moduleManager.GetModule(request.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get module: %v", err)
	}

	return &dto.GetModuleResponse{
		Name:          module.GetName(),
		Image:         module.GetImage(),
		Configuration: module.GetConfiguration(),
		IsRunning:     module.IsRunning(),
	}, nil
}

func (svc *moduleService) ListModules(ctx context.Context, request *dto.ListModulesRequest) (*dto.ListModulesResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("List modules request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	modules := make([]*dto.ListModulesResponseModule, 0)
	for _, module := range svc.moduleManager.ListModules() {
		modules = append(modules, &dto.ListModulesResponseModule{
			ID:            module.GetID(),
			Name:          module.GetName(),
			Image:         module.GetImage(),
			Configuration: module.GetConfiguration(),
			IsRunning:     module.IsRunning(),
		})
	}
	return &dto.ListModulesResponse{
		Modules: modules,
	}, nil
}

func (svc *moduleService) UpdateModule(ctx context.Context, request *dto.UpdateModuleRequest) (*dto.UpdateModuleResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Update modules request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	if !svc.imageManager.ImageExists(request.Image) {
		return nil, fmt.Errorf("failed to find image: %s", request.Image)
	}

	module, err := svc.moduleManager.GetModule(request.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get module: %v", err)
	}
	module.SetName(request.Name)
	module.SetImage(request.Image)
	module.SetConfiguration(request.Configuration)
	return &dto.UpdateModuleResponse{}, nil
}

func (svc *moduleService) DeleteModule(ctx context.Context, request *dto.DeleteModuleRequest) (*dto.DeleteModuleResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Delete module request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	module, err := svc.moduleManager.GetModule(request.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get module: %v", err)
	}
	if module.IsRunning() {
		return nil, errs.ErrNotAllowed
	}

	if err := svc.moduleManager.RemoveModule(request.ID); err != nil {
		return nil, fmt.Errorf("failed to remove module: %v", err)
	}
	return &dto.DeleteModuleResponse{}, nil
}

func (svc *moduleService) StartModule(ctx context.Context, request *dto.StartModuleRequest) (*dto.StartModuleResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Start modules request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	module, err := svc.moduleManager.GetModule(request.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get module: %v", err)
	}

	for _, agent := range svc.agentManager.ListAgents() {
		agentID := agent.GetID()
		moduleID := module.GetID()
		moduleCfg := module.GetConfiguration()
		imageID := module.GetImage()

		c := agent.GetModuleServiceClient()
		if c == nil {
			continue
		}
		log.Info().Msgf("Starting module: agentID=%s, moduleID=%s, moduleCfg=%v, imageID=%s", agentID, moduleID, moduleCfg, imageID)

		if _, err := c.StartModule(ctx, &pb.ModuleConfiguration{
			Module: &pb.ModuleIdentifier{
				Id: moduleID,
			},
			Image: &pb.ImageIdentifier{
				Id: imageID,
			},
			Env: moduleCfg,
		}); err != nil {
			log.Info().Msgf("could not get response: %v", err)
			continue
		}
		log.Info().Msgf("Module start response: agentID=%s, moduleID=%s, moduleCfg=%v, imageID=%s", agentID, moduleID, moduleCfg, imageID)
	}
	module.SetRunning(true)

	return &dto.StartModuleResponse{}, nil
}

func (svc *moduleService) StopModule(ctx context.Context, request *dto.StopModuleRequest) (*dto.StopModuleResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Stop module request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	module, err := svc.moduleManager.GetModule(request.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get module: %v", err)
	}

	for _, agent := range svc.agentManager.ListAgents() {
		agentID := agent.GetID()
		moduleID := module.GetID()

		c := agent.GetModuleServiceClient()
		if c == nil {
			continue
		}
		log.Info().Msgf("Stopping module: agentID=%s, moduleID=%s", agentID, moduleID)

		if _, err := c.StopModule(ctx, &pb.ModuleIdentifier{
			Id: moduleID,
		}); err != nil {
			log.Info().Msgf("could not get response: %v", err)
			continue
		}

		log.Info().Msgf("Module stop response: agentID=%s, moduleID=%s", agentID, moduleID)
	}
	module.SetRunning(false)

	return &dto.StopModuleResponse{}, nil
}

func (svc *moduleService) SendData(ctx context.Context, request *dto.SendDataRequest) (*dto.SendDataResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Send data request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	if !svc.moduleManager.ModuleExists(request.ModuleID) {
		return nil, errs.ErrNotFound
	}

	for _, agent := range svc.agentManager.ListAgents() {
		agentID := agent.GetID()

		c := agent.GetShareServiceClient()
		if c == nil {
			continue
		}
		log.Info().Msgf("Sending data: agentID=%s, moduleID=%s", agentID, request.ModuleID)

		if _, err := c.PushData(ctx, &pb.ShareData{
			Receiver: &pb.ModuleIdentifier{
				Id: request.ModuleID,
			},
			Data: request.Data,
		}); err != nil {
			log.Info().Msgf("could not get response: %v", err)
			continue
		}

		log.Info().Msgf("Module data response: agentID=%s, moduleID=%s", agentID, request.ModuleID)
	}

	return &dto.SendDataResponse{}, nil
}
