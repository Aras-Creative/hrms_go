package delivery

type CreateRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}

type UpdateRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
}
