package models

import (
	"encoding/json"
	"net/http"

	"github.com/pajtaand/dmap-zero/internal/common/utils"
)

type WebhookRegistrationRequest struct {
	ModuleID string `json:"moduleID"`
	URL      string `json:"url"`
}

func (req *WebhookRegistrationRequest) FromHttpRequest(r *http.Request) error {
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	if err := utils.CheckStringNotEmpty(req, "ModuleID"); err != nil {
		return err
	}
	if err := utils.CheckStringNotEmpty(req, "URL"); err != nil {
		return err
	}
	return nil
}

type WebhookRegistrationResponse struct {
	ID string `json:"ID"`
}
