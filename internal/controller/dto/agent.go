package dto

type CreateAgentRequest struct {
	Name          string
	Configuration map[string]string
}

type CreateAgentResponse struct {
	ID string
}

type GetAgentRequest struct {
	ID string
}

type GetAgentResponse struct {
	Name           string
	Configuration  map[string]string
	IsEnrolled     bool
	IsOnline       bool
	PresentImages  []string
	PresentModules []string
}

type ListAgentsRequest struct {
}

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
	Agents []*ListAgentsResponseAgent
}

type UpdateAgentRequest struct {
	ID            string
	Name          string
	Configuration map[string]string
}

type UpdateAgentResponse struct {
}

type DeleteAgentRequest struct {
	ID string
}

type DeleteAgentResponse struct {
}
