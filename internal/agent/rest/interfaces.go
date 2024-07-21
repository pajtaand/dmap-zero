package rest

import (
	"net/http"
)

type EndpointHandler interface {
	ListEndpoints(w http.ResponseWriter, r *http.Request)
	PushBlobToEndpoint(w http.ResponseWriter, r *http.Request)
}

type ControllerHandler interface {
	PushBlobToController(w http.ResponseWriter, r *http.Request)
}

type WebhookHandler interface {
	ListWebhooks(w http.ResponseWriter, r *http.Request)
	RegisterWebhook(w http.ResponseWriter, r *http.Request)
	DeleteWebhook(w http.ResponseWriter, r *http.Request)
}
