package patient

// CreateRequest بدنه ایجاد بیمار.
type CreateRequest struct {
	FirstName         string  `json:"first_name" binding:"required"`
	LastName          string  `json:"last_name" binding:"required"`
	NationalCode      string  `json:"national_code" binding:"required"`
	BirthDate         string  `json:"birth_date" binding:"required"`
	Address           *string `json:"address"`
	HomePhoneNumber   *string `json:"home_phone_number"`
	MobilePhoneNumber *string `json:"mobile_phone_number"`
	FileNumber        string  `json:"file_number" binding:"required"`
	Sex               bool    `json:"sex"`
}

// UpdateRequest بدنه بروزرسانی بیمار.
type UpdateRequest struct {
	FirstName         string  `json:"first_name" binding:"required"`
	LastName          string  `json:"last_name" binding:"required"`
	NationalCode      string  `json:"national_code" binding:"required"`
	BirthDate         string  `json:"birth_date" binding:"required"`
	Address           *string `json:"address"`
	HomePhoneNumber   *string `json:"home_phone_number"`
	MobilePhoneNumber *string `json:"mobile_phone_number"`
	FileNumber        string  `json:"file_number" binding:"required"`
	Sex               bool    `json:"sex"`
}

// SearchRequest پارامترهای جستجوی بیمار از query string.
type SearchRequest struct {
	FirstName         string `form:"first_name"`
	LastName          string `form:"last_name"`
	NationalCode      string `form:"national_code"`
	BirthDate         string `form:"birth_date"`
	Address           string `form:"address"`
	HomePhoneNumber   string `form:"home_phone_number"`
	MobilePhoneNumber string `form:"mobile_phone_number"`
	FileNumber        string `form:"file_number"`
	Sex               *bool  `form:"sex"`
}

// Response پاسخ API بیمار (بدون IntegrityHash).
type Response struct {
	ID                uint    `json:"id"`
	FirstName         string  `json:"first_name"`
	LastName          string  `json:"last_name"`
	NationalCode      string  `json:"national_code"`
	BirthDate         string  `json:"birth_date"`
	Address           *string `json:"address"`
	HomePhoneNumber   *string `json:"home_phone_number"`
	MobilePhoneNumber *string `json:"mobile_phone_number"`
	FileNumber        string  `json:"file_number"`
	Sex               bool    `json:"sex"`
}
