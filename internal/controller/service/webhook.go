package service

import (
	"context"
	"errors"
	"fmt"

	errs "github.com/pajtaand/dmap-zero/internal/common/errors"
	"github.com/pajtaand/dmap-zero/internal/controller/dto"
	"github.com/pajtaand/dmap-zero/internal/controller/manager"
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

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	whs := svc.webhookManager.ListWebhooks()
	webhooks := make([]*dto.ListWebhooksResponseWebhook, 0, len(whs))
	for _, webhook := range whs {
		webhooks = append(webhooks, &dto.ListWebhooksResponseWebhook{
			ID:       webhook.GetID(),
			ModuleID: webhook.GetModuleID(),
			URL:      webhook.GetURL(),
		})
	}

	return &dto.ListWebhooksResponse{
		Webhooks: webhooks,
	}, nil
}

func (svc *webhookService) RegisterWebhook(ctx context.Context, request *dto.RegisterWebhookRequest) (*dto.RegisterWebhookResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Register webhook request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	if !svc.moduleManager.ModuleExists(request.ModuleID) {
		return nil, errs.ErrNotFound
	}

	webhookID, err := svc.webhookManager.AddWebhook(request.ModuleID, request.URL)
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

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	if err := svc.webhookManager.RemoveWebhook(request.ID); err != nil {
		return nil, fmt.Errorf("failed to delete webhook: %v", err)
	}

	return &dto.DeleteWebhookResponse{}, nil
}
