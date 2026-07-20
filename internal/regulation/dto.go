package regulation

// CreateRequest بدنه ایجاد ضابطه.
type CreateRequest struct {
	ServiceIDs   []uint `json:"service_ids" binding:"required"`
	DurationDays int    `json:"duration_days" binding:"required"`
	IsActive     bool   `json:"is_active"`
	PhotoCount   int    `json:"photo_count"`
	Description  string `json:"description"`
}

// UpdateRequest بدنه بروزرسانی ضابطه.
type UpdateRequest struct {
	ServiceIDs   []uint `json:"service_ids" binding:"required"`
	DurationDays int    `json:"duration_days" binding:"required"`
	IsActive     bool   `json:"is_active"`
	PhotoCount   int    `json:"photo_count"`
	Description  string `json:"description"`
}

// Response پاسخ API ضابطه.
type Response struct {
	ID           uint   `json:"id"`
	ServiceIDs   []uint `json:"service_ids"`
	DurationDays int    `json:"duration_days"`
	IsActive     bool   `json:"is_active"`
	PhotoCount   int    `json:"photo_count"`
	Description  string `json:"description"`
}
