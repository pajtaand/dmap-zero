package dto

type ControllerPushBlobRequest struct {
	SourceModuleID   string
	ReceiverModuleID string
	Blob             []byte
}

type ControllerPushBlobResponse struct {
}
