package rest

import (
	"net/http"
)

type WebAppHandler interface {
	GetWebApplication(w http.ResponseWriter, r *http.Request)
	GetFavicon(w http.ResponseWriter, r *http.Request)
	GetIcon(w http.ResponseWriter, r *http.Request)
	GetCSS(w http.ResponseWriter, r *http.Request)
	GetJS(w http.ResponseWriter, r *http.Request)
}

type AgentHandler interface {
	CreateAgent(w http.ResponseWriter, r *http.Request)
	ListAgents(w http.ResponseWriter, r *http.Request)
	GetAgent(w http.ResponseWriter, r *http.Request)
	UpdateAgent(w http.ResponseWriter, r *http.Request)
	DeleteAgent(w http.ResponseWriter, r *http.Request)
}

type ModuleHandler interface {
	CreateModule(w http.ResponseWriter, r *http.Request)
	ListModules(w http.ResponseWriter, r *http.Request)
	GetModule(w http.ResponseWriter, r *http.Request)
	UpdateModule(w http.ResponseWriter, r *http.Request)
	DeleteModule(w http.ResponseWriter, r *http.Request)
	StartModule(w http.ResponseWriter, r *http.Request)
	StopModule(w http.ResponseWriter, r *http.Request)
	SendData(w http.ResponseWriter, r *http.Request)
}

type ImageHandler interface {
	UploadImage(w http.ResponseWriter, r *http.Request)
	ListImages(w http.ResponseWriter, r *http.Request)
	GetImage(w http.ResponseWriter, r *http.Request)
	DeleteImage(w http.ResponseWriter, r *http.Request)
}

type WebhookHandler interface {
	ListWebhooks(w http.ResponseWriter, r *http.Request)
	RegisterWebhook(w http.ResponseWriter, r *http.Request)
	DeleteWebhook(w http.ResponseWriter, r *http.Request)
}

type EnrollmentHandler interface {
	GetEnrollment(w http.ResponseWriter, r *http.Request)
	CreateEnrollment(w http.ResponseWriter, r *http.Request)
	DeleteEnrollment(w http.ResponseWriter, r *http.Request)
}
