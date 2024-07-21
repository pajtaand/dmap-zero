package manager

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"

	errs "github.com/andreepyro/dmap-zero/internal/common/errors"
	"github.com/andreepyro/dmap-zero/internal/common/utils"
	"github.com/andreepyro/dmap-zero/internal/controller/rest/models"
	"github.com/google/uuid"
)

type Webhook struct {
	id       string
	moduleID string
	URL      string

	mu sync.RWMutex
}

func NewWebhook(id, moduleID, URL string) *Webhook {
	return &Webhook{
		id:       id,
		moduleID: moduleID,
		URL:      URL,
	}
}

func (a *Webhook) GetID() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.id
}

func (a *Webhook) GetModuleID() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.moduleID
}

func (a *Webhook) GetURL() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.URL
}

type WebhookManager struct {
	mu       sync.RWMutex
	webhooks map[string]*Webhook
}

func NewWebhookManager() (*WebhookManager, error) {
	log.Debug().Msg("Creating new WebhookManager")

	return &WebhookManager{
		webhooks: map[string]*Webhook{},
	}, nil
}

func (mgr *WebhookManager) AddWebhook(moduleID, URL string) (string, error) {
	log.Info().Msgf("Adding new webhook: moduleID=%s, URL=%s", moduleID, URL)

	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	webhookID := uuid.New().String()
	mgr.webhooks[webhookID] = NewWebhook(webhookID, moduleID, URL)
	return webhookID, nil
}

func (mgr *WebhookManager) GetWebhook(webhookID string) (*Webhook, error) {
	log.Info().Msgf("Getting webhook: webhookID=%s", webhookID)

	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	webhook, ok := mgr.webhooks[webhookID]
	if !ok {
		return nil, errs.ErrNotFound
	}

	return webhook, nil
}

func (mgr *WebhookManager) ListWebhooks() []*Webhook {
	log.Info().Msg("Listing all webhooks")

	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	webhooks := []*Webhook{}
	for _, webhook := range mgr.webhooks {
		webhooks = append(webhooks, webhook)
	}
	return webhooks
}

func (mgr *WebhookManager) ListWebhooksForModule(ModuleID string) []*Webhook {
	log.Info().Msgf("Listing all webhooks for ModuleID: %s", ModuleID)

	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	webhooks := []*Webhook{}
	for _, webhook := range mgr.webhooks {
		if webhook.GetModuleID() == ModuleID {
			webhooks = append(webhooks, webhook)
		}
	}
	return webhooks
}

func (mgr *WebhookManager) RemoveWebhook(webhookID string) error {
	log.Info().Msgf("Removing webhook: webhookID=%s", webhookID)

	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	if _, ok := mgr.webhooks[webhookID]; !ok {
		return errs.ErrNotFound
	}

	delete(mgr.webhooks, webhookID)
	return nil
}

func (mgr *WebhookManager) WebhookExists(webhookID string) bool {
	log.Info().Msgf("Checking if webhook exists: webhookID=%s", webhookID)

	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	_, ok := mgr.webhooks[webhookID]
	return ok
}

func (mgr *WebhookManager) SendData(moduleID, receiver string, data []byte) error {
	log.Info().Msgf("Sending data to webhook: moduleID=%s", moduleID)

	// prepare payload
	base64String := base64.StdEncoding.EncodeToString(data)
	webhookData := models.WebhookData{
		ModuleID: moduleID,
		Blob:     base64String,
		Receiver: receiver,
	}

	payload, err := json.Marshal(webhookData)
	if err != nil {
		return fmt.Errorf("failed to marshal data into JSON: %v", err)
	}

	// send payload to all registered urls concurrently
	webhooks := mgr.ListWebhooksForModule(moduleID)

	if len(webhooks) == 0 {
		log.Debug().Msg("No webhooks registered. Skipping...")
		return nil
	}

	var wg sync.WaitGroup
	results := make(chan bool, len(webhooks))

	for _, webhook := range webhooks {
		wg.Add(1)
		go func(URL string, res chan bool) {
			log.Debug().Msgf("Sending data to webhook: %s", URL)
			err := utils.SendPOSTRequest(URL, payload)
			if err != nil {
				log.Warn().Msgf("failed to send data: %v", err)
			}
			res <- err == nil
			wg.Done()
		}(webhook.GetURL(), results)
	}

	wg.Wait()
	close(results)

	// succeed if at least one call was successfully
	for result := range results {
		if result {
			return nil
		}
	}

	return fmt.Errorf("none of %d registered webhook urls were reached", len(webhooks))
}
