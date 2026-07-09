// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package reception

// CreateReceptionRequest درخواست ایجاد پذیرش.
type CreateReceptionRequest struct {
	PatientName   string `json:"patient_name" binding:"required"`
	DoctorID      int    `json:"doctor_id" binding:"required"`
	ReceptionDate string `json:"reception_date" binding:"required"`
}

// ReceptionResponse پاسخ اطلاعات پذیرش.
type ReceptionResponse struct {
	ID            int    `json:"id"`
	PatientName   string `json:"patient_name"`
	DoctorID      int    `json:"doctor_id"`
	ReceptionDate string `json:"reception_date"`
	Status        string `json:"status"`
}
