package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/andreepyro/dmap-zero/internal/agent/dto"
	"github.com/andreepyro/dmap-zero/internal/agent/manager"
	errs "github.com/andreepyro/dmap-zero/internal/common/errors"
	"github.com/rs/zerolog"
)

type webhookService struct {
	webhookManager *manager.WebhookManager
	moduleManager  *manager.ModuleManager
}

func NewWebhookService(webhookManager *manager.WebhookManager, moduleManager *manager.ModuleManager) (*webhookService, error) {
	if webhookManager == nil {
		return nil, errors.New("WebhookManager must not be nil")
	}
	if moduleManager == nil {
		return nil, errors.New("ModuleManager must not be nil")
	}

	return &webhookService{
		webhookManager: webhookManager,
		moduleManager:  moduleManager,
	}, nil
}

func (svc *webhookService) ListWebhooks(ctx context.Context, request *dto.ListWebhooksRequest) (*dto.ListWebhooksResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("List webhooks request")

	whs, err := svc.webhookManager.ListWebhooks(request.SourceModuleID)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %v", err)
	}

	webhooks := make([]*dto.ListWebhooksResponseWebhook, 0, len(whs))
	for _, webhook := range whs {
		webhooks = append(webhooks, &dto.ListWebhooksResponseWebhook{
			ID:      webhook.GetID(),
			URLPath: webhook.GetURLPath(),
			Event:   webhook.GetEvent(),
		})
	}

	return &dto.ListWebhooksResponse{
		Webhooks: webhooks,
	}, nil
}

func (svc *webhookService) RegisterWebhook(ctx context.Context, request *dto.RegisterWebhookRequest) (*dto.RegisterWebhookResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Register webhook request")

	module, err := svc.moduleManager.GetModule(request.SourceModuleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get module: %v", err)
	}
	port := module.GetGivenPort()

	webhookID, err := svc.webhookManager.AddWebhook(request.SourceModuleID, request.URLPath, port, request.Event)
	if err != nil {
		return nil, fmt.Errorf("failed to register webhook: %v", err)
	}

	return &dto.RegisterWebhookResponse{
		ID: webhookID,
	}, nil
}

func (svc *webhookService) DeleteWebhook(ctx context.Context, request *dto.DeleteWebhookRequest) (*dto.DeleteWebhookResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Delete webhook request")

	exists, err := svc.webhookManager.WebhookExists(request.SourceModuleID, request.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %v", err)
	}

	if !exists {
		return nil, errs.ErrNotFound
	}

	if err := svc.webhookManager.RemoveWebhook(request.SourceModuleID, request.ID); err != nil {
		return nil, fmt.Errorf("failed to delete webhook: %v", err)
	}

	return &dto.DeleteWebhookResponse{}, nil
}
