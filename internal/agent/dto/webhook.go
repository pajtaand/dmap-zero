package dto

import "errors"

type WebhookEvent string

const (
	EventControllerData WebhookEvent = "CONTROLLER_DATA"
	EventEndpointData   WebhookEvent = "ENDPOINT_DATA"
)

func ParseWebhookEvent(eventStr string) (WebhookEvent, error) {
	switch eventStr {
	case string(EventControllerData):
		return EventControllerData, nil
	case string(EventEndpointData):
		return EventEndpointData, nil
	default:
		return "", errors.New("invalid event")
	}
}

type RegisterWebhookRequest struct {
	SourceModuleID string
	URLPath        string
	Event          WebhookEvent
}

type RegisterWebhookResponse struct {
	ID string
}

type ListWebhooksRequest struct {
	SourceModuleID string
}

type ListWebhooksResponse struct {
	Webhooks []*ListWebhooksResponseWebhook
}

type ListWebhooksResponseWebhook struct {
	ID      string
	URLPath string
	Event   WebhookEvent
}

type DeleteWebhookRequest struct {
	SourceModuleID string
	ID             string
}

type DeleteWebhookResponse struct {
}
