package logs

import (
	"strings"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/listquery"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestR1_01_ListSQLUsesLeftJoin نام فاعل را با LEFT JOIN و همان عبارت فیلتر می‌سازد.
func TestR1_01_ListSQLUsesLeftJoin(t *testing.T) {
	db, err := gorm.Open(sqlserver.Open("sqlserver://dryrun.invalid/none"), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
		Logger:               logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	q, err := listquery.Parse(mapValues("filter[actor_username]=ali&filter[ip]=192.168.1"), logSpec())
	if err != nil {
		t.Fatal(err)
	}
	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		var rows []eventRow
		return q.Apply(tx.Model(&audit.SecurityEvent{}).Joins(userJoin)).Select(eventSelectSQL).Find(&rows)
	})
	if !strings.Contains(sql, "LEFT JOIN") || strings.Contains(sql, "INNER JOIN") {
		t.Fatalf("باید LEFT JOIN باشد: %s", sql)
	}
	if !strings.Contains(sql, unknownActorLabel) || !strings.Contains(sql, "LIKE") {
		t.Fatalf("عبارت نام یا LIKE در SQL نیست: %s", sql)
	}
}

// TestR1_01_ToItemAddsPersianLabels برچسب نوع و دلیل را از کاتالوگ روی ردیف می‌گذارد.
func TestR1_01_ToItemAddsPersianLabels(t *testing.T) {
	item := toItem(eventRow{EventType: audit.EventLoginFailed, Reason: audit.ReasonWrongPassword, ActorUsername: unknownActorLabel})
	if item.EventTypeLabel != "ورود ناموفق" || item.ReasonLabel != "رمز نادرست" || item.ActorUsername != unknownActorLabel {
		t.Fatalf("برچسب ردیف نادرست است: %+v", item)
	}
}

func mapValues(raw string) map[string][]string {
	values := map[string][]string{}
	for _, part := range strings.Split(raw, "&") {
		key, val, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		values[key] = []string{val}
	}
	return values
}
