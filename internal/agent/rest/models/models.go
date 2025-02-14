package models

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pajtaand/dmap-zero/internal/common/utils"
)

type Endpoint struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type BlobPushRequest struct {
	ID string `json:"id"`
}

type Webhook struct {
	ID      string `json:"id"`
	URLPath string `json:"urlPath"`
	Event   string `json:"event"`
}

type WebhookRegistrationRequest struct {
	URLPath string `json:"urlPath"`
	Event   string `json:"event"`
}

func (req *WebhookRegistrationRequest) FromHttpRequest(r *http.Request) error {
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	if err := utils.CheckStringNotEmpty(req, "URLPath"); err != nil {
		return err
	}
	if err := utils.CheckStringNotEmpty(req, "Event"); err != nil {
		return err
	}
	return nil
}

type WebhookRegistrationResponse struct {
	ID string `json:"ID"`
}

type ControllerPushRequest struct {
	ReceiverID string
	Blob       []byte
}

func (req *ControllerPushRequest) FromHttpRequest(r *http.Request) error {
	tmp := struct {
		ReceiverID string `json:"receiverId"`
		Blob       string `json:"blob"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&tmp); err != nil {
		return err
	}
	if err := utils.CheckStringNotEmpty(tmp, "ReceiverID"); err != nil {
		return err
	}
	if err := utils.CheckStringNotEmpty(tmp, "Blob"); err != nil {
		return err
	}

	blob, err := base64.StdEncoding.DecodeString(tmp.Blob)
	if err != nil {
		return fmt.Errorf("failed to parse blob: %v", err)
	}

	req.ReceiverID = tmp.ReceiverID
	req.Blob = blob

	return nil
}

type WebhookData struct {
	SourceEndpointID string `json:"sourceEndpointID"`
	Blob             string `json:"blob"`
}
