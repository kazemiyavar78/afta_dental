package session

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// MSSQLUUID نوع UUID سازگار با uniqueidentifier در SQL Server است.
// SQL Server سه بخش اول GUID را little-endian ذخیره می‌کند (mixed-endian)،
// در حالی که RFC 4122 و github.com/google/uuid از big-endian استفاده می‌کنند.
type MSSQLUUID uuid.UUID

// NewMSSQLUUID یک UUID تصادفی نسخه ۴ می‌سازد.
func NewMSSQLUUID() MSSQLUUID {
	return MSSQLUUID(uuid.New())
}

// ParseMSSQLUUID رشته UUID استاندارد را به MSSQLUUID تبدیل می‌کند.
func ParseMSSQLUUID(s string) (MSSQLUUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return MSSQLUUID{}, err
	}
	return MSSQLUUID(id), nil
}

// UUID مقدار را به uuid.UUID استاندارد تبدیل می‌کند.
func (id MSSQLUUID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

// String نمایش استاندارد RFC 4122 را برمی‌گرداند.
func (id MSSQLUUID) String() string {
	return uuid.UUID(id).String()
}

// GormDataType نوع ستون را برای GORM مشخص می‌کند.
func (MSSQLUUID) GormDataType() string {
	return "uniqueidentifier"
}

// Scan مقدار خوانده‌شده از دیتابیس را از فرمت مایکروسافت به RFC 4122 تبدیل می‌کند.
func (id *MSSQLUUID) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*id = MSSQLUUID(uuid.Nil)
		return nil
	case []byte:
		if len(v) == 16 {
			var raw [16]byte
			copy(raw[:], v)
			transposeUUIDEndianness(&raw)
			*id = MSSQLUUID(raw)
			return nil
		}
		// برخی درایورها GUID را به‌صورت رشته بایتی برمی‌گردانند.
		parsed, err := uuid.ParseBytes(v)
		if err != nil {
			return fmt.Errorf("mssql uuid: اسکن []byte نامعتبر: %w", err)
		}
		*id = MSSQLUUID(parsed)
		return nil
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return fmt.Errorf("mssql uuid: اسکن string نامعتبر: %w", err)
		}
		*id = MSSQLUUID(parsed)
		return nil
	case [16]byte:
		raw := v
		transposeUUIDEndianness(&raw)
		*id = MSSQLUUID(raw)
		return nil
	default:
		return fmt.Errorf("mssql uuid: نوع پشتیبانی‌نشده %T", src)
	}
}

// Value مقدار را برای ذخیره در SQL Server به فرمت mixed-endian تبدیل می‌کند.
func (id MSSQLUUID) Value() (driver.Value, error) {
	raw := [16]byte(id)
	transposeUUIDEndianness(&raw)
	out := make([]byte, 16)
	copy(out, raw[:])
	return out, nil
}

// MarshalJSON برای سازگاری با API/لاگ، UUID را به‌صورت رشته استاندارد سریال می‌کند.
func (id MSSQLUUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// UnmarshalJSON رشته JSON را به MSSQLUUID تبدیل می‌کند.
func (id *MSSQLUUID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseMSSQLUUID(s)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

// transposeUUIDEndianness سه بخش اول GUID را جابه‌جا می‌کند (Data1/Data2/Data3).
// این تبدیل برگشت‌پذیر است؛ همان تابع برای Scan و Value استفاده می‌شود.
func transposeUUIDEndianness(b *[16]byte) {
	b[0], b[1], b[2], b[3] = b[3], b[2], b[1], b[0]
	b[4], b[5] = b[5], b[4]
	b[6], b[7] = b[7], b[6]
}
