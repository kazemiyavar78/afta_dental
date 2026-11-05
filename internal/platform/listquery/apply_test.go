package listquery

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func testSpec() Spec {
	return Spec{
		DefaultSort: "id",
		DefaultDesc: true,
		Columns: map[string]Column{
			"id":         {Expr: "Items.ID", Type: KindNumber, Sortable: true, Filterable: true},
			"name":       {Expr: "Items.Name", Type: KindText, Sortable: true, Filterable: true},
			"code":       {Expr: "Items.Code", Type: KindEnum, Sortable: true, Filterable: true},
			"created_at": {Expr: "Items.CreatedAt", Type: KindDate, Sortable: true, Filterable: true},
			"ip":         {Expr: "Items.Ip", Type: KindText, Sortable: true, Filterable: true, Trim: true},
			"active":     {Expr: "Items.Active", Type: KindBool, Sortable: false, Filterable: true},
			"locked":     {Expr: "Items.Locked", Type: KindExact, Sortable: false, Filterable: false},
		},
	}
}

func drySQL(t *testing.T, q Query) (string, []any) {
	t.Helper()
	db, err := gorm.Open(sqlserver.Open("sqlserver://dryrun.invalid/none"), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
		Logger:               logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("dialector: %v", err)
	}
	var rows []struct {
		ID int `gorm:"column:ID"`
	}
	tx := q.Apply(db.Table("Items"))
	if err := tx.Find(&rows).Error; err != nil {
		t.Fatalf("dry run: %v", err)
	}
	return tx.Statement.SQL.String(), append([]any(nil), tx.Statement.Vars...)
}

func mustParse(t *testing.T, raw string) Query {
	t.Helper()
	values, err := url.ParseQuery(raw)
	if err != nil {
		t.Fatal(err)
	}
	return mustParseValues(t, values)
}

func mustParseValues(t *testing.T, values url.Values) Query {
	t.Helper()
	q, err := Parse(values, testSpec())
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return q
}

func reject(t *testing.T, raw string) {
	t.Helper()
	values, err := url.ParseQuery(raw)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Parse(values, testSpec())
	if err != apperror.ErrValidation {
		t.Fatalf("انتظار VALIDATION_ERROR برای %s، دریافت %v", raw, err)
	}
}

// TestListQuery_UnknownSortAndFilterRejected ستون خارج از تعریف را رد می‌کند.
func TestListQuery_UnknownSortAndFilterRejected(t *testing.T) {
	reject(t, "sort=drop_table")
	reject(t, "sort=locked")
	reject(t, "order=sideways")
	reject(t, "filter[password]=secret")
	reject(t, "filter[created_at]=2026-01-01")
	reject(t, "filter[name_from]=a")
	reject(t, "filter[id]=4")
	reject(t, "page=0")
	reject(t, "page_size=0")
	reject(t, "filter[code]=a,,b")
	reject(t, "filter[active]=yes")
	reject(t, "filter[created_at_from]=1404/08/15")
	reject(t, "filter[id_from]=1.5")

	q := mustParse(t, "sort=name&order=asc")
	sql, _ := drySQL(t, q)
	if strings.Contains(sql, "drop_table") || !strings.Contains(sql, "Items.Name") {
		t.Fatalf("مرتب‌سازی باید از عبارت سرور باشد: %s", sql)
	}
}

// TestListQuery_TextLikeEscapesWildcards درصد و زیرخط را در الگوی LIKE فرار می‌دهد.
func TestListQuery_TextLikeEscapesWildcards(t *testing.T) {
	q := mustParseValues(t, url.Values{"filter[name]": {"100%_"}})
	sql, vars := drySQL(t, q)
	if !strings.Contains(sql, "LIKE") || !strings.Contains(sql, "ESCAPE") {
		t.Fatalf("LIKE پارامتری نیست: %s", sql)
	}
	if strings.Contains(sql, "100%") {
		t.Fatalf("مقدار فیلتر نباید در متن SQL باشد: %s", sql)
	}
	if len(vars) == 0 || vars[0] != `%100\%\_%` {
		t.Fatalf("الگوی فرارشده نادرست است: %#v", vars)
	}
}

// TestListQuery_TrimBeforeTextFilter فاصله کنار IP را قبل از LIKE برمی‌دارد.
func TestListQuery_TrimBeforeTextFilter(t *testing.T) {
	q := mustParse(t, "filter[ip]= 192.168.1 ")
	_, vars := drySQL(t, q)
	if len(vars) == 0 || vars[0] != "%192.168.1%" {
		t.Fatalf("IP باید بدون فاصله در الگو باشد: %#v", vars)
	}
}

// TestListQuery_TextLikeEscapesBracket کروشه باز را در LIKE به کلاس حرفی تبدیل می‌کند.
// در SQL Server نویسه [ با ESCAPE بی‌اثر نمی‌شود و شکل لفظی‌اش [[] است.
func TestListQuery_TextLikeEscapesBracket(t *testing.T) {
	q := mustParseValues(t, url.Values{"filter[name]": {"a[b"}})
	_, vars := drySQL(t, q)
	if len(vars) == 0 || vars[0] != `%a[[]b%` {
		t.Fatalf("کروشه باید لفظی شود: %#v", vars)
	}
}

// TestListQuery_EnumRejectsMoreThan50Values بیش از ۵۰ مقدار انتخابی را رد می‌کند.
// سقف جلوی رسیدن یک فیلتر به حد ۲۱۰۰ پارامتر SQL Server را می‌گیرد.
func TestListQuery_EnumRejectsMoreThan50Values(t *testing.T) {
	reject(t, "filter[code]="+enumValues(maxEnumFilterValues+1))
	q := mustParse(t, "filter[code]="+enumValues(maxEnumFilterValues))
	_, vars := drySQL(t, q)
	if len(vars) != maxEnumFilterValues {
		t.Fatalf("پنجاه مقدار باید پذیرفته شود: %d", len(vars))
	}
}

// enumValues به تعداد n مقدار جداشده با ویرگول می‌سازد.
// n تعداد مقدار است. خروجی برای فیلتر enum تست است.
func enumValues(n int) string {
	parts := make([]string, n)
	for i := range parts {
		parts[i] = fmt.Sprintf("v%d", i)
	}
	return strings.Join(parts, ",")
}

// TestListQuery_EnumInMultipleValues چند مقدار enum را به IN پارامتری تبدیل می‌کند.
func TestListQuery_EnumInMultipleValues(t *testing.T) {
	q := mustParse(t, "filter[code]=success,failure")
	sql, vars := drySQL(t, q)
	if !strings.Contains(sql, "IN (") || strings.Contains(sql, "success") {
		t.Fatalf("IN باید پارامتری باشد: %s", sql)
	}
	found := map[string]bool{}
	for _, v := range vars {
		if s, ok := v.(string); ok {
			found[s] = true
		}
	}
	if !found["success"] || !found["failure"] {
		t.Fatalf("هر دو مقدار enum باید پارامتر باشند: %#v", vars)
	}
}

// TestListQuery_DateRangeThroughEndOfDay پایان بازه را تا شروع روز بعد می‌برد.
func TestListQuery_DateRangeThroughEndOfDay(t *testing.T) {
	q := mustParse(t, "filter[created_at_from]=2026-01-15&filter[created_at_to]=2026-01-15")
	sql, vars := drySQL(t, q)
	if !strings.Contains(sql, ">=") || !strings.Contains(sql, "<") {
		t.Fatalf("بازه تاریخ در SQL نیست: %s", sql)
	}
	start := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC)
	if !hasTime(vars, start) || !hasTime(vars, end) {
		t.Fatalf("مرزهای روز نادرست است: %#v", vars)
	}
}

// TestListQuery_NumberRange هر دو سر بازه عددی را پارامتر می‌کند.
func TestListQuery_NumberRange(t *testing.T) {
	q := mustParse(t, "filter[id_from]=2&filter[id_to]=9")
	sql, vars := drySQL(t, q)
	if !strings.Contains(sql, ">=") || !strings.Contains(sql, "<=") {
		t.Fatalf("بازه عدد در SQL نیست: %s", sql)
	}
	if !hasInt(vars, 2) || !hasInt(vars, 9) {
		t.Fatalf("مرزهای عدد نادرست است: %#v", vars)
	}
}

// TestListQuery_PageSizeCappedAt100 اندازه صفحه بزرگ‌تر از ۱۰۰ را در SQL به ۱۰۰ محدود می‌کند.
func TestListQuery_PageSizeCappedAt100(t *testing.T) {
	q := mustParse(t, "page=3&page_size=500")
	if q.PageSize != MaxPageSize {
		t.Fatalf("سقف page_size اعمال نشد: %d", q.PageSize)
	}
	sql, _ := drySQL(t, q)
	if strings.Contains(sql, "500") || !strings.Contains(sql, "OFFSET 200") || !strings.Contains(sql, "FETCH NEXT 100") {
		t.Fatalf("OFFSET/FETCH سقف را نشان نمی‌دهد: %s", sql)
	}
}

// TestListQuery_BoolExact فقط true و false را برابری می‌کند.
func TestListQuery_BoolExact(t *testing.T) {
	q := mustParse(t, "filter[active]=TRUE")
	sql, vars := drySQL(t, q)
	if !strings.Contains(sql, "Items.Active") || !hasBool(vars, true) {
		t.Fatalf("بول درست نیست: %s %#v", sql, vars)
	}
}

// TestNewPage_NilItemsBecomesEmptyArray فهرست تهی را به آرایه خالی تبدیل می‌کند.
func TestNewPage_NilItemsBecomesEmptyArray(t *testing.T) {
	page := NewPage[int](nil, 0, Query{Page: 1, PageSize: 20})
	if page.Items == nil || page.Total != 0 || page.Page != 1 || page.PageSize != 20 {
		t.Fatalf("پاسخ استاندارد نادرست است: %+v", page)
	}
}

func hasTime(vars []any, want time.Time) bool {
	for _, v := range vars {
		if got, ok := v.(time.Time); ok && got.Equal(want) {
			return true
		}
	}
	return false
}

func hasInt(vars []any, want int) bool {
	for _, v := range vars {
		switch n := v.(type) {
		case int:
			if n == want {
				return true
			}
		case int64:
			if n == int64(want) {
				return true
			}
		}
	}
	return false
}

func hasBool(vars []any, want bool) bool {
	for _, v := range vars {
		if got, ok := v.(bool); ok && got == want {
			return true
		}
	}
	return false
}
