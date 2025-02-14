package handler

import (
	"context"

	"github.com/pajtaand/dmap-zero/internal/controller/dto"
)

type AgentService interface {
	CreateAgent(ctx context.Context, req *dto.CreateAgentRequest) (*dto.CreateAgentResponse, error)
	GetAgent(ctx context.Context, req *dto.GetAgentRequest) (*dto.GetAgentResponse, error)
	ListAgents(ctx context.Context, req *dto.ListAgentsRequest) (*dto.ListAgentsResponse, error)
	UpdateAgent(ctx context.Context, req *dto.UpdateAgentRequest) (*dto.UpdateAgentResponse, error)
	DeleteAgent(ctx context.Context, req *dto.DeleteAgentRequest) (*dto.DeleteAgentResponse, error)
}

type WebhookService interface {
	ListWebhooks(ctx context.Context, req *dto.ListWebhooksRequest) (*dto.ListWebhooksResponse, error)
	RegisterWebhook(ctx context.Context, req *dto.RegisterWebhookRequest) (*dto.RegisterWebhookResponse, error)
	DeleteWebhook(ctx context.Context, req *dto.DeleteWebhookRequest) (*dto.DeleteWebhookResponse, error)
}

type EnrollmentService interface {
	CreateEnrollment(ctx context.Context, req *dto.CreateEnrollmentRequest) (*dto.CreateEnrollmentResponse, error)
	GetEnrollment(ctx context.Context, req *dto.GetEnrollmentRequest) (*dto.GetEnrollmentResponse, error)
	DeleteEnrollment(ctx context.Context, req *dto.DeleteEnrollmentRequest) (*dto.DeleteEnrollmentResponse, error)
}

type ImageService interface {
	UploadImage(ctx context.Context, req *dto.UploadImageRequest) (*dto.UploadImageResponse, error)
	GetImage(ctx context.Context, req *dto.GetImageRequest) (*dto.GetImageResponse, error)
	ListImages(ctx context.Context, req *dto.ListImagesRequest) (*dto.ListImagesResponse, error)
	DeleteImage(ctx context.Context, req *dto.DeleteImageRequest) (*dto.DeleteImageResponse, error)
}

type ModuleService interface {
	CreateModule(ctx context.Context, req *dto.CreateModuleRequest) (*dto.CreateModuleResponse, error)
	GetModule(ctx context.Context, req *dto.GetModuleRequest) (*dto.GetModuleResponse, error)
	ListModules(ctx context.Context, req *dto.ListModulesRequest) (*dto.ListModulesResponse, error)
	UpdateModule(ctx context.Context, req *dto.UpdateModuleRequest) (*dto.UpdateModuleResponse, error)
	DeleteModule(ctx context.Context, req *dto.DeleteModuleRequest) (*dto.DeleteModuleResponse, error)
	StartModule(ctx context.Context, req *dto.StartModuleRequest) (*dto.StartModuleResponse, error)
	StopModule(ctx context.Context, req *dto.StopModuleRequest) (*dto.StopModuleResponse, error)
	SendData(ctx context.Context, req *dto.SendDataRequest) (*dto.SendDataResponse, error)
}
