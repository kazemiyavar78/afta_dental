package user

import (
	"context"
	"strings"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/encryption"
	"github.com/tpdenta/afta-reception/internal/platform/security/loginguard"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
	"golang.org/x/crypto/bcrypt"
)

func loginService(repo *stubRepo, gate *fakeGate, rec *testutil.FakeAuditRecorder) *Service {
	return &Service{
		repo: repo, signer: testSigner(), audit: rec, loginGuard: gate, session: fakeSessions{},
	}
}

func readyUser(t *testing.T, password string, locked, active bool) *User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	signer := testSigner()
	role := signedRole(signer, 1, "پذیرش", "", nil)
	user := &User{
		ID: 7, Username: "ali", PasswordHash: string(hash), Name: "علی", Family: "رضایی",
		RoleID: 1, Role: *role, IsActive: active, IsLocked: locked, UserType: UserTypeUser,
	}
	enc, err := encryption.NewEncryptor("12345678901234567890123456789012")
	if err != nil {
		t.Fatal(err)
	}
	code, err := encryption.NewUserEncryptionService(enc).CreateSecurityCode(user.ToSensitiveData())
	if err != nil {
		t.Fatal(err)
	}
	user.SecurityCode = code
	user.IntegrityHash = SignUserIntegrityHash(signer, user)
	return user
}

func loginActor() audit.Actor {
	return audit.Actor{IP: "203.0.113.10", UserAgent: "UnitBrowser"}
}

// TestR1_18_LoginFailedReasons هر دلیل ورود ناموفق را با نام کاربری واردشده ثبت می‌کند.
func TestR1_18_LoginFailedReasons(t *testing.T) {
	signerUser := readyUser(t, "Right@123", false, true)
	cases := []struct {
		name   string
		user   *User
		miss   bool
		pass   string
		work   error
		reason string
		count  bool
	}{
		{name: "نبود کاربر", miss: true, pass: "x", reason: audit.ReasonUserNotFound, count: true},
		{name: "یکپارچگی", user: &User{ID: 7, Username: "ali", PasswordHash: "x", IntegrityHash: "bad", Role: Role{ID: 1}}, pass: "x", reason: audit.ReasonIntegrityViolation},
		{name: "قفل حساب", user: readyUser(t, "Right@123", true, true), pass: "Right@123", reason: audit.ReasonAccountLocked},
		{name: "غیرفعال", user: readyUser(t, "Right@123", false, false), pass: "Right@123", reason: audit.ReasonAccountInactive},
		{name: "ساعت کاری", user: signerUser, pass: "Right@123", work: apperror.ErrWorkHours, reason: audit.ReasonOutsideWorkHours},
		{name: "رمز", user: signerUser, pass: "Wrong@123", reason: audit.ReasonWrongPassword, count: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &stubRepo{user: tc.user, missing: tc.miss}
			gate := &fakeGate{workErr: tc.work, minutes: 12}
			rec := &testutil.FakeAuditRecorder{}
			svc := loginService(repo, gate, rec)
			_, _, err := svc.HandleUserLogin(context.Background(), LoginRequest{Username: "ali", Password: tc.pass}, loginActor())
			if err == nil {
				t.Fatal("ورود باید شکست بخورد")
			}
			got := rec.Events[0].Event
			if got.Type != audit.EventLoginFailed || got.Reason != tc.reason || got.Target.Name != "ali" {
				t.Fatalf("رویداد: type=%s reason=%s target=%s", got.Type, got.Reason, got.Target.Name)
			}
			if tc.count && gate.failCalls != 1 {
				t.Fatalf("شمارنده IP باید یک بار زیاد شود، %d", gate.failCalls)
			}
			if !tc.count && gate.failCalls != 0 {
				t.Fatalf("این دلیل نباید شمارنده IP را زیاد کند: %d", gate.failCalls)
			}
		})
	}
}

// TestR1_18_IPLockedDetails رسیدن شمارنده به آستانه، مدت قفل را در IP_LOCKED می‌نویسد.
func TestR1_18_IPLockedDetails(t *testing.T) {
	repo := &stubRepo{missing: true}
	gate := &fakeGate{failLocks: true, minutes: 30}
	rec := &testutil.FakeAuditRecorder{}
	svc := loginService(repo, gate, rec)
	_, _, err := svc.HandleUserLogin(context.Background(), LoginRequest{Username: "ghost", Password: "x"}, loginActor())
	if err == nil {
		t.Fatal("باید شکست بخورد")
	}
	if len(rec.Events) != 2 || rec.Events[1].Event.Type != audit.EventIPLocked {
		t.Fatalf("باید LOGIN_FAILED و سپس IP_LOCKED باشد: %+v", rec.Events)
	}
	details := rec.Events[1].Event.Details.(map[string]any)
	if details["lock_minutes"] != 30 || rec.Events[0].Event.Reason != audit.ReasonUserNotFound {
		t.Fatalf("جزئیات قفل نادرست است: %#v", details)
	}
}

// TestR1_18_LoginBlockedOncePerMinute در زمان قفل IP فقط یک LOGIN_BLOCKED در دقیقه ثبت می‌شود.
func TestR1_18_LoginBlockedOncePerMinute(t *testing.T) {
	gate := &fakeGate{ipLocked: true, blocks: loginguard.NewBlockedLoginLog()}
	rec := &testutil.FakeAuditRecorder{}
	svc := loginService(&stubRepo{}, gate, rec)
	for i := 0; i < 2; i++ {
		_, _, err := svc.HandleUserLogin(context.Background(), LoginRequest{Username: "ali", Password: "x"}, loginActor())
		if err != apperror.ErrIPLocked {
			t.Fatalf("باید قفل IP باشد: %v", err)
		}
	}
	if len(rec.Events) != 1 || rec.Events[0].Event.Type != audit.EventLoginBlocked || rec.Events[0].Event.Reason != audit.ReasonIPLocked {
		t.Fatalf("باید دقیقا یک LOGIN_BLOCKED باشد: %+v", rec.Events)
	}
	if rec.Events[0].Event.Target.Name != "ali" {
		t.Fatalf("هدف باید نام واردشده باشد: %s", rec.Events[0].Event.Target.Name)
	}
}

// TestR1_18_LoginSuccessRecordsSession ورود موفق SESSION_CREATED را با IP و مرورگر می‌نویسد.
func TestR1_18_LoginSuccessRecordsSession(t *testing.T) {
	user := readyUser(t, "Right@123", false, true)
	enc, err := encryption.NewEncryptor("12345678901234567890123456789012")
	if err != nil {
		t.Fatal(err)
	}
	rec := &testutil.FakeAuditRecorder{}
	svc := loginService(&stubRepo{user: user}, &fakeGate{}, rec)
	svc.encryptSvc = encryption.NewUserEncryptionService(enc)
	logged, _, err := svc.HandleUserLogin(context.Background(), LoginRequest{Username: "ali", Password: "Right@123"}, loginActor())
	if err != nil {
		t.Fatalf("ورود: %v", err)
	}
	if logged.Username != "ali" || len(rec.Events) != 1 {
		t.Fatalf("باید فقط ایجاد نشست ثبت شود: %d", len(rec.Events))
	}
	got := rec.Events[0].Event
	details := got.Details.(map[string]any)
	if got.Type != audit.EventSessionCreated || got.Actor.Username != "ali" || details["ip"] != "203.0.113.10" || details["browser"] != "UnitBrowser" {
		t.Fatalf("ایجاد نشست نادرست است: %+v", got)
	}
	if strings.Contains(got.Description, "-") {
		t.Fatalf("توضیح نباید شناسه نشست داشته باشد: %s", got.Description)
	}
}
