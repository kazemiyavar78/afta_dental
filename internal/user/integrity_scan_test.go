package user

import (
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
)

// TestD23_InvalidUsers_ReturnsEveryBrokenRow هر دو کاربر خراب را برمی‌گرداند و رویداد نمی‌سازد.
// ردیف سالم وسط فهرست نباید حلقه را قطع کند. D-23: ثبت هشدار کار این تابع نیست.
func TestD23_InvalidUsers_ReturnsEveryBrokenRow(t *testing.T) {
	signer := testSigner()
	good := User{ID: 3, Username: "ok"}
	good.IntegrityHash = SignUserIntegrityHash(signer, &good)
	repo := &stubRepo{users: []User{
		{ID: 1, Username: "ali"},
		good,
		{ID: 2, Username: "reza"},
	}}
	rec := &testutil.FakeAuditRecorder{}
	got, err := (&Service{repo: repo, signer: signer, audit: rec}).InvalidUsers()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].EntityID != "1" || got[0].EntityName != "ali" || got[1].EntityName != "reza" {
		t.Fatalf("هر دو کاربر خراب باید برگردند: %+v", got)
	}
	if got[0].EntityType != audit.TargetUser || len(rec.Events) != 0 {
		t.Fatalf("سرویس نباید رویداد بسازد: %+v events=%d", got, len(rec.Events))
	}
}

// TestD23_InvalidRoles_ReturnsEveryBrokenRow هر دو نقش خراب را برمی‌گرداند.
// نقش سالم بین آن‌ها حلقه را قطع نمی‌کند و رویدادی ثبت نمی‌شود.
func TestD23_InvalidRoles_ReturnsEveryBrokenRow(t *testing.T) {
	signer := testSigner()
	good := Role{ID: 2, Name: "ok"}
	good.IntegrityHash = SignRoleIntegrityHash(signer, &good, nil)
	repo := &stubRepo{roles: []Role{
		{ID: 1, Name: "broken-a"},
		good,
		{ID: 3, Name: "broken-b"},
	}}
	rec := &testutil.FakeAuditRecorder{}
	got, err := (&Service{repo: repo, signer: signer, audit: rec}).InvalidRoles()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].EntityName != "broken-a" || got[1].EntityID != "3" || got[0].EntityType != audit.TargetRole {
		t.Fatalf("هر دو نقش خراب باید برگردند: %+v", got)
	}
	if len(rec.Events) != 0 {
		t.Fatal("سرویس نباید رویداد بسازد")
	}
}
