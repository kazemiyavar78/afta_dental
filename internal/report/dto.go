package report

// RevenueByOrganizationQuery پارامترهای query گزارش درآمد به تفکیک سازمان.
type RevenueByOrganizationQuery struct {
	FromDate                    string `form:"from_date" binding:"required"`
	ToDate                      string `form:"to_date" binding:"required"`
	FromTime                    string `form:"from_time"`
	ToTime                      string `form:"to_time"`
	BaseOrganizationIDs          string `form:"base_organization_ids"`
	SupplementaryOrganizationIDs string `form:"supplementary_organization_ids"`
	DetailSupplementary          bool   `form:"detail_supplementary"`
	OnlySupplementary            bool   `form:"only_supplementary"`
}

// RevenueByOrganizationRow یک سطر تجمیعی گزارش درآمد.
type RevenueByOrganizationRow struct {
	RowLabel                      string `json:"row_label"`
	BaseOrganizationID            uint   `json:"base_organization_id"`
	BaseOrganizationName          string `json:"base_organization_name"`
	SupplementaryOrganizationID   *uint  `json:"supplementary_organization_id,omitempty"`
	SupplementaryOrganizationName string `json:"supplementary_organization_name,omitempty"`
	ServiceAmount                 int64  `json:"service_amount"`
	ServiceTariff                 int64  `json:"service_tariff"`
	OrganizationShare             int64  `json:"organization_share"`
	SupplementaryShare            int64  `json:"supplementary_share"`
	SubsidyShare                  int64  `json:"subsidy_share"`
	PayableAmount                 int64  `json:"payable_amount"`
	ReceptionCount                int64  `json:"reception_count"`
}

// RevenueByOrganizationResponse خروجی گزارش درآمد به تفکیک سازمان.
type RevenueByOrganizationResponse struct {
	Rows    []RevenueByOrganizationRow `json:"rows"`
	Summary RevenueByOrganizationRow   `json:"summary"`
	Meta    RevenueByOrganizationMeta    `json:"meta"`
}

// RevenueByOrganizationMeta اطلاعات فیلتر اعمال‌شده برای نمایش در چاپ.
type RevenueByOrganizationMeta struct {
	FromDate                    string `json:"from_date"`
	ToDate                      string `json:"to_date"`
	FromTime                    string `json:"from_time"`
	ToTime                      string `json:"to_time"`
	DetailSupplementary         bool   `json:"detail_supplementary"`
	HasSupplementaryFilter      bool   `json:"has_supplementary_filter"`
	OnlySupplementary           bool   `json:"only_supplementary"`
	FreeOrganizationName        string `json:"free_organization_name,omitempty"`
}

// DoctorPerformanceBaseQuery فیلترهای مشترک گزارش کارکرد پزشکان.
type DoctorPerformanceBaseQuery struct {
	FromDate         string `form:"from_date" binding:"required"`
	ToDate           string `form:"to_date" binding:"required"`
	FromTime         string `form:"from_time"`
	ToTime           string `form:"to_time"`
	DoctorIDs        string `form:"doctor_ids" binding:"required"`
	SeparateDoctors  bool   `form:"separate_doctors"`
}

// DoctorPerformanceByOrganizationQuery پارامترهای گزارش کارکرد پزشکان به تفکیک سازمان.
type DoctorPerformanceByOrganizationQuery struct {
	DoctorPerformanceBaseQuery
	BaseOrganizationIDs          string `form:"base_organization_ids"`
	SupplementaryOrganizationIDs string `form:"supplementary_organization_ids"`
	DetailSupplementary          bool   `form:"detail_supplementary"`
	OnlySupplementary            bool   `form:"only_supplementary"`
}

// DoctorPerformanceByServiceQuery پارامترهای گزارش کارکرد پزشکان به تفکیک خدمت.
type DoctorPerformanceByServiceQuery struct {
	DoctorPerformanceBaseQuery
	ServiceIDs string `form:"service_ids" binding:"required"`
}

// DoctorPerformancePatientListQuery پارامترهای لیست بیماران پزشکان.
type DoctorPerformancePatientListQuery struct {
	DoctorPerformanceBaseQuery
}

// RevenueByDoctorsQuery پارامترهای گزارش درآمد به تفکیک پزشکان.
type RevenueByDoctorsQuery struct {
	DoctorPerformanceBaseQuery
}

// RevenueByDoctorsRow سطر گزارش درآمد به تفکیک پزشک.
type RevenueByDoctorsRow struct {
	DoctorID           uint   `json:"doctor_id"`
	MedicalCode        string `json:"medical_code"`
	DoctorName         string `json:"doctor_name"`
	ServiceAmount      int64  `json:"service_amount"`
	ServiceTariff      int64  `json:"service_tariff"`
	OrganizationShare  int64  `json:"organization_share"`
	SupplementaryShare int64  `json:"supplementary_share"`
	SubsidyShare       int64  `json:"subsidy_share"`
	PayableAmount      int64  `json:"payable_amount"`
	ReceptionCount     int64  `json:"reception_count"`
}

// RevenueByDoctorsResponse خروجی گزارش درآمد به تفکیک پزشکان.
type RevenueByDoctorsResponse struct {
	Rows    []RevenueByDoctorsRow       `json:"rows"`
	Summary DoctorPerformanceSummaryRow `json:"summary"`
	Meta    DoctorPerformanceMeta       `json:"meta"`
}

// DoctorPerformanceSummaryRow جمع مالی مشترک گزارش‌های کارکرد پزشکان.
type DoctorPerformanceSummaryRow struct {
	ServiceAmount      int64  `json:"service_amount"`
	ServiceTariff      int64  `json:"service_tariff"`
	OrganizationShare  int64  `json:"organization_share"`
	SupplementaryShare int64  `json:"supplementary_share"`
	SubsidyShare       int64  `json:"subsidy_share"`
	PayableAmount      int64  `json:"payable_amount"`
	ReceptionCount     int64  `json:"reception_count"`
	RowLabel           string `json:"row_label,omitempty"`
}

// DoctorPerformanceByOrganizationRow سطر گزارش کارکرد به تفکیک سازمان.
type DoctorPerformanceByOrganizationRow struct {
	RowLabel                      string `json:"row_label"`
	BaseOrganizationID            uint   `json:"base_organization_id"`
	BaseOrganizationName          string `json:"base_organization_name"`
	SupplementaryOrganizationID   *uint  `json:"supplementary_organization_id,omitempty"`
	SupplementaryOrganizationName string `json:"supplementary_organization_name,omitempty"`
	ServiceAmount                 int64  `json:"service_amount"`
	ServiceTariff                 int64  `json:"service_tariff"`
	OrganizationShare             int64  `json:"organization_share"`
	SupplementaryShare            int64  `json:"supplementary_share"`
	SubsidyShare                  int64  `json:"subsidy_share"`
	PayableAmount                 int64  `json:"payable_amount"`
	ReceptionCount                int64  `json:"reception_count"`
}

// DoctorPerformanceByServiceRow سطر گزارش کارکرد به تفکیک خدمت.
type DoctorPerformanceByServiceRow struct {
	ServiceID          uint   `json:"service_id"`
	ServiceName        string `json:"service_name"`
	ServiceAmount      int64  `json:"service_amount"`
	ServiceTariff      int64  `json:"service_tariff"`
	OrganizationShare  int64  `json:"organization_share"`
	SupplementaryShare int64  `json:"supplementary_share"`
	SubsidyShare       int64  `json:"subsidy_share"`
	PayableAmount      int64  `json:"payable_amount"`
	ReceptionCount     int64  `json:"reception_count"`
}

// DoctorPerformancePatientRow سطر لیست بیماران پزشکان.
type DoctorPerformancePatientRow struct {
	PatientID          uint   `json:"patient_id"`
	PatientName        string `json:"patient_name"`
	NationalCode       string `json:"national_code"`
	ServiceAmount      int64  `json:"service_amount"`
	ServiceTariff      int64  `json:"service_tariff"`
	OrganizationShare  int64  `json:"organization_share"`
	SupplementaryShare int64  `json:"supplementary_share"`
	SubsidyShare       int64  `json:"subsidy_share"`
	PayableAmount      int64  `json:"payable_amount"`
	ServicesReceived   string `json:"services_received"`
}

// DoctorPerformanceOrganizationSection بخش گزارش به تفکیک سازمان برای یک پزشک.
type DoctorPerformanceOrganizationSection struct {
	DoctorID   uint                               `json:"doctor_id"`
	DoctorName string                             `json:"doctor_name"`
	Rows       []DoctorPerformanceByOrganizationRow `json:"rows"`
	Summary    DoctorPerformanceSummaryRow        `json:"summary"`
}

// DoctorPerformanceServiceSection بخش گزارش به تفکیک خدمت برای یک پزشک.
type DoctorPerformanceServiceSection struct {
	DoctorID   uint                          `json:"doctor_id"`
	DoctorName string                        `json:"doctor_name"`
	Rows       []DoctorPerformanceByServiceRow `json:"rows"`
	Summary    DoctorPerformanceSummaryRow     `json:"summary"`
}

// DoctorPerformancePatientSection بخش لیست بیماران برای یک پزشک.
type DoctorPerformancePatientSection struct {
	DoctorID   uint                            `json:"doctor_id"`
	DoctorName string                          `json:"doctor_name"`
	Rows       []DoctorPerformancePatientRow   `json:"rows"`
	Summary    DoctorPerformanceSummaryRow     `json:"summary"`
}

// DoctorPerformanceMeta اطلاعات فیلتر اعمال‌شده برای چاپ گزارش کارکرد پزشکان.
type DoctorPerformanceMeta struct {
	ReportType             string   `json:"report_type"`
	FromDate               string   `json:"from_date"`
	ToDate                 string   `json:"to_date"`
	FromTime               string   `json:"from_time"`
	ToTime                 string   `json:"to_time"`
	DoctorNames            []string `json:"doctor_names"`
	SeparateDoctors        bool     `json:"separate_doctors,omitempty"`
	DetailSupplementary    bool     `json:"detail_supplementary,omitempty"`
	HasSupplementaryFilter bool     `json:"has_supplementary_filter,omitempty"`
	OnlySupplementary      bool     `json:"only_supplementary,omitempty"`
	FreeOrganizationName   string   `json:"free_organization_name,omitempty"`
	ServiceNames           []string `json:"service_names,omitempty"`
}

// DoctorPerformanceByOrganizationResponse خروجی گزارش کارکرد به تفکیک سازمان.
type DoctorPerformanceByOrganizationResponse struct {
	Rows     []DoctorPerformanceByOrganizationRow   `json:"rows"`
	Summary  DoctorPerformanceSummaryRow          `json:"summary"`
	Sections []DoctorPerformanceOrganizationSection `json:"sections,omitempty"`
	Meta     DoctorPerformanceMeta                `json:"meta"`
}

// DoctorPerformanceByServiceResponse خروجی گزارش کارکرد به تفکیک خدمت.
type DoctorPerformanceByServiceResponse struct {
	Rows     []DoctorPerformanceByServiceRow   `json:"rows"`
	Summary  DoctorPerformanceSummaryRow       `json:"summary"`
	Sections []DoctorPerformanceServiceSection `json:"sections,omitempty"`
	Meta     DoctorPerformanceMeta             `json:"meta"`
}

// DoctorPerformancePatientListResponse خروجی لیست بیماران پزشکان.
type DoctorPerformancePatientListResponse struct {
	Rows     []DoctorPerformancePatientRow   `json:"rows"`
	Summary  DoctorPerformanceSummaryRow     `json:"summary"`
	Sections []DoctorPerformancePatientSection `json:"sections,omitempty"`
	Meta     DoctorPerformanceMeta           `json:"meta"`
}

// UserPerformanceBaseQuery فیلترهای مشترک گزارش کارکرد کاربران.
type UserPerformanceBaseQuery struct {
	FromDate           string `form:"from_date" binding:"required"`
	ToDate             string `form:"to_date" binding:"required"`
	FromTime           string `form:"from_time"`
	ToTime             string `form:"to_time"`
	UserIDs            string `form:"user_ids" binding:"required"`
	SeparateByPatient  bool   `form:"separate_by_patient"`
}

// UserPerformanceByReceptionQuery پارامترهای گزارش کارکرد بر اساس پذیرش.
type UserPerformanceByReceptionQuery struct {
	UserPerformanceBaseQuery
}

// UserPerformanceByTransactionQuery پارامترهای گزارش کارکرد بر اساس تراکنش مالی.
type UserPerformanceByTransactionQuery struct {
	UserPerformanceBaseQuery
}

// UserPerformanceMeta متادیتای چاپ گزارش کارکرد کاربران.
type UserPerformanceMeta struct {
	ReportType          string   `json:"report_type"`
	FromDate            string   `json:"from_date"`
	ToDate              string   `json:"to_date"`
	FromTime            string   `json:"from_time"`
	ToTime              string   `json:"to_time"`
	UserNames           []string `json:"user_names"`
	SeparateByPatient   bool     `json:"separate_by_patient,omitempty"`
}

// UserPerformanceReceptionRow سطر تجمیعی کارکرد کاربر بر اساس پذیرش.
type UserPerformanceReceptionRow struct {
	UserID             uint   `json:"user_id"`
	UserName           string `json:"user_name"`
	ServiceAmount      int64  `json:"service_amount"`
	ServiceTariff      int64  `json:"service_tariff"`
	OrganizationShare  int64  `json:"organization_share"`
	SupplementaryShare int64  `json:"supplementary_share"`
	SubsidyShare       int64  `json:"subsidy_share"`
	PayableAmount      int64  `json:"payable_amount"`
	ReceptionCount     int64  `json:"reception_count"`
}

// UserPerformanceReceptionDetailRow سطر ریز پذیرش در گزارش کارکرد کاربر.
type UserPerformanceReceptionDetailRow struct {
	ReceptionID        uint   `json:"reception_id"`
	PatientName        string `json:"patient_name"`
	NationalCode       string `json:"national_code"`
	ServiceAmount      int64  `json:"service_amount"`
	ServiceTariff      int64  `json:"service_tariff"`
	OrganizationShare  int64  `json:"organization_share"`
	SupplementaryShare int64  `json:"supplementary_share"`
	SubsidyShare       int64  `json:"subsidy_share"`
	PayableAmount      int64  `json:"payable_amount"`
	ServicesReceived   string `json:"services_received"`
}

// UserPerformanceTransactionRow سطر تجمیعی کارکرد کاربر بر اساس تراکنش مالی.
type UserPerformanceTransactionRow struct {
	UserID              uint   `json:"user_id"`
	UserName            string `json:"user_name"`
	TotalReceived       int64  `json:"total_received"`
	TotalPaid           int64  `json:"total_paid"`
	ReceivedTxCount     int64  `json:"received_tx_count"`
	PaidTxCount         int64  `json:"paid_tx_count"`
	TotalTxCount        int64  `json:"total_tx_count"`
}

// UserPerformanceTransactionDetailRow سطر ریز تراکنش مالی کاربر.
type UserPerformanceTransactionDetailRow struct {
	TransactionID uint   `json:"transaction_id"`
	FileNumber    string `json:"file_number"`
	PatientName   string `json:"patient_name"`
	NationalCode  string `json:"national_code"`
	Received      int64  `json:"received"`
	Paid          int64  `json:"paid"`
}

// UserPerformanceReceptionSummary جمع مالی گزارش پذیرش کاربران.
type UserPerformanceReceptionSummary struct {
	ServiceAmount      int64  `json:"service_amount"`
	ServiceTariff      int64  `json:"service_tariff"`
	OrganizationShare  int64  `json:"organization_share"`
	SupplementaryShare int64  `json:"supplementary_share"`
	SubsidyShare       int64  `json:"subsidy_share"`
	PayableAmount      int64  `json:"payable_amount"`
	ReceptionCount     int64  `json:"reception_count"`
	RowLabel           string `json:"row_label,omitempty"`
}

// UserPerformanceTransactionSummary جمع گزارش تراکنش کاربران.
type UserPerformanceTransactionSummary struct {
	TotalReceived   int64  `json:"total_received"`
	TotalPaid       int64  `json:"total_paid"`
	ReceivedTxCount int64  `json:"received_tx_count"`
	PaidTxCount     int64  `json:"paid_tx_count"`
	TotalTxCount    int64  `json:"total_tx_count"`
	RowLabel        string `json:"row_label,omitempty"`
}

// UserPerformanceReceptionSection بخش گزارش پذیرش برای یک کاربر.
type UserPerformanceReceptionSection struct {
	UserID   uint                                `json:"user_id"`
	UserName string                              `json:"user_name"`
	Rows     []UserPerformanceReceptionDetailRow `json:"rows"`
	Summary  UserPerformanceReceptionSummary     `json:"summary"`
}

// UserPerformanceTransactionSection بخش گزارش تراکنش برای یک کاربر.
type UserPerformanceTransactionSection struct {
	UserID   uint                                  `json:"user_id"`
	UserName string                                `json:"user_name"`
	Rows     []UserPerformanceTransactionDetailRow `json:"rows"`
	Summary  UserPerformanceTransactionSummary     `json:"summary"`
}

// UserPerformanceByReceptionResponse خروجی گزارش کارکرد بر اساس پذیرش.
type UserPerformanceByReceptionResponse struct {
	Rows     []UserPerformanceReceptionRow     `json:"rows"`
	Summary  UserPerformanceReceptionSummary   `json:"summary"`
	Sections []UserPerformanceReceptionSection `json:"sections,omitempty"`
	Meta     UserPerformanceMeta               `json:"meta"`
}

// UserPerformanceByTransactionResponse خروجی گزارش کارکرد بر اساس تراکنش مالی.
type UserPerformanceByTransactionResponse struct {
	Rows     []UserPerformanceTransactionRow     `json:"rows"`
	Summary  UserPerformanceTransactionSummary   `json:"summary"`
	Sections []UserPerformanceTransactionSection `json:"sections,omitempty"`
	Meta     UserPerformanceMeta                 `json:"meta"`
}
