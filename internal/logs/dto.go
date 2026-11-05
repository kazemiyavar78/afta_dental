package logs

import "time"

// LogFilterQuery پارامترهای query برای فیلتر لاگ (فقط ادمین).
type LogFilterQuery struct {
	IP       string     `form:"ip"`
	Username string     `form:"username"`
	FromDate *time.Time `form:"from_date" time_format:"2006-01-02"`
	ToDate   *time.Time `form:"to_date" time_format:"2006-01-02"`
}
