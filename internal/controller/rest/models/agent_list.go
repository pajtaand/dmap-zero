package models

type ListAgentsResponseAgent struct {
	ID             string
	Name           string
	Configuration  map[string]string
	IsEnrolled     bool
	IsOnline       bool
	PresentImages  []string
	PresentModules []string
}

type ListAgentsResponse struct {
	Agents []ListAgentsResponseAgent
}
