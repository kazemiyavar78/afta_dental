package specialcode

// CreateRequest بدنه ایجاد کد خاص.
type CreateRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Percentage  uint8  `json:"percentage"`
	IsActive    bool   `json:"is_active"`
}

// UpdateRequest بدنه بروزرسانی کد خاص.
type UpdateRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Percentage  uint8  `json:"percentage"`
	IsActive    bool   `json:"is_active"`
}

// Response پاسخ API کد خاص.
type Response struct {
	ID          uint   `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Percentage  uint8  `json:"percentage"`
	IsActive    bool   `json:"is_active"`
}
