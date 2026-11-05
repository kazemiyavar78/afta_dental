// Package listquery صفحه‌بندی، فیلتر و مرتب‌سازی سمت سرور است.
// هر API فقط ستون‌هایی را می‌پذیرد که خودش با عبارت SQL ثابت تعریف کرده است.
// نام ستون و مقدار از درخواست می‌آیند؛ عبارت SQL هرگز از ورودی کاربر ساخته نمی‌شود.
package listquery

// Kind نوع مقایسه یک ستون فیلتر است.
type Kind string

const (
	// KindText جستجوی LIKE با پیش و پس درصد است و % و _ در مقدار escape می‌شوند.
	KindText Kind = "text"
	// KindExact برابری دقیق یک مقدار است.
	KindExact Kind = "exact"
	// KindEnum عضویت در فهرست جداشده با ویرگول است.
	KindEnum Kind = "enum"
	// KindNumber بازه عددی با filter[col_from] و filter[col_to] است.
	KindNumber Kind = "number"
	// KindDate بازه تاریخ میلادی YYYY-MM-DD است. پایان بازه تا آخر همان روز است.
	KindDate Kind = "date"
	// KindBool برابری true یا false است.
	KindBool Kind = "bool"

	// DefaultPage صفحه اول وقتی پارامتر page نیامده باشد.
	DefaultPage = 1
	// DefaultPageSize اندازه صفحه وقتی page_size نیامده باشد.
	DefaultPageSize = 20
	// MaxPageSize سقف اندازه صفحه است. مقدار بزرگ‌تر به همین سقف بریده می‌شود.
	MaxPageSize = 100
)

// Column یک ستون مجاز در قرارداد یک API است.
// Expr عبارت SQL ثابت سمت سرور است. Type شکل فیلتر را تعیین می‌کند.
// Sortable و Filterable جدا هستند تا ستونی فقط یکی از این دو را داشته باشد.
// Trim فاصله دو طرف مقدار را قبل از فیلتر برمی‌دارد؛ برای نشانی IP استفاده می‌شود.
type Column struct {
	Expr       string
	Type       Kind
	Sortable   bool
	Filterable bool
	Trim       bool
}

// Spec ستون‌های یک فهرست و مرتب‌سازی پیش‌فرض آن است.
// DefaultSort باید نام یک ستون Sortable باشد. DefaultDesc جهت پیش‌فرض است وقتی order نیامده.
type Spec struct {
	Columns     map[string]Column
	DefaultSort string
	DefaultDesc bool
}

// filterClause یک شرط پارامتری است که Apply به کوئری اضافه می‌کند.
// sql فقط از Expr تعریف سرور و عملگر ثابت ساخته می‌شود. args مقادیر کاربر هستند.
type filterClause struct {
	sql  string
	args []any
}

// Query پارامترهای اعتبارسنجی‌شده یک فهرست است.
// Page و PageSize بعد از Parse در بازه مجازند. بقیه فیلدها فقط داخل همین پکیج پر می‌شوند.
type Query struct {
	Page     int
	PageSize int
	sortExpr string
	sortDesc bool
	filters  []filterClause
	applied  map[string]string
}

// FilterSummary کپی فیلترهای پذیرفته‌شده را برمی‌گرداند.
// کلید نام ستون یا پسوند _from و _to است. مقدار، متن نرمال‌شده همان فیلتر است.
// برای ممیزی خواندن لاگ استفاده می‌شود و بدنه ردیف‌ها را ندارد.
func (q Query) FilterSummary() map[string]string {
	out := make(map[string]string, len(q.applied))
	for key, value := range q.applied {
		out[key] = value
	}
	return out
}
