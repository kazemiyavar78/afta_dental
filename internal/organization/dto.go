// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package organization

type CreateRequest struct {
	Name string `json:"name" binding:"required"`
}

type Response struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
