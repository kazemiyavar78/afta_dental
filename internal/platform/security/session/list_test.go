package session

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/listquery"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestR1_05_UserIDColumnIsOnlyForManager ستون user_id را فقط در spec مدیر می‌پذیرد.
func TestR1_05_UserIDColumnIsOnlyForManager(t *testing.T) {
	values := url.Values{"sort": {"user_id"}}
	if _, err := listquery.Parse(values, ListSpec(false)); err != apperror.ErrValidation {
		t.Fatalf("کاربر عادی نباید user_id را مرتب کند: %v", err)
	}
	if _, err := listquery.Parse(values, ListSpec(true)); err != nil {
		t.Fatalf("مدیر باید user_id را مرتب کند: %v", err)
	}
}

// TestR1_05_SearchSQLJoinsUsers فهرست نشست را با LEFT JOIN نام کاربر می‌سازد و برای غیرمدیر به خودش محدود می‌کند.
func TestR1_05_SearchSQLJoinsUsers(t *testing.T) {
	db, err := gorm.Open(sqlserver.Open("sqlserver://dryrun.invalid/none"), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
		Logger:               logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	q, err := listquery.Parse(url.Values{"filter[ip]": {"10.0.0"}}, ListSpec(false))
	if err != nil {
		t.Fatal(err)
	}
	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		var rows []sessionScan
		return q.Apply(sessionBase(context.Background(), tx, 7, false)).Select(sessionSelectSQL).Find(&rows)
	})
	if !strings.Contains(sql, "LEFT JOIN") || strings.Contains(sql, "INNER JOIN") {
		t.Fatalf("باید LEFT JOIN باشد: %s", sql)
	}
	if !strings.Contains(sql, "PersonnelAccountID") || !strings.Contains(sql, "Users.Username") {
		t.Fatalf("نام یا محدودیت مالک در SQL نیست: %s", sql)
	}
}
