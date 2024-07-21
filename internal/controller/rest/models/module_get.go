package models

type GetModuleResponse struct {
	Name          string
	Image         string
	Configuration map[string]string
	IsRunning     bool
}
