package services

// CreateRequest بدنه ایجاد خدمت.
type CreateRequest struct {
	ServiceCode             string  `json:"service_code" binding:"required"`
	Name                    string  `json:"name" binding:"required"`
	TechnicalCoefficient    float64 `json:"technical_coefficient"`
	ProfessionalCoefficient float64 `json:"professional_coefficient"`
	ConsumptionCoefficient  float64 `json:"consumption_coefficient"`
	ServiceRate             int `json:"service_rate"`
	ServiceTariff           int `json:"service_tariff"`
	InternationalCode       string  `json:"international_code"`
	DefaultCount            int     `json:"default_count"`
	MaximumCount            int     `json:"maximum_count"`
	ServiceFeatures         string  `json:"service_features"`
	IsActive                bool    `json:"is_active"`
	IsDentalDirection       bool    `json:"is_dental_direction"`
	HasTooth                bool    `json:"has_tooth"`
	AllowMultipleUse        bool    `json:"allow_multiple_use"`
}

// UpdateRequest بدنه بروزرسانی خدمت.
type UpdateRequest struct {
	ServiceCode             string  `json:"service_code" binding:"required"`
	Name                    string  `json:"name" binding:"required"`
	TechnicalCoefficient    float64 `json:"technical_coefficient"`
	ProfessionalCoefficient float64 `json:"professional_coefficient"`
	ConsumptionCoefficient  float64 `json:"consumption_coefficient"`
	ServiceRate             int `json:"service_rate"`
	ServiceTariff           int `json:"service_tariff"`
	InternationalCode       string  `json:"international_code"`
	DefaultCount            int     `json:"default_count"`
	MaximumCount            int     `json:"maximum_count"`
	ServiceFeatures         string  `json:"service_features"`
	IsActive                bool    `json:"is_active"`
	IsDentalDirection       bool    `json:"is_dental_direction"`
	HasTooth                bool    `json:"has_tooth"`
	AllowMultipleUse        bool    `json:"allow_multiple_use"`
}

// Response پاسخ API خدمت.
type Response struct {
	ID                      uint    `json:"id"`
	ServiceCode             string  `json:"service_code"`
	Name                    string  `json:"name"`
	TechnicalCoefficient    float64 `json:"technical_coefficient"`
	ProfessionalCoefficient float64 `json:"professional_coefficient"`
	ConsumptionCoefficient  float64 `json:"consumption_coefficient"`
	ServiceRate             int `json:"service_rate"`
	ServiceTariff           int `json:"service_tariff"`
	InternationalCode       string  `json:"international_code"`
	DefaultCount            int     `json:"default_count"`
	MaximumCount            int     `json:"maximum_count"`
	ServiceFeatures         string  `json:"service_features"`
	IsActive                bool    `json:"is_active"`
	IsDentalDirection       bool    `json:"is_dental_direction"`
	HasTooth                bool    `json:"has_tooth"`
	AllowMultipleUse        bool    `json:"allow_multiple_use"`
}
