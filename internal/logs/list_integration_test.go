//go:build integration

package logs

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/listquery"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
	"github.com/tpdenta/afta-reception/internal/user"
	"gorm.io/gorm"
)

// TestIntegration_R1_01_LogsFilterSortAndLegacyActor فیلتر، مرتب‌سازی و نام ردیف قدیمی را روی پایگاه واقعی چک می‌کند.
// D-16: فیلتر متنی IP ردیف 192.168.1.x را برمی‌گرداند و نشانی دیگر را نه.
func TestIntegration_R1_01_LogsFilterSortAndLegacyActor(t *testing.T) {
	database := testutil.OpenTestDB(t)
	if err := database.AutoMigrate(&user.Role{}, &user.User{}, &audit.SecurityEvent{}, &audit.AuditChainState{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	testutil.WithTestTx(t, func(tx *gorm.DB) {
		suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
		account := insertLogUser(t, tx, suffix)
		token := "r101-" + suffix
		legacy := insertLogEvent(t, tx, audit.SecurityEvent{
			UserID: account, HashVersion: audit.HashVersionV1, Ip: "192.168.1.10",
			EventType: audit.EventLoginFailed, Description: token + "-legacy",
			CreatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		})
		otherIP := insertLogEvent(t, tx, audit.SecurityEvent{
			ActorUsername: "sara", Ip: "10.0.0.8", EventType: audit.EventLogout,
			Description: token + "-other", CreatedAt: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC),
			HashVersion: audit.HashVersionV2,
		})
		kept := insertLogEvent(t, tx, audit.SecurityEvent{
			UserID: account, ActorUsername: "stored-name", Ip: "192.168.1.20",
			EventType: audit.EventAuditRead, Description: token + "-kept",
			CreatedAt: time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC), HashVersion: audit.HashVersionV2,
		})
		system := insertLogEvent(t, tx, audit.SecurityEvent{
			Ip: "192.168.1.30", EventType: audit.EventAuditStarted, Description: token + "-system",
			CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), HashVersion: audit.HashVersionV1,
		})

		svc := NewService(tx, audit.NewRepository(tx))
		ipPage := mustList(t, svc, "filter[description]="+token+"&filter[ip]=192.168.1&sort=id&order=asc", nil)
		ids := itemIDs(ipPage.Items)
		if !containsID(ids, legacy.ID) || !containsID(ids, kept.ID) || !containsID(ids, system.ID) || containsID(ids, otherIP.ID) {
			t.Fatalf("فیلتر IP متنی ردیف‌ها را درست جدا نکرد: %#v", ipPage.Items)
		}
		for _, item := range ipPage.Items {
			if !strings.Contains(item.IP, "192.168.1") {
				t.Fatalf("ردیف خارج از پیشوند IP برگشته: %+v", item)
			}
		}

		sorted := mustList(t, svc, "filter[description]="+token+"&sort=event_type&order=asc", nil)
		if len(sorted.Items) != 4 || sorted.Items[0].EventType != audit.EventAuditRead || sorted.Items[len(sorted.Items)-1].EventType != audit.EventLogout {
			t.Fatalf("مرتب‌سازی event_type نادرست است: %#v", sorted.Items)
		}

		byName := mustList(t, svc, "filter[actor_username]=u"+suffix, nil)
		nameIDs := itemIDs(byName.Items)
		if !containsID(nameIDs, legacy.ID) || containsID(nameIDs, kept.ID) || containsID(nameIDs, system.ID) {
			t.Fatalf("فیلتر نام باید ردیف قدیمی را از join بیاورد: %#v", byName.Items)
		}
		if actorOf(byName.Items, legacy.ID) != "u"+suffix {
			t.Fatalf("نام ردیف نسخه ۱ باید از Users باشد: %+v", byName.Items)
		}
		systemPage := mustList(t, svc, "filter[description]="+token+"-system", nil)
		if len(systemPage.Items) != 1 || systemPage.Items[0].ActorUsername != unknownActorLabel {
			t.Fatalf("ردیف بدون کاربر باید %s باشد: %+v", unknownActorLabel, systemPage.Items)
		}
		keptPage := mustList(t, svc, "filter[description]="+token+"-kept", nil)
		if len(keptPage.Items) != 1 || keptPage.Items[0].ActorUsername != "stored-name" || keptPage.Items[0].EventTypeLabel == "" {
			t.Fatalf("نام ذخیره‌شده باید مقدم بر join باشد: %+v", keptPage.Items)
		}
	})
}

func insertLogUser(t *testing.T, tx *gorm.DB, suffix string) *int {
	t.Helper()
	role := user.Role{Name: "role" + suffix, IntegrityHash: "x", CreatedAt: time.Now().UTC()}
	if err := tx.Create(&role).Error; err != nil {
		t.Fatalf("role: %v", err)
	}
	account := user.User{
		Username: "u" + suffix, PasswordHash: "x", Name: "تست", Family: "لاگ",
		RoleID: role.ID, IsActive: true, SecurityCode: "x", IntegrityHash: "x",
		PasswordChangedAt: time.Now().UTC(), UserType: user.UserTypeUser,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := tx.Create(&account).Error; err != nil {
		t.Fatalf("user: %v", err)
	}
	if account.ID == 0 {
		if err := tx.Where("Username = ?", account.Username).First(&account).Error; err != nil {
			t.Fatal(err)
		}
	}
	id := account.ID
	return &id
}

func insertLogEvent(t *testing.T, tx *gorm.DB, row audit.SecurityEvent) audit.SecurityEvent {
	t.Helper()
	if row.PrevHash == "" {
		row.PrevHash = "prev"
	}
	if row.RowHash == "" {
		row.RowHash = "row"
	}
	if row.HashVersion == 0 {
		row.HashVersion = audit.HashVersionV1
	}
	if err := tx.Create(&row).Error; err != nil {
		t.Fatalf("event: %v", err)
	}
	if row.ID == 0 {
		if err := tx.Where("Description = ?", row.Description).First(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	return row
}

func mustList(t *testing.T, svc *Service, raw string, only *int) listquery.Page[EventItem] {
	t.Helper()
	q, err := listquery.Parse(mapValues(raw), logSpec())
	if err != nil {
		t.Fatal(err)
	}
	page, err := svc.List(context.Background(), q, only)
	if err != nil {
		t.Fatal(err)
	}
	return page
}

func itemIDs(items []EventItem) []int64 {
	out := make([]int64, 0, len(items))
	for _, item := range items {
		out = append(out, item.ID)
	}
	return out
}

func containsID(ids []int64, want int64) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}

func actorOf(items []EventItem, id int64) string {
	for _, item := range items {
		if item.ID == id {
			return item.ActorUsername
		}
	}
	return ""
}
