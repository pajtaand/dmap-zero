package models

import (
	"encoding/json"
	"net/http"

	"github.com/andreepyro/dmap-zero/internal/common/utils"
)

type SendDataRequest struct {
	Data []byte
}

func (req *SendDataRequest) FromHttpRequest(r *http.Request) error {
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	if err := utils.CheckNotNil(req, "Data"); err != nil {
		return err
	}
	return nil
}
