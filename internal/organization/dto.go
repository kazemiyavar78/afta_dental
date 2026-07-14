// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package organization

// CreateRequest بدنه ایجاد سازمان.
type CreateRequest struct {
	Name      string `json:"name" binding:"required"`
	IsTakmili bool   `json:"is_takmili"`
	IsActive  bool   `json:"is_active"`
}

// UpdateRequest بدنه بروزرسانی سازمان.
type UpdateRequest struct {
	Name      string `json:"name" binding:"required"`
	IsTakmili bool   `json:"is_takmili"`
	IsActive  bool   `json:"is_active"`
}

// Response پاسخ API سازمان.
type Response struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	IsTakmili bool   `json:"is_takmili"`
	IsActive  bool   `json:"is_active"`
}
