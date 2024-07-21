package constants

import (
	"time"
)

const (
	// Shared
	LoggerKeyRequestID   = "request_id"
	LoggerKeyRequestUser = "user"
	LoggerKeyUrl         = "url"
	LoggerKeyMethod      = "method"
	LoggerKeyUserAgent   = "user_agent"
	LoggerKeyElapsedTime = "elapsed_ms"
	LoggerKeyStatusCode  = "status_code"

	OpenZitiIdentityController         = "controller"
	OpenZitiRoleAgent                  = "agent-role"
	OpenZitiAdminAgent                 = false
	OpenZitiServiceController          = "service-controller"
	OpenZitiServiceAgent               = "service-agent"
	OpenZitiServiceP2P                 = "service-p2p"
	OpenZitiEnrollmentTokenValidity    = 1 * time.Hour
	OpenZitiManagementAPIReloadReserve = 15 * time.Second

	// Controller
	ControllerEnvAPICredentials        = "API_CREDENTIALS"
	ControllerEnvAPICertFile           = "API_CERT_FILE"
	ControllerEnvAPIKeyFile            = "API_KEY_FILE"
	ControllerEnvEnrollmentToken       = "ENROLLMENT_TOKEN"
	ControllerAPIAddress               = "0.0.0.0:6969"
	ControllerMetricsAPIAddress        = "0.0.0.0:9090"
	ControllerAgentMaxDiagnosticsDelay = 15 * time.Second

	// Agent
	AgentDockerHostAddress               = "127.0.0.1"
	AgentModuleServerIP                  = "0.0.0.0"
	AgentModuleServerDefaultPort         = 4499
	AgentModuleServerCertificateValidity = time.Hour * 24 * 365
	AgentPhonehomeInterval               = 10 * time.Second
	AgentPingInterval                    = 60 * time.Second
	AgentImageStreamChunkSize            = 1024

	// Module
	ModuleEnvAPIBaseUrl  = "MODULE_API_BASE_URL"
	ModuleEnvUsername    = "MODULE_API_BASEAUTH_USER"
	ModuleEnvPassword    = "MODULE_API_BASEAUTH_PASS"
	ModuleEnvCertificate = "MODULE_API_CERTIFICATE"
	ModuleEnvGivenPort   = "MODULE_GIVEN_PORT"
	ModulePortRangeMin   = 33000
	ModulePortRangeMax   = 33999
)
