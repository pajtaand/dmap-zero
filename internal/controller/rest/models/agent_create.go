package models

import (
	"encoding/json"
	"net/http"

	"github.com/andreepyro/dmap-zero/internal/common/utils"
)

type CreateAgentRequest struct {
	Name          string
	Configuration map[string]string
}

func (req *CreateAgentRequest) FromHttpRequest(r *http.Request) error {
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	if err := utils.CheckStringNotEmpty(req, "Name"); err != nil {
		return err
	}
	if err := utils.CheckNotNil(req, "Configuration"); err != nil {
		return err
	}
	return nil
}

type CreateAgentResponse struct {
	ID string
}
