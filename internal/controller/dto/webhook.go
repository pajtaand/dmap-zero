package dto

type RegisterWebhookRequest struct {
	ModuleID string
	URL      string
}

type RegisterWebhookResponse struct {
	ID string
}

type ListWebhooksRequest struct {
}

type ListWebhooksResponse struct {
	Webhooks []*ListWebhooksResponseWebhook
}

type ListWebhooksResponseWebhook struct {
	ID       string
	ModuleID string
	URL      string
}

type DeleteWebhookRequest struct {
	ID string
}

type DeleteWebhookResponse struct {
}
