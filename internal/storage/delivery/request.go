package delivery

type UploadRequest struct {
	Module      string  `form:"module"`
	ReferenceID *string `form:"reference_id"`
}

type ListByModuleRequest struct {
	Module      string `params:"module"`
	ReferenceID string `params:"reference_id"`
}
