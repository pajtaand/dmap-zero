package models

type Webhook struct {
	ID       string `json:"id"`
	ModuleID string `json:"moduleID"`
	URL      string `json:"url"`
}
