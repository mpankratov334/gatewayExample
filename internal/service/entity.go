package service

// TaskRequest - структура, представляющая тело запроса
type TaskRequest struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
}

type TaskGetRequest struct {
	ID string `validate:"required,intString,min=1"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}
