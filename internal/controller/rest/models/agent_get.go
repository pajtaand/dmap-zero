package models

type GetAgentResponse struct {
	Name           string
	Configuration  map[string]string
	IsEnrolled     bool
	IsOnline       bool
	PresentImages  []string
	PresentModules []string
}
