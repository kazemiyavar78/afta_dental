//go:build integration

package user

import (
	"context"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/listquery"
	"github.com/tpdenta/afta-reception/internal/platform/security/session"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
	"gorm.io/gorm"
)

// TestIntegration_R1_05_SessionFilterAndSort فیلتر IP و مرتب‌سازی مرورگر را روی نشست‌های واقعی چک می‌کند.
// کاربر عادی نشست کاربر دیگر را نمی‌بیند، حتی اگر پیشوند IP یکی باشد.
func TestIntegration_R1_05_SessionFilterAndSort(t *testing.T) {
	database := testutil.OpenTestDB(t)
	if err := database.AutoMigrate(&Role{}, &User{}, &session.Session{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	testutil.WithTestTx(t, func(tx *gorm.DB) {
		suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
		owner := insertSessionUser(t, tx, "own"+suffix)
		other := insertSessionUser(t, tx, "oth"+suffix)
		alpha := insertSessionRow(t, tx, owner, "192.168.1.10", "AAA-"+suffix, time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))
		late := insertSessionRow(t, tx, owner, "10.1.1.1", "MMM-"+suffix, time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC))
		insertSessionRow(t, tx, other, "192.168.1.40", "ZZZ-"+suffix, time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC))

		svc := session.NewService(tx, nil, nil)
		filtered := mustSearch(t, svc, "filter[ip]=192.168.1", owner, false)
		if len(filtered.Items) != 1 || filtered.Items[0].ID != alpha.String() || filtered.Items[0].Username != "own"+suffix {
			t.Fatalf("فیلتر IP یا نام کاربر نادرست است: %+v", filtered.Items)
		}

		sorted := mustSearch(t, svc, "sort=browser&order=asc", owner, false)
		if len(sorted.Items) != 2 || sorted.Items[0].ID != alpha.String() || sorted.Items[1].ID != late.String() {
			t.Fatalf("مرتب‌سازی مرورگر نادرست است: %+v", sorted.Items)
		}

		byUser := mustSearch(t, svc, "filter[user_id_from]="+strconv.Itoa(owner)+"&filter[user_id_to]="+strconv.Itoa(owner)+"&sort=ip&order=asc", owner, true)
		if len(byUser.Items) != 2 {
			t.Fatalf("فیلتر user_id مدیر باید فقط نشست‌های همان کاربر را بدهد: %+v", byUser.Items)
		}
	})
}

func insertSessionUser(t *testing.T, tx *gorm.DB, name string) int {
	t.Helper()
	role := Role{Name: "role-" + name, IntegrityHash: "x", CreatedAt: time.Now().UTC()}
	if err := tx.Create(&role).Error; err != nil {
		t.Fatalf("role: %v", err)
	}
	account := User{
		Username: name, PasswordHash: "x", Name: "تست", Family: "نشست",
		RoleID: role.ID, IsActive: true, SecurityCode: "x", IntegrityHash: "x",
		PasswordChangedAt: time.Now().UTC(), UserType: UserTypeUser,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := tx.Create(&account).Error; err != nil {
		t.Fatalf("user: %v", err)
	}
	if account.ID == 0 {
		if err := tx.Where("Username = ?", name).First(&account).Error; err != nil {
			t.Fatal(err)
		}
	}
	return account.ID
}

func insertSessionRow(t *testing.T, tx *gorm.DB, userID int, ip, browser string, at time.Time) session.MSSQLUUID {
	t.Helper()
	row := session.Session{
		ID: session.NewMSSQLUUID(), Ip: ip, Browser: browser, CreationTime: at, PersonnelAccountID: userID,
	}
	if err := tx.Create(&row).Error; err != nil {
		t.Fatalf("session: %v", err)
	}
	return row.ID
}

func mustSearch(t *testing.T, svc *session.Service, raw string, actorID int, canListAll bool) listquery.Page[session.ListedSession] {
	t.Helper()
	values, err := url.ParseQuery(raw)
	if err != nil {
		t.Fatal(err)
	}
	q, err := listquery.Parse(values, session.ListSpec(canListAll))
	if err != nil {
		t.Fatal(err)
	}
	page, err := svc.Search(context.Background(), q, actorID, canListAll)
	if err != nil {
		t.Fatal(err)
	}
	return page
}
