package handler

import (
	"context"

	"github.com/andreepyro/dmap-zero/internal/agent/dto"
)

type ControllerService interface {
	PushBlob(ctx context.Context, req *dto.ControllerPushBlobRequest) (*dto.ControllerPushBlobResponse, error)
}

type EndpointService interface {
	ListEndpoints(ctx context.Context, req *dto.ListEndpointsRequest) (*dto.ListEndpointsResponse, error)
	PushBlob(ctx context.Context, req *dto.EndpointPushBlobRequest) (*dto.EndpointPushBlobResponse, error)
}

type WebhookService interface {
	ListWebhooks(ctx context.Context, req *dto.ListWebhooksRequest) (*dto.ListWebhooksResponse, error)
	RegisterWebhook(ctx context.Context, req *dto.RegisterWebhookRequest) (*dto.RegisterWebhookResponse, error)
	DeleteWebhook(ctx context.Context, req *dto.DeleteWebhookRequest) (*dto.DeleteWebhookResponse, error)
}
