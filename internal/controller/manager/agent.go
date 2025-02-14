package manager

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/google/uuid"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/pajtaand/dmap-zero/internal/common/constants"
	errs "github.com/pajtaand/dmap-zero/internal/common/errors"
	"github.com/pajtaand/dmap-zero/internal/common/wrapper"
	pb "github.com/pajtaand/dmap-zero/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Diagnostics struct {
	PresentImages  map[string]string
	PresentModules map[string]string
}

type diagnostics struct {
	time           time.Time
	presentImages  map[string]string
	presentModules map[string]string
}

type Agent struct {
	id            string
	name          string
	identityID    string
	configuration map[string]string
	diag          *diagnostics
	conn          *grpc.ClientConn

	mu sync.RWMutex
}

func NewAgent(id, name string, configuration map[string]string) *Agent {
	if configuration == nil {
		configuration = map[string]string{}
	}

	return &Agent{
		id:            id,
		name:          name,
		configuration: configuration,
	}
}

func (a *Agent) Connect(conn *grpc.ClientConn) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.conn = conn
	log.Info().Msgf("Agent connected: agentID=%s", a.id)
}

func (a *Agent) IsConnected() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.conn != nil
}

func (a *Agent) GetConfigurationServiceClient() pb.ConfigurationServiceClient {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.conn == nil {
		return nil
	}
	return pb.NewConfigurationServiceClient(a.conn)
}

func (a *Agent) GetImageServiceClient() pb.ImageServiceClient {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.conn == nil {
		return nil
	}
	return pb.NewImageServiceClient(a.conn)
}

func (a *Agent) GetModuleServiceClient() pb.ModuleServiceClient {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.conn == nil {
		return nil
	}
	return pb.NewModuleServiceClient(a.conn)
}

func (a *Agent) GetShareServiceClient() pb.ShareServiceClient {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.conn == nil {
		return nil
	}
	return pb.NewShareServiceClient(a.conn)
}

func (a *Agent) Cleanup() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.conn != nil {
		a.conn.Close()
		a.conn = nil
	}
}

func (a *Agent) GetID() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.id
}

func (a *Agent) GetIdentityID() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.identityID
}

func (a *Agent) GetName() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.name
}

func (a *Agent) GetConfiguration() map[string]string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.configuration
}

func (a *Agent) SetIdentityID(identityID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.identityID = identityID
	// changing identity invalidates the connection
	if a.conn != nil {
		a.conn.Close()
		a.conn = nil
	}
}

func (a *Agent) SetName(name string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.name = name
}

func (a *Agent) SetConfiguration(configuration map[string]string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if configuration == nil {
		configuration = map[string]string{}
	}
	a.configuration = configuration
}

func (a *Agent) GetDiagnostics() *Diagnostics {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.diag != nil && time.Since(a.diag.time) < constants.ControllerAgentMaxDiagnosticsDelay {
		return &Diagnostics{
			PresentImages:  a.diag.presentImages,
			PresentModules: a.diag.presentModules,
		}
	}
	return nil
}

func (a *Agent) setDiagnostics(diag *Diagnostics) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.diag = &diagnostics{
		time:           time.Now(),
		presentImages:  diag.PresentImages,
		presentModules: diag.PresentModules,
	}
}

type AgentManagerConfig struct {
	AgentServiceName string
}

type AgentManager struct {
	config         *AgentManagerConfig
	mu             sync.RWMutex
	agents         map[string]*Agent
	openZitiClient *wrapper.OpenZitiClientWrapper
}

func NewAgentManager(config *AgentManagerConfig, openZitiClient *wrapper.OpenZitiClientWrapper) (*AgentManager, error) {
	log.Debug().Msg("Creating new AgentManager")

	if config == nil {
		return nil, errors.New("config must not be nil")
	}

	return &AgentManager{
		config:         config,
		agents:         map[string]*Agent{},
		openZitiClient: openZitiClient,
	}, nil
}

func (mgr *AgentManager) AddAgent(name string, configuration map[string]string) string {
	log.Info().Msgf("Adding new agent: %s", name)

	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	agentID := uuid.New().String()
	mgr.agents[agentID] = NewAgent(agentID, name, configuration)
	return agentID
}

func (mgr *AgentManager) GetAgent(agentID string) (*Agent, error) {
	log.Info().Msgf("Getting agent: %s", agentID)

	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	agent, ok := mgr.agents[agentID]
	if !ok {
		return nil, errs.ErrNotFound
	}

	return agent, nil
}

func (mgr *AgentManager) GetAgentByIdentityID(agentIdentityID string) (*Agent, error) {
	log.Info().Msgf("Getting agent with identity ID: %s", agentIdentityID)

	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	for _, agent := range mgr.agents {
		if agent.GetIdentityID() == agentIdentityID {
			return agent, nil
		}
	}
	return nil, errs.ErrNotFound
}

func (mgr *AgentManager) ListAgents() []*Agent {
	log.Info().Msg("Listing all agents")

	mgr.mu.RLock()
	defer mgr.mu.RUnlock()
	agents := []*Agent{}
	for _, agent := range mgr.agents {
		agents = append(agents, agent)
	}
	return agents
}

func (mgr *AgentManager) RemoveAgent(agentID string) error {
	log.Info().Msgf("Removing agent: %s", agentID)

	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	agent, ok := mgr.agents[agentID]
	if !ok {
		return errs.ErrNotFound
	}

	agent.Cleanup()

	delete(mgr.agents, agentID)
	return nil
}

func (mgr *AgentManager) ReceiveAgentDiagnostics(agentID string, diag *Diagnostics) error {
	log.Info().Msgf("Receiving agent diagnostics data: %s", agentID)

	agent, err := mgr.GetAgent(agentID)
	if err != nil {
		return err
	}

	if !agent.IsConnected() {
		conn, err := grpc.NewClient(
			fmt.Sprintf("passthrough:///%s", mgr.config.AgentServiceName),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithContextDialer(mgr.openZitiClient.GetContextDialerWithOptions(&ziti.DialOptions{
				Identity: agentID,
			})),
		)
		if err != nil {
			return fmt.Errorf("failed to initialize agent connection: %v", err)
		}
		agent.Connect(conn)
	}

	agent.setDiagnostics(diag)
	return nil
}
