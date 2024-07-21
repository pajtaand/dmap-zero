package dto

type CreateModuleRequest struct {
	Name          string
	Image         string
	Configuration map[string]string
}

type CreateModuleResponse struct {
	ID string
}

type GetModuleRequest struct {
	ID string
}

type GetModuleResponse struct {
	Name          string
	Image         string
	Configuration map[string]string
	IsRunning     bool
}

type ListModulesRequest struct {
}

type ListModulesResponseModule struct {
	ID            string
	Name          string
	Image         string
	Configuration map[string]string
	IsRunning     bool
}

type ListModulesResponse struct {
	Modules []*ListModulesResponseModule
}

type UpdateModuleRequest struct {
	ID            string
	Name          string
	Image         string
	Configuration map[string]string
}

type UpdateModuleResponse struct {
}

type DeleteModuleRequest struct {
	ID string
}

type DeleteModuleResponse struct {
}

type StartModuleRequest struct {
	ID string
}

type StartModuleResponse struct {
}

type StopModuleRequest struct {
	ID string
}

type StopModuleResponse struct {
}

type SendDataRequest struct {
	ModuleID string
	Data     []byte
}

type SendDataResponse struct {
}
