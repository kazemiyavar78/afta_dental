package reception

import "github.com/tpdenta/afta-reception/internal/patient"

// PatientInput اطلاعات بیمار برای ایجاد/به‌روزرسانی همراه پذیرش.
type PatientInput struct {
	ID                  *uint   `json:"id"`
	FirstName           string  `json:"first_name"`
	LastName            string  `json:"last_name"`
	NationalCode        string  `json:"national_code"`
	BirthDate           string  `json:"birth_date"`
	Address             *string `json:"address"`
	HomePhoneNumber     *string `json:"home_phone_number"`
	MobilePhoneNumber   *string `json:"mobile_phone_number"`
	FileNumber          string  `json:"file_number"`
	Sex                 bool    `json:"sex"`
	IsForeignNational   bool    `json:"is_foreign_national"`
}

// ServiceLineInput یک سطر خدمت در درخواست پذیرش یا محاسبه.
type ServiceLineInput struct {
	ServiceID          uint   `json:"service_id"`
	ServiceCode        string `json:"service_code"`
	Quantity           int    `json:"quantity"`
	TeethNumber        *uint8 `json:"teeth_number"`
	TeethDirection     *uint8 `json:"teeth_direction"`
	ServiceDescription string `json:"service_description"`
}

// UpsertReceptionRequest بدنه ایجاد یا ویرایش پذیرش.
type UpsertReceptionRequest struct {
	Patient                         PatientInput       `json:"patient" binding:"required"`
	InsuranceID                     *uint              `json:"insurance_id"`
	AdditionalInsuranceID           *uint              `json:"additional_insurance_id"`
	SpecialCodeID                   *uint              `json:"special_code_id"`
	DoctorID                        *uint              `json:"doctor_id"`
	AssistantID                     *uint              `json:"assistant_id"`
	BookingDate                     *string            `json:"booking_date"`
	ReceptionDate                   string             `json:"reception_date" binding:"required"`
	Description                     string             `json:"description"`
	Discount                        int64              `json:"discount"`
	ReferralCode                    *int64             `json:"referral_code"`
	AdditionalInsuranceCoverage     *int64             `json:"additional_insurance_coverage"`
	AdditionalInsurancePercentage   *uint8             `json:"additional_insurance_percentage"`
	Services                        []ServiceLineInput `json:"services"`
	Save                            bool               `json:"save"` // true = وضعیت saved
}

// CalculateRequest درخواست محاسبه خدمات بدون ذخیره.
type CalculateRequest struct {
	InsuranceID                   *uint              `json:"insurance_id"`
	AdditionalInsuranceID         *uint              `json:"additional_insurance_id"`
	SpecialCodeID                 *uint              `json:"special_code_id"`
	AdditionalInsuranceCoverage   *int64             `json:"additional_insurance_coverage"`
	AdditionalInsurancePercentage *uint8             `json:"additional_insurance_percentage"`
	Services                      []ServiceLineInput `json:"services" binding:"required"`
}

// CalculatedServiceLine نتیجه محاسبه یک سطر خدمت.
type CalculatedServiceLine struct {
	ServiceID                          uint   `json:"service_id"`
	ServiceCode                        string `json:"service_code"`
	ServiceName                        string `json:"service_name"`
	Quantity                           int    `json:"quantity"`
	ServiceAmount                      int64  `json:"service_amount"`
	ServiceTariff                      int64  `json:"service_tariff"`
	ServiceOrganizationShare           int64  `json:"service_organization_share"`
	ServiceSupplementaryInsuranceShare int64  `json:"service_supplementary_insurance_share"`
	ServiceSubsidyShare                int64  `json:"service_subsidy_share"`
	ServiceDescription                 string `json:"service_description"`
	TeethNumber                        *uint8 `json:"teeth_number"`
	TeethDirection                     *uint8 `json:"teeth_direction"`
	HasDentalDirection                 bool   `json:"has_dental_direction"`
	HasTooth                           bool   `json:"has_tooth"`
}

// CalculateResponse پاسخ محاسبه خدمات.
type CalculateResponse struct {
	Services []CalculatedServiceLine `json:"services"`
}

// ReceptionServiceResponse پاسخ یک سطر خدمت پذیرش.
type ReceptionServiceResponse struct {
	ID                                 uint   `json:"id"`
	ServiceID                          uint   `json:"service_id"`
	ServiceName                        string `json:"service_name"`
	Quantity                           int    `json:"quantity"`
	ServiceAmount                      int64  `json:"service_amount"`
	ServiceTariff                      int64  `json:"service_tariff"`
	ServiceOrganizationShare           int64  `json:"service_organization_share"`
	ServiceSupplementaryInsuranceShare int64  `json:"service_supplementary_insurance_share"`
	ServiceSubsidyShare                int64  `json:"service_subsidy_share"`
	ServiceDescription                 string `json:"service_description"`
	TeethNumber                        *uint8 `json:"teeth_number"`
	TeethDirection                     *uint8 `json:"teeth_direction"`
	HasDentalDirection                 bool   `json:"has_dental_direction"`
	HasTooth                           bool   `json:"has_tooth"`
}

// ReceptionPhotoResponse پاسخ یک عکس پذیرش.
type ReceptionPhotoResponse struct {
	ID       uint   `json:"id"`
	FileName string `json:"file_name"`
	URL      string `json:"url"`
}

// ReceptionResponse پاسخ کامل پذیرش.
type ReceptionResponse struct {
	ID                              uint                       `json:"id"`
	PatientID                       uint                       `json:"patient_id"`
	Patient                         *patient.Response          `json:"patient,omitempty"`
	InsuranceID                     *uint                      `json:"insurance_id"`
	AdditionalInsuranceID           *uint                      `json:"additional_insurance_id"`
	SpecialCodeID                   *uint                      `json:"special_code_id"`
	SpecialCodeName                 string                     `json:"special_code_name,omitempty"`
	SpecialCodeValue                string                     `json:"special_code_value,omitempty"`
	DoctorID                        *uint                      `json:"doctor_id"`
	AssistantID                     *uint                      `json:"assistant_id"`
	DoctorName                      string                     `json:"doctor_name,omitempty"`
	DoctorMedicalCode               *string                    `json:"doctor_medical_code,omitempty"`
	AssistantName                   string                     `json:"assistant_name,omitempty"`
	BookingDate                     *string                    `json:"booking_date"`
	ReceptionDate                   string                     `json:"reception_date"`
	Status                          string                     `json:"status"`
	Description                     string                     `json:"description"`
	Discount                        int64                      `json:"discount"`
	ReferralCode                    *int64                     `json:"referral_code"`
	AdditionalInsuranceCoverage     *int64                     `json:"additional_insurance_coverage"`
	AdditionalInsurancePercentage   *uint8                     `json:"additional_insurance_percentage"`
	ReceptionEnded                  bool                       `json:"reception_ended"`
	PhotoCount                      int                        `json:"photo_count"`
	RegisteredByID                  *uint                      `json:"registered_by_id"`
	Services                        []ReceptionServiceResponse `json:"services"`
	Photos                          []ReceptionPhotoResponse   `json:"photos,omitempty"`
	Deleted                         bool                       `json:"deleted"`
	// Empty یعنی هیچ پذیرشی در دیتابیس نیست (فرم خالی نمایش داده شود)
	Empty bool `json:"empty,omitempty"`
}

// EndReceptionResponse نتیجه تلاش برای پایان پذیرش.
type EndReceptionResponse struct {
	Success               bool     `json:"success"`
	ReceptionEnded        bool     `json:"reception_ended"`
	PreviousReceptionID   *uint    `json:"previous_reception_id,omitempty"`
	RegulationDescriptions []string `json:"regulation_descriptions,omitempty"`
	RequiredPhotoCount    int      `json:"required_photo_count"`
	UploadedPhotoCount    int      `json:"uploaded_photo_count"`
	Message               string   `json:"message"`
}

// PatientServiceHistoryItem یک ردیف تاریخچه خدمات پرونده.
type PatientServiceHistoryItem struct {
	ReceptionID             uint     `json:"reception_id"`
	ReceptionDate           string   `json:"reception_date"`
	InsuranceName           string   `json:"insurance_name"`
	AdditionalInsuranceName string   `json:"additional_insurance_name"`
	CashAmount              int64    `json:"cash_amount"`
	ServiceNames            []string `json:"service_names"`
}

// NavQuery پارامترهای ناوبری پذیرش.
type NavQuery struct {
	Cursor *uint  `form:"cursor"`
	Dir    string `form:"dir" binding:"required"`
}
