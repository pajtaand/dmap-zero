package models

type ListModulesResponseModule struct {
	ID            string
	Name          string
	Image         string
	Configuration map[string]string
	IsRunning     bool
}

type ListModulesResponse struct {
	Modules []ListModulesResponseModule
}
