package manager

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/google/uuid"
	"github.com/pajtaand/dmap-zero/internal/agent/dto"
	"github.com/pajtaand/dmap-zero/internal/agent/rest/models"
	errs "github.com/pajtaand/dmap-zero/internal/common/errors"
	"github.com/pajtaand/dmap-zero/internal/common/utils"
)

type Webhook struct {
	id      string
	urlPath string
	port    string
	event   dto.WebhookEvent

	mu sync.RWMutex
}

func NewWebhook(id, urlPath, port string, event dto.WebhookEvent) *Webhook {
	return &Webhook{
		id:      id,
		urlPath: urlPath,
		port:    port,
		event:   event,
	}
}

func (a *Webhook) GetID() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.id
}

func (a *Webhook) GetURLPath() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.urlPath
}

func (a *Webhook) GetPort() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.port
}

func (a *Webhook) GetEvent() dto.WebhookEvent {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.event
}

type WebhookManager struct {
	mu       sync.RWMutex
	webhooks map[string]map[string]*Webhook
}

func NewWebhookManager() (*WebhookManager, error) {
	log.Debug().Msg("Creating new WebhookManager")

	return &WebhookManager{
		webhooks: map[string]map[string]*Webhook{},
	}, nil
}

func (mgr *WebhookManager) AddModule(sourceModuleID string) error {
	log.Info().Msgf("Adding new source module: %s", sourceModuleID)

	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	if _, ok := mgr.webhooks[sourceModuleID]; ok {
		log.Warn().Msg("Module with the same ID already exists. Replacing...")
	}

	mgr.webhooks[sourceModuleID] = map[string]*Webhook{}
	return nil
}

func (mgr *WebhookManager) RemoveModule(sourceModuleID string) error {
	log.Info().Msgf("Removing source module: %s", sourceModuleID)

	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	if _, ok := mgr.webhooks[sourceModuleID]; !ok {
		return errs.ErrNotFound
	}

	delete(mgr.webhooks, sourceModuleID)
	return nil
}

func (mgr *WebhookManager) AddWebhook(sourceModuleID, urlPath, port string, event dto.WebhookEvent) (string, error) {
	log.Info().Msgf("Adding new webhook: sourceModuleID=%s, urlPath=%s, port=%s, event=%s", sourceModuleID, urlPath, port, event)

	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	module, ok := mgr.webhooks[sourceModuleID]
	if !ok {
		return "", errors.New("source module doesn't exist")
	}

	webhookID := uuid.New().String()
	module[webhookID] = NewWebhook(webhookID, urlPath, port, event)
	return webhookID, nil
}

func (mgr *WebhookManager) GetWebhook(sourceModuleID, webhookID string) (*Webhook, error) {
	log.Info().Msgf("Getting webhook: sourceModuleID=%s, webhookID=%s", sourceModuleID, webhookID)

	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	module, ok := mgr.webhooks[sourceModuleID]
	if !ok {
		return nil, errors.New("source module doesn't exist")
	}

	webhook, ok := module[webhookID]
	if !ok {
		return nil, errs.ErrNotFound
	}

	return webhook, nil
}

func (mgr *WebhookManager) ListWebhooks(sourceModuleID string) ([]*Webhook, error) {
	log.Info().Msgf("Listing all webhooks for source module: %s", sourceModuleID)

	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	module, ok := mgr.webhooks[sourceModuleID]
	if !ok {
		return nil, errors.New("source module doesn't exist")
	}

	webhooks := []*Webhook{}
	for _, webhook := range module {
		webhooks = append(webhooks, webhook)
	}
	return webhooks, nil
}

func (mgr *WebhookManager) ListWebhooksForEvent(sourceModuleID string, event dto.WebhookEvent) ([]*Webhook, error) {
	log.Info().Msgf("Listing all webhooks for event: %s", event)

	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	module, ok := mgr.webhooks[sourceModuleID]
	if !ok {
		return nil, errors.New("source module doesn't exist")
	}

	webhooks := []*Webhook{}
	for _, webhook := range module {
		if webhook.GetEvent() == event {
			webhooks = append(webhooks, webhook)
		}
	}
	return webhooks, nil
}

func (mgr *WebhookManager) RemoveWebhook(sourceModuleID, webhookID string) error {
	log.Info().Msgf("Removing webhook: sourceModuleID=%s, webhookID=%s", sourceModuleID, webhookID)

	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	module, ok := mgr.webhooks[sourceModuleID]
	if !ok {
		return errors.New("source module doesn't exist")
	}

	if _, ok := module[webhookID]; !ok {
		return errs.ErrNotFound
	}

	delete(module, webhookID)
	return nil
}

func (mgr *WebhookManager) WebhookExists(sourceModuleID, webhookID string) (bool, error) {
	log.Info().Msgf("Checking if webhook exists: sourceModuleID=%s, webhookID=%s", sourceModuleID, webhookID)

	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	module, ok := mgr.webhooks[sourceModuleID]
	if !ok {
		return false, errors.New("source module doesn't exist")
	}

	_, ok = module[webhookID]
	return ok, nil
}

func (mgr *WebhookManager) SendData(sourceEndpointID, receiverModuleID string, event dto.WebhookEvent, data []byte) error {
	log.Info().Msgf("Sending data to webhook: sourceModuleID=%s, receiverModuleID=%s, event=%s", sourceEndpointID, receiverModuleID, event)

	// prepare payload
	base64String := base64.StdEncoding.EncodeToString(data)
	webhookData := models.WebhookData{
		SourceEndpointID: sourceEndpointID,
		Blob:             base64String,
	}

	payload, err := json.Marshal(webhookData)
	if err != nil {
		return fmt.Errorf("failed to marshal data into JSON: %v", err)
	}

	// send payload to all registered urls concurrently
	webhooks, err := mgr.ListWebhooksForEvent(receiverModuleID, event)
	if err != nil {
		return fmt.Errorf("failed to list webhooks: %v", err)
	}

	if len(webhooks) == 0 {
		log.Debug().Msg("No webhooks registered. Skipping...")
		return nil
	}

	var wg sync.WaitGroup
	results := make(chan bool, len(webhooks))

	for _, webhook := range webhooks {
		wg.Add(1)
		go func(URLPath, port string, res chan bool) {
			address := fmt.Sprintf("http://localhost:%s%s", port, URLPath)
			log.Debug().Msgf("Sending data to webhook: %s", address)
			err := utils.SendPOSTRequest(address, payload)
			if err != nil {
				log.Warn().Msgf("failed to send data: %v", err)
			}
			res <- err == nil
			wg.Done()
		}(webhook.GetURLPath(), webhook.GetPort(), results)
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
