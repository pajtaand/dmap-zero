package models

type WebhookData struct {
	ModuleID string `json:"moduleID"`
	Blob     string `json:"blob"`
	Receiver string `json:"Receiver"`
}
