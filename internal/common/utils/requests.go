package utils

import (
	"bytes"
	"fmt"
	"net/http"
)

func SendPOSTRequest(URL string, payload []byte) error {
	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create POST request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", code)
	}

	return nil
}
