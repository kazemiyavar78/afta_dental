package report

import (
	"time"

	"github.com/tpdenta/afta-reception/internal/reception"
	"github.com/tpdenta/afta-reception/internal/user"
	"github.com/tpdenta/afta-reception/internal/wallet"
	"gorm.io/gorm"
)

// revenueAggRow نتیجه تجمیع SQL گزارش درآمد.
type revenueAggRow struct {
	InsuranceID           uint  `gorm:"column:insurance_id"`
	AdditionalInsuranceID *uint `gorm:"column:additional_insurance_id"`
	ServiceAmount         int64 `gorm:"column:service_amount"`
	ServiceTariff         int64 `gorm:"column:service_tariff"`
	OrganizationShare     int64 `gorm:"column:organization_share"`
	SupplementaryShare    int64 `gorm:"column:supplementary_share"`
	SubsidyShare          int64 `gorm:"column:subsidy_share"`
	ReceptionCount        int64 `gorm:"column:reception_count"`
}

// doctorServiceAggRow نتیجه تجمیع گزارش کارکرد به تفکیک خدمت.
type doctorServiceAggRow struct {
	ServiceID          uint   `gorm:"column:service_id"`
	ServiceName        string `gorm:"column:service_name"`
	ServiceAmount      int64  `gorm:"column:service_amount"`
	ServiceTariff      int64  `gorm:"column:service_tariff"`
	OrganizationShare  int64  `gorm:"column:organization_share"`
	SupplementaryShare int64  `gorm:"column:supplementary_share"`
	SubsidyShare       int64  `gorm:"column:subsidy_share"`
	ReceptionCount     int64  `gorm:"column:reception_count"`
}

// doctorPatientAggRow نتیجه تجمیع لیست بیماران پزشکان.
type doctorPatientAggRow struct {
	PatientID          uint   `gorm:"column:patient_id"`
	PatientFirstName   string `gorm:"column:patient_first_name"`
	PatientLastName    string `gorm:"column:patient_last_name"`
	NationalCode       string `gorm:"column:national_code"`
	ServiceAmount      int64  `gorm:"column:service_amount"`
	ServiceTariff      int64  `gorm:"column:service_tariff"`
	OrganizationShare  int64  `gorm:"column:organization_share"`
	SupplementaryShare int64  `gorm:"column:supplementary_share"`
	SubsidyShare       int64  `gorm:"column:subsidy_share"`
	ServicesReceived   string `gorm:"column:services_received"`
}

// userInfoRow اطلاعات پایه کاربر برای گزارش کارکرد.
type userInfoRow struct {
	ID       uint   `gorm:"column:ID"`
	Name     string `gorm:"column:Name"`
	Family   string `gorm:"column:Family"`
	UserType string `gorm:"column:UserType"`
}

// userReceptionAggRow نتیجه تجمیع پذیرش یک کاربر.
type userReceptionAggRow struct {
	ServiceAmount      int64 `gorm:"column:service_amount"`
	ServiceTariff      int64 `gorm:"column:service_tariff"`
	OrganizationShare  int64 `gorm:"column:organization_share"`
	SupplementaryShare int64 `gorm:"column:supplementary_share"`
	SubsidyShare       int64 `gorm:"column:subsidy_share"`
	ReceptionCount     int64 `gorm:"column:reception_count"`
}

// userReceptionDetailAggRow نتیجه ریز پذیرش کاربر.
type userReceptionDetailAggRow struct {
	ReceptionID        uint   `gorm:"column:reception_id"`
	PatientFirstName   string `gorm:"column:patient_first_name"`
	PatientLastName    string `gorm:"column:patient_last_name"`
	NationalCode       string `gorm:"column:national_code"`
	ServiceAmount      int64  `gorm:"column:service_amount"`
	ServiceTariff      int64  `gorm:"column:service_tariff"`
	OrganizationShare  int64  `gorm:"column:organization_share"`
	SupplementaryShare int64 `gorm:"column:supplementary_share"`
	SubsidyShare       int64  `gorm:"column:subsidy_share"`
	ServicesReceived   string `gorm:"column:services_received"`
}

// userTransactionAggRow نتیجه تجمیع تراکنش‌های یک کاربر.
type userTransactionAggRow struct {
	TotalReceived   int64 `gorm:"column:total_received"`
	TotalPaid       int64 `gorm:"column:total_paid"`
	ReceivedTxCount int64 `gorm:"column:received_tx_count"`
	PaidTxCount     int64 `gorm:"column:paid_tx_count"`
	TotalTxCount    int64 `gorm:"column:total_tx_count"`
}

// userTransactionDetailAggRow نتیجه ریز تراکنش مالی کاربر.
type userTransactionDetailAggRow struct {
	TransactionID    uint   `gorm:"column:transaction_id"`
	FileNumber       string `gorm:"column:file_number"`
	PatientFirstName string `gorm:"column:patient_first_name"`
	PatientLastName  string `gorm:"column:patient_last_name"`
	NationalCode     string `gorm:"column:national_code"`
	Received         int64  `gorm:"column:received"`
	Paid             int64  `gorm:"column:paid"`
}

// revenueByDoctorAggRow نتیجه تجمیع درآمد به تفکیک پزشک.
type revenueByDoctorAggRow struct {
	DoctorID         uint    `gorm:"column:doctor_id"`
	MedicalCode      *string `gorm:"column:medical_code"`
	DoctorFirstName  string  `gorm:"column:doctor_first_name"`
	DoctorLastName   string  `gorm:"column:doctor_last_name"`
	ServiceAmount    int64   `gorm:"column:service_amount"`
	ServiceTariff    int64   `gorm:"column:service_tariff"`
	OrganizationShare int64  `gorm:"column:organization_share"`
	SupplementaryShare int64 `gorm:"column:supplementary_share"`
	SubsidyShare     int64   `gorm:"column:subsidy_share"`
	ReceptionCount   int64   `gorm:"column:reception_count"`
}

// Repository اینترفیس دسترسی داده گزارش‌ها.
type Repository interface {
	AggregateRevenueByOrganization(
		from, to time.Time,
		baseIDs, suppIDs []uint,
		detailSupplementary bool,
		onlySupplementary bool,
	) ([]revenueAggRow, error)
	AggregateDoctorPerformanceByOrganization(
		from, to time.Time,
		doctorIDs, baseIDs, suppIDs []uint,
		detailSupplementary bool,
		onlySupplementary bool,
	) ([]revenueAggRow, error)
	AggregateDoctorPerformanceByService(
		from, to time.Time,
		doctorIDs, serviceIDs []uint,
	) ([]doctorServiceAggRow, error)
	AggregateDoctorPerformancePatientList(
		from, to time.Time,
		doctorIDs []uint,
	) ([]doctorPatientAggRow, error)
	AggregateRevenueByDoctors(
		from, to time.Time,
		doctorIDs []uint,
	) ([]revenueByDoctorAggRow, error)
	FindDoctorNames(ids []uint) (map[uint]string, error)
	FindUserInfos(ids []uint) (map[uint]userInfoRow, error)
	AggregateUserPerformanceByReception(
		from, to time.Time,
		userID uint,
		userType user.UserType,
	) (*userReceptionAggRow, error)
	ListUserPerformanceReceptions(
		from, to time.Time,
		userID uint,
		userType user.UserType,
	) ([]userReceptionDetailAggRow, error)
	AggregateUserPerformanceByTransaction(
		from, to time.Time,
		userID uint,
	) (*userTransactionAggRow, error)
	ListUserPerformanceTransactions(
		from, to time.Time,
		userID uint,
	) ([]userTransactionDetailAggRow, error)
}

type gormRepo struct {
	db *gorm.DB
}

// NewRepository نمونه Repository گزارش می‌سازد.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepo{db: db}
}

// AggregateRevenueByOrganization درآمد پذیرش‌ها را بر اساس سازمان پایه (و در صورت نیاز تکمیلی) تجمیع می‌کند.
func (r *gormRepo) AggregateRevenueByOrganization(
	from, to time.Time,
	baseIDs, suppIDs []uint,
	detailSupplementary bool,
	onlySupplementary bool,
) ([]revenueAggRow, error) {
	if onlySupplementary {
		return r.aggregateOnlySupplementary(from, to, baseIDs, suppIDs)
	}

	selectCols := `
		r.InsuranceID AS insurance_id,
		COALESCE(SUM(rs.ServiceAmount), 0) AS service_amount,
		COALESCE(SUM(rs.ServiceTariff), 0) AS service_tariff,
		COALESCE(SUM(rs.ServiceOrganizationShare), 0) AS organization_share,
		COALESCE(SUM(rs.ServiceSupplementaryInsuranceShare), 0) AS supplementary_share,
		COALESCE(SUM(rs.ServiceSubsidyShare), 0) AS subsidy_share,
		COUNT(DISTINCT r.ID) AS reception_count
	`

	groupBy := "r.InsuranceID"
	hasSuppFilter := len(suppIDs) > 0

	if hasSuppFilter && detailSupplementary {
		selectCols = `
			r.InsuranceID AS insurance_id,
			r.AdditionalInsuranceID AS additional_insurance_id,
			COALESCE(SUM(rs.ServiceAmount), 0) AS service_amount,
			COALESCE(SUM(rs.ServiceTariff), 0) AS service_tariff,
			COALESCE(SUM(rs.ServiceOrganizationShare), 0) AS organization_share,
			COALESCE(SUM(rs.ServiceSupplementaryInsuranceShare), 0) AS supplementary_share,
			COALESCE(SUM(rs.ServiceSubsidyShare), 0) AS subsidy_share,
			COUNT(DISTINCT r.ID) AS reception_count
		`
		groupBy = "r.InsuranceID, r.AdditionalInsuranceID"
	}

	q := r.baseReceptionQuery(from, to).
		Select(selectCols).
		Where("r.InsuranceID IN ?", baseIDs)

	if !hasSuppFilter {
		q = q.Where("r.AdditionalInsuranceID IS NULL")
	} else {
		q = q.Where("(r.AdditionalInsuranceID IS NULL OR r.AdditionalInsuranceID IN ?)", suppIDs)
	}

	var rows []revenueAggRow
	err := q.Group(groupBy).Order(groupBy).Scan(&rows).Error
	return rows, err
}

// aggregateOnlySupplementary درآمد پذیرش‌های دارای سازمان پایه آزاد + تکمیلی را به تفکیک تکمیلی تجمیع می‌کند.
func (r *gormRepo) aggregateOnlySupplementary(
	from, to time.Time,
	baseIDs, suppIDs []uint,
) ([]revenueAggRow, error) {
	selectCols := `
		r.InsuranceID AS insurance_id,
		r.AdditionalInsuranceID AS additional_insurance_id,
		COALESCE(SUM(rs.ServiceAmount), 0) AS service_amount,
		COALESCE(SUM(rs.ServiceTariff), 0) AS service_tariff,
		COALESCE(SUM(rs.ServiceOrganizationShare), 0) AS organization_share,
		COALESCE(SUM(rs.ServiceSupplementaryInsuranceShare), 0) AS supplementary_share,
		COALESCE(SUM(rs.ServiceSubsidyShare), 0) AS subsidy_share,
		COUNT(DISTINCT r.ID) AS reception_count
	`

	q := r.baseReceptionQuery(from, to).
		Select(selectCols).
		Where("r.InsuranceID IN ?", baseIDs).
		Where("r.AdditionalInsuranceID IS NOT NULL")

	if len(suppIDs) > 0 {
		q = q.Where("r.AdditionalInsuranceID IN ?", suppIDs)
	}

	var rows []revenueAggRow
	err := q.Group("r.AdditionalInsuranceID, r.InsuranceID").Order("r.AdditionalInsuranceID").Scan(&rows).Error
	return rows, err
}

// baseReceptionQuery کوئری پایه پذیرش‌های ذخیره‌شده در بازه زمانی را می‌سازد.
func (r *gormRepo) baseReceptionQuery(from, to time.Time) *gorm.DB {
	return r.db.Table("Receptions AS r").
		Joins("INNER JOIN ReceptionServices rs ON rs.ReceptionID = r.ID AND rs.deleted_at IS NULL").
		Where("r.deleted_at IS NULL").
		Where("r.Status = ?", string(reception.ReceptionStatusSaved)).
		Where("r.ReceptionDate >= ? AND r.ReceptionDate <= ?", from, to)
}

// baseDoctorReceptionQuery کوئری پایه پذیرش‌های پزشکان انتخاب‌شده را می‌سازد.
func (r *gormRepo) baseDoctorReceptionQuery(from, to time.Time, doctorIDs []uint) *gorm.DB {
	return r.baseReceptionQuery(from, to).
		Where("r.DoctorID IN ?", doctorIDs)
}

// AggregateDoctorPerformanceByOrganization کارکرد پزشکان را به تفکیک سازمان تجمیع می‌کند.
func (r *gormRepo) AggregateDoctorPerformanceByOrganization(
	from, to time.Time,
	doctorIDs, baseIDs, suppIDs []uint,
	detailSupplementary bool,
	onlySupplementary bool,
) ([]revenueAggRow, error) {
	if onlySupplementary {
		return r.aggregateDoctorOnlySupplementary(from, to, doctorIDs, baseIDs, suppIDs)
	}

	selectCols := `
		r.InsuranceID AS insurance_id,
		COALESCE(SUM(rs.ServiceAmount), 0) AS service_amount,
		COALESCE(SUM(rs.ServiceTariff), 0) AS service_tariff,
		COALESCE(SUM(rs.ServiceOrganizationShare), 0) AS organization_share,
		COALESCE(SUM(rs.ServiceSupplementaryInsuranceShare), 0) AS supplementary_share,
		COALESCE(SUM(rs.ServiceSubsidyShare), 0) AS subsidy_share,
		COUNT(DISTINCT r.ID) AS reception_count
	`

	groupBy := "r.InsuranceID"
	hasSuppFilter := len(suppIDs) > 0

	if hasSuppFilter && detailSupplementary {
		selectCols = `
			r.InsuranceID AS insurance_id,
			r.AdditionalInsuranceID AS additional_insurance_id,
			COALESCE(SUM(rs.ServiceAmount), 0) AS service_amount,
			COALESCE(SUM(rs.ServiceTariff), 0) AS service_tariff,
			COALESCE(SUM(rs.ServiceOrganizationShare), 0) AS organization_share,
			COALESCE(SUM(rs.ServiceSupplementaryInsuranceShare), 0) AS supplementary_share,
			COALESCE(SUM(rs.ServiceSubsidyShare), 0) AS subsidy_share,
			COUNT(DISTINCT r.ID) AS reception_count
		`
		groupBy = "r.InsuranceID, r.AdditionalInsuranceID"
	}

	q := r.baseDoctorReceptionQuery(from, to, doctorIDs).
		Select(selectCols).
		Where("r.InsuranceID IN ?", baseIDs)

	if !hasSuppFilter {
		q = q.Where("r.AdditionalInsuranceID IS NULL")
	} else {
		q = q.Where("(r.AdditionalInsuranceID IS NULL OR r.AdditionalInsuranceID IN ?)", suppIDs)
	}

	var rows []revenueAggRow
	err := q.Group(groupBy).Order(groupBy).Scan(&rows).Error
	return rows, err
}

// aggregateDoctorOnlySupplementary کارکرد پزشکان را در حالت «فقط تکمیلی» تجمیع می‌کند.
func (r *gormRepo) aggregateDoctorOnlySupplementary(
	from, to time.Time,
	doctorIDs, baseIDs, suppIDs []uint,
) ([]revenueAggRow, error) {
	selectCols := `
		r.InsuranceID AS insurance_id,
		r.AdditionalInsuranceID AS additional_insurance_id,
		COALESCE(SUM(rs.ServiceAmount), 0) AS service_amount,
		COALESCE(SUM(rs.ServiceTariff), 0) AS service_tariff,
		COALESCE(SUM(rs.ServiceOrganizationShare), 0) AS organization_share,
		COALESCE(SUM(rs.ServiceSupplementaryInsuranceShare), 0) AS supplementary_share,
		COALESCE(SUM(rs.ServiceSubsidyShare), 0) AS subsidy_share,
		COUNT(DISTINCT r.ID) AS reception_count
	`

	q := r.baseDoctorReceptionQuery(from, to, doctorIDs).
		Select(selectCols).
		Where("r.InsuranceID IN ?", baseIDs).
		Where("r.AdditionalInsuranceID IS NOT NULL")

	if len(suppIDs) > 0 {
		q = q.Where("r.AdditionalInsuranceID IN ?", suppIDs)
	}

	var rows []revenueAggRow
	err := q.Group("r.AdditionalInsuranceID, r.InsuranceID").Order("r.AdditionalInsuranceID").Scan(&rows).Error
	return rows, err
}

// AggregateDoctorPerformanceByService کارکرد پزشکان را به تفکیک خدمت تجمیع می‌کند.
func (r *gormRepo) AggregateDoctorPerformanceByService(
	from, to time.Time,
	doctorIDs, serviceIDs []uint,
) ([]doctorServiceAggRow, error) {
	selectCols := `
		rs.ServiceID AS service_id,
		rs.ServiceName AS service_name,
		COALESCE(SUM(rs.ServiceAmount), 0) AS service_amount,
		COALESCE(SUM(rs.ServiceTariff), 0) AS service_tariff,
		COALESCE(SUM(rs.ServiceOrganizationShare), 0) AS organization_share,
		COALESCE(SUM(rs.ServiceSupplementaryInsuranceShare), 0) AS supplementary_share,
		COALESCE(SUM(rs.ServiceSubsidyShare), 0) AS subsidy_share,
		COUNT(DISTINCT r.ID) AS reception_count
	`

	q := r.baseDoctorReceptionQuery(from, to, doctorIDs).
		Select(selectCols).
		Where("rs.ServiceID IN ?", serviceIDs)

	var rows []doctorServiceAggRow
	err := q.Group("rs.ServiceID, rs.ServiceName").Order("rs.ServiceName").Scan(&rows).Error
	return rows, err
}

// AggregateDoctorPerformancePatientList لیست بیماران پذیرش‌شده توسط پزشکان را برمی‌گرداند.
func (r *gormRepo) AggregateDoctorPerformancePatientList(
	from, to time.Time,
	doctorIDs []uint,
) ([]doctorPatientAggRow, error) {
	selectCols := `
		p.ID AS patient_id,
		p.first_name AS patient_first_name,
		p.last_name AS patient_last_name,
		p.national_code AS national_code,
		COALESCE(SUM(rs.ServiceAmount), 0) AS service_amount,
		COALESCE(SUM(rs.ServiceTariff), 0) AS service_tariff,
		COALESCE(SUM(rs.ServiceOrganizationShare), 0) AS organization_share,
		COALESCE(SUM(rs.ServiceSupplementaryInsuranceShare), 0) AS supplementary_share,
		COALESCE(SUM(rs.ServiceSubsidyShare), 0) AS subsidy_share,
		STRING_AGG(rs.ServiceName, N'، ') WITHIN GROUP (ORDER BY rs.ServiceName) AS services_received
	`

	q := r.baseDoctorReceptionQuery(from, to, doctorIDs).
		Joins("INNER JOIN patients p ON p.ID = r.PatientID AND p.deleted_at IS NULL").
		Select(selectCols)

	var rows []doctorPatientAggRow
	err := q.Group("p.ID, p.first_name, p.last_name, p.national_code").
		Order("p.last_name, p.first_name").
		Scan(&rows).Error
	return rows, err
}

// AggregateRevenueByDoctors درآمد پذیرش‌ها را به تفکیک پزشک تجمیع می‌کند.
func (r *gormRepo) AggregateRevenueByDoctors(
	from, to time.Time,
	doctorIDs []uint,
) ([]revenueByDoctorAggRow, error) {
	selectCols := `
		r.DoctorID AS doctor_id,
		u.MedicalCode AS medical_code,
		u.Name AS doctor_first_name,
		u.Family AS doctor_last_name,
		COALESCE(SUM(rs.ServiceAmount), 0) AS service_amount,
		COALESCE(SUM(rs.ServiceTariff), 0) AS service_tariff,
		COALESCE(SUM(rs.ServiceOrganizationShare), 0) AS organization_share,
		COALESCE(SUM(rs.ServiceSupplementaryInsuranceShare), 0) AS supplementary_share,
		COALESCE(SUM(rs.ServiceSubsidyShare), 0) AS subsidy_share,
		COUNT(DISTINCT r.ID) AS reception_count
	`

	q := r.baseDoctorReceptionQuery(from, to, doctorIDs).
		Joins("INNER JOIN Users u ON u.ID = r.DoctorID").
		Where("r.DoctorID IS NOT NULL").
		Select(selectCols)

	var rows []revenueByDoctorAggRow
	err := q.Group("r.DoctorID, u.MedicalCode, u.Name, u.Family").
		Order("u.Family, u.Name").
		Scan(&rows).Error
	return rows, err
}

// applyUserReceptionFilter شرط تطبیق پذیرش با نقش کاربر را روی کوئری اعمال می‌کند.
func applyUserReceptionFilter(q *gorm.DB, userID uint, userType user.UserType) *gorm.DB {
	switch userType {
	case user.UserTypeDoctor, user.UserTypeSpecialist:
		return q.Where("r.DoctorID = ?", userID)
	case user.UserTypeAssistant:
		return q.Where("r.AssistantID = ?", userID)
	default:
		return q.Where("r.RegisteredByID = ?", userID)
	}
}

// baseUserReceptionQuery کوئری پایه پذیرش‌های مرتبط با یک کاربر را می‌سازد.
func (r *gormRepo) baseUserReceptionQuery(from, to time.Time, userID uint, userType user.UserType) *gorm.DB {
	q := r.baseReceptionQuery(from, to)
	return applyUserReceptionFilter(q, userID, userType)
}

// baseUserTransactionQuery کوئری پایه تراکنش‌های کیف پول یک کاربر را می‌سازد.
func (r *gormRepo) baseUserTransactionQuery(from, to time.Time, userID uint) *gorm.DB {
	return r.db.Table("wallet_transactions AS t").
		Where("t.deleted_at IS NULL").
		Where("t.category = ?", string(wallet.CategoryWallet)).
		Where("t.performed_by = ?", userID).
		Where("COALESCE(t.paid_at, t.created_at) >= ? AND COALESCE(t.paid_at, t.created_at) <= ?", from, to)
}

// FindUserInfos اطلاعات پایه کاربران را بر اساس شناسه برمی‌گرداند.
func (r *gormRepo) FindUserInfos(ids []uint) (map[uint]userInfoRow, error) {
	if len(ids) == 0 {
		return map[uint]userInfoRow{}, nil
	}
	var users []userInfoRow
	err := r.db.Table("Users").
		Select("ID, Name, Family, UserType").
		Where("ID IN ?", ids).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	out := make(map[uint]userInfoRow, len(users))
	for _, u := range users {
		out[u.ID] = u
	}
	return out, nil
}

// AggregateUserPerformanceByReception کارکرد یک کاربر را بر اساس پذیرش تجمیع می‌کند.
func (r *gormRepo) AggregateUserPerformanceByReception(
	from, to time.Time,
	userID uint,
	userType user.UserType,
) (*userReceptionAggRow, error) {
	selectCols := `
		COALESCE(SUM(rs.ServiceAmount), 0) AS service_amount,
		COALESCE(SUM(rs.ServiceTariff), 0) AS service_tariff,
		COALESCE(SUM(rs.ServiceOrganizationShare), 0) AS organization_share,
		COALESCE(SUM(rs.ServiceSupplementaryInsuranceShare), 0) AS supplementary_share,
		COALESCE(SUM(rs.ServiceSubsidyShare), 0) AS subsidy_share,
		COUNT(DISTINCT r.ID) AS reception_count
	`
	var row userReceptionAggRow
	err := r.baseUserReceptionQuery(from, to, userID, userType).
		Select(selectCols).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ListUserPerformanceReceptions تمام پذیرش‌های یک کاربر را به‌صورت ریز برمی‌گرداند.
func (r *gormRepo) ListUserPerformanceReceptions(
	from, to time.Time,
	userID uint,
	userType user.UserType,
) ([]userReceptionDetailAggRow, error) {
	selectCols := `
		r.ID AS reception_id,
		p.first_name AS patient_first_name,
		p.last_name AS patient_last_name,
		p.national_code AS national_code,
		COALESCE(SUM(rs.ServiceAmount), 0) AS service_amount,
		COALESCE(SUM(rs.ServiceTariff), 0) AS service_tariff,
		COALESCE(SUM(rs.ServiceOrganizationShare), 0) AS organization_share,
		COALESCE(SUM(rs.ServiceSupplementaryInsuranceShare), 0) AS supplementary_share,
		COALESCE(SUM(rs.ServiceSubsidyShare), 0) AS subsidy_share,
		STRING_AGG(rs.ServiceName, N'، ') WITHIN GROUP (ORDER BY rs.ServiceName) AS services_received
	`
	var rows []userReceptionDetailAggRow
	err := r.baseUserReceptionQuery(from, to, userID, userType).
		Joins("INNER JOIN patients p ON p.ID = r.PatientID AND p.deleted_at IS NULL").
		Select(selectCols).
		Group("r.ID, p.first_name, p.last_name, p.national_code, r.ReceptionDate").
		Order("r.ReceptionDate, r.ID").
		Scan(&rows).Error
	return rows, err
}

// AggregateUserPerformanceByTransaction تراکنش‌های مالی یک کاربر را تجمیع می‌کند.
func (r *gormRepo) AggregateUserPerformanceByTransaction(
	from, to time.Time,
	userID uint,
) (*userTransactionAggRow, error) {
	selectCols := `
		COALESCE(SUM(CASE WHEN t.action = 'charge' THEN t.amount ELSE 0 END), 0) AS total_received,
		COALESCE(SUM(CASE WHEN t.action IN ('payment', 'refund') THEN t.amount ELSE 0 END), 0) AS total_paid,
		COALESCE(SUM(CASE WHEN t.action = 'charge' THEN 1 ELSE 0 END), 0) AS received_tx_count,
		COALESCE(SUM(CASE WHEN t.action IN ('payment', 'refund') THEN 1 ELSE 0 END), 0) AS paid_tx_count,
		COUNT(*) AS total_tx_count
	`
	var row userTransactionAggRow
	err := r.baseUserTransactionQuery(from, to, userID).
		Select(selectCols).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ListUserPerformanceTransactions تمام تراکنش‌های مالی یک کاربر را به‌صورت ریز برمی‌گرداند.
func (r *gormRepo) ListUserPerformanceTransactions(
	from, to time.Time,
	userID uint,
) ([]userTransactionDetailAggRow, error) {
	selectCols := `
		t.ID AS transaction_id,
		p.file_number AS file_number,
		p.first_name AS patient_first_name,
		p.last_name AS patient_last_name,
		p.national_code AS national_code,
		CASE WHEN t.action = 'charge' THEN t.amount ELSE 0 END AS received,
		CASE WHEN t.action IN ('payment', 'refund') THEN t.amount ELSE 0 END AS paid
	`
	var rows []userTransactionDetailAggRow
	err := r.baseUserTransactionQuery(from, to, userID).
		Joins("INNER JOIN patients p ON p.ID = t.file_id AND p.deleted_at IS NULL").
		Select(selectCols).
		Order("COALESCE(t.paid_at, t.created_at), t.ID").
		Scan(&rows).Error
	return rows, err
}

// FindDoctorNames نام پزشکان را بر اساس شناسه برمی‌گرداند.
func (r *gormRepo) FindDoctorNames(ids []uint) (map[uint]string, error) {
	if len(ids) == 0 {
		return map[uint]string{}, nil
	}
	type row struct {
		ID     uint   `gorm:"column:ID"`
		Name   string `gorm:"column:Name"`
		Family string `gorm:"column:Family"`
	}
	var users []row
	err := r.db.Table("Users").
		Select("ID, Name, Family").
		Where("ID IN ?", ids).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	out := make(map[uint]string, len(users))
	for _, u := range users {
		out[u.ID] = u.Name + " " + u.Family
	}
	return out, nil
}
