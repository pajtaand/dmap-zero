package dto

type ListEndpointsRequest struct {
	SourceModuleID string
}

type ListEndpointsResponse struct {
	Endpoints []*ListEndpointsResponseEndpoint
}

type ListEndpointsResponseEndpoint struct {
	ID string
}

type EndpointPushBlobRequest struct {
	SourceModuleID     string
	ReceiverIdentityID string
	ReceiverModuleID   string
	Blob               []byte
}

type EndpointPushBlobResponse struct {
}
