package fund

type CreateRequest struct {
	Name string `json:"name" binding:"required"`
}

type Response struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}
