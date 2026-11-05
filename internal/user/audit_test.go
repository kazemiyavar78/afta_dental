package user

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/encryption"
	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
	"github.com/tpdenta/afta-reception/internal/platform/security/loginguard"
	"github.com/tpdenta/afta-reception/internal/platform/security/session"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
	"gorm.io/gorm"
)

type allowPassword struct{}

func (allowPassword) IsValidPassword(string) (bool, []string) { return true, nil }

type fakeGate struct {
	ipLocked  bool
	failLocks bool
	minutes   int
	workErr   error
	failCalls int
	blocks    *loginguard.BlockedLoginLog
}

func (g *fakeGate) CheckIPLock(string) error {
	if g.ipLocked {
		return apperror.ErrIPLocked
	}
	return nil
}

func (g *fakeGate) RecordFailedLogin(string) (bool, int, error) {
	g.failCalls++
	return g.failLocks, g.minutes, nil
}

func (g *fakeGate) ResetLoginAttempts(string) error { return nil }

func (g *fakeGate) CheckWorkHours(bool) error { return g.workErr }

func (g *fakeGate) AllowBlockedLoginLog(ip string) bool {
	if g.blocks == nil {
		g.blocks = loginguard.NewBlockedLoginLog()
	}
	return g.blocks.Allow(ip)
}

type fakeSessions struct{}

func (fakeSessions) CreateSession(context.Context, int, string, string) (*session.Session, error) {
	return &session.Session{ID: session.NewMSSQLUUID(), Ip: "203.0.113.10", Browser: "Unit"}, nil
}

func (fakeSessions) GetSessionsByUserID(int) ([]session.Session, error) { return nil, nil }

type stubRepo struct {
	user      *User
	role      *Role
	rolePerms []Permission
	byID      map[int]Permission
	missing   bool
}

func (r *stubRepo) Create(user *User) error {
	user.ID = 8
	cp := *user
	r.user = &cp
	return nil
}
func (r *stubRepo) FindByID(id int) (*User, error) {
	if r.user == nil || (id != 0 && r.user.ID != id) {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *r.user
	return &cp, nil
}
func (r *stubRepo) FindUsernameByID(int) (string, error) {
	if r.user == nil {
		return "", gorm.ErrRecordNotFound
	}
	return r.user.Username, nil
}
func (r *stubRepo) FindByUsername(name string) (*User, error) {
	if r.missing || r.user == nil || r.user.Username != name {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *r.user
	return &cp, nil
}
func (r *stubRepo) Update(user *User) error {
	cp := *user
	r.user = &cp
	return nil
}
func (r *stubRepo) Delete(int) error                      { return nil }
func (r *stubRepo) FindAll() ([]User, error)              { return nil, nil }
func (r *stubRepo) CountUsersByRoleID(int) (int64, error) { return 0, nil }
func (r *stubRepo) FindDoctors() ([]User, error)          { return nil, nil }
func (r *stubRepo) FindAssistants() ([]User, error)       { return nil, nil }
func (r *stubRepo) CreateRole(role *Role) error {
	role.ID = 3
	cp := *role
	r.role = &cp
	return nil
}
func (r *stubRepo) FindRoleByID(id int) (*Role, error) {
	if r.role == nil || r.role.ID != id {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *r.role
	return &cp, nil
}
func (r *stubRepo) FindRoleByName(name string) (*Role, error) {
	if r.role != nil && r.role.Name == name {
		cp := *r.role
		return &cp, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *stubRepo) FindAllRoles() ([]Role, error) { return nil, nil }
func (r *stubRepo) UpdateRole(role *Role) error {
	cp := *role
	r.role = &cp
	return nil
}
func (r *stubRepo) DeleteRole(int) error                      { return nil }
func (r *stubRepo) FindAllPermissions() ([]Permission, error) { return r.rolePerms, nil }
func (r *stubRepo) CreatePermission(*Permission) error        { return nil }
func (r *stubRepo) UpdatePermission(*Permission) error        { return nil }
func (r *stubRepo) FindPermissionByName(string) (*Permission, error) {
	return nil, gorm.ErrRecordNotFound
}
func (r *stubRepo) FindPermissionsByIDs(ids []int) ([]Permission, error) {
	out := make([]Permission, 0, len(ids))
	for _, id := range ids {
		p, ok := r.byID[id]
		if !ok {
			return nil, nil
		}
		out = append(out, p)
	}
	if len(out) != len(ids) {
		return nil, nil
	}
	return out, nil
}
func (r *stubRepo) ReplaceRolePermissions(int, []RolePermission) error { return nil }
func (r *stubRepo) FindRolePermissionByRoleID(int) ([]RolePermission, error) {
	return nil, nil
}
func (r *stubRepo) FindPermissionIDsByRoleID(int) ([]int, error) {
	ids := make([]int, len(r.rolePerms))
	for i, p := range r.rolePerms {
		ids[i] = p.ID
	}
	return ids, nil
}
func (r *stubRepo) FindPermissionsByRoleID(int) ([]Permission, error) {
	return append([]Permission(nil), r.rolePerms...), nil
}
func (r *stubRepo) DeleteRolePermissionsByRoleID(int) error { return nil }

func testActor(id int, name string) audit.Actor {
	return audit.Actor{UserID: &id, Username: name, IP: "203.0.113.10", UserAgent: "UnitBrowser"}
}

func testSigner() *integrity.Signer {
	return integrity.NewSigner("test-integrity-key-32bytes-long!!")
}

func signedRole(signer *integrity.Signer, id int, name, description string, perms []Permission) *Role {
	role := &Role{ID: id, Name: name, Description: description}
	ids := make([]int, len(perms))
	for i, p := range perms {
		ids[i] = p.ID
	}
	role.IntegrityHash = SignRoleIntegrityHash(signer, role, ids)
	return role
}

// TestR1_02_RoleCreatedRecordsActorAndTarget نام سازنده و نام نقش را در ROLE_CREATED می‌نویسد.
func TestR1_02_RoleCreatedRecordsActorAndTarget(t *testing.T) {
	repo := &stubRepo{byID: map[int]Permission{
		1: {ID: 1, Name: "users.read"},
	}}
	rec := &testutil.FakeAuditRecorder{}
	svc := &Service{repo: repo, signer: testSigner(), audit: rec}
	resp, err := svc.CreateRole(context.Background(), CreateRoleRequest{
		Name: "پذیرش", Description: "میز", PermissionIDs: []int{1},
	}, testActor(4, "narges"))
	if err != nil {
		t.Fatalf("ایجاد نقش: %v", err)
	}
	if resp.Name != "پذیرش" || len(rec.Events) != 1 {
		t.Fatalf("پاسخ یا تعداد رویداد نادرست است: %+v %d", resp, len(rec.Events))
	}
	got := rec.Events[0].Event
	if got.Type != audit.EventRoleCreated || got.Actor.Username != "narges" || got.Target.Name != "پذیرش" || got.Target.Type != audit.TargetRole {
		t.Fatalf("رویداد ایجاد نقش نادرست است: %+v", got)
	}
}

// TestR1_02_RoleCreateAuditFailureStops شکست ممیزی را به کاربر برمی‌گرداند.
func TestR1_02_RoleCreateAuditFailureStops(t *testing.T) {
	repo := &stubRepo{byID: map[int]Permission{1: {ID: 1, Name: "users.read"}}}
	rec := &testutil.FakeAuditRecorder{Err: errors.New("db down")}
	svc := &Service{repo: repo, signer: testSigner(), audit: rec}
	_, err := svc.CreateRole(context.Background(), CreateRoleRequest{Name: "نقش", PermissionIDs: []int{1}}, testActor(1, "admin"))
	app, ok := err.(*apperror.AppError)
	if !ok || app.Code != apperror.ErrAuditUnavailable.Code {
		t.Fatalf("باید AUDIT_UNAVAILABLE باشد: %v", err)
	}
}

// TestR1_04_RoleUpdatedListsPermissionDiff فقط مجوزهای عوض‌شده را با جمله فارسی می‌نویسد.
func TestR1_04_RoleUpdatedListsPermissionDiff(t *testing.T) {
	signer := testSigner()
	oldPerms := []Permission{{ID: 1, Name: "users.read"}, {ID: 2, Name: "patients.read"}}
	role := signedRole(signer, 2, "پذیرش", "همان", oldPerms)
	repo := &stubRepo{
		role:      role,
		rolePerms: oldPerms,
		byID: map[int]Permission{
			2: {ID: 2, Name: "patients.read"},
			3: {ID: 3, Name: "reception.read"},
		},
	}
	rec := &testutil.FakeAuditRecorder{}
	svc := &Service{repo: repo, signer: signer, audit: rec}
	if _, err := svc.UpdateRole(context.Background(), 2, UpdateRoleRequest{
		Name: "پذیرش", Description: "همان", PermissionIDs: []int{2, 3},
	}, testActor(4, "narges")); err != nil {
		t.Fatalf("به‌روزرسانی نقش: %v", err)
	}
	got := rec.Events[0].Event
	if got.Type != audit.EventRoleUpdated || got.Actor.Username != "narges" || got.Target.Name != "پذیرش" {
		t.Fatalf("رویداد نقش نادرست است: %+v", got)
	}
	want := "مجوزهای اضافه‌شده: reception.read؛ مجوزهای حذف‌شده: users.read"
	if got.Description != want {
		t.Fatalf("توضیح: %s", got.Description)
	}
	details := got.Details.(map[string]any)
	if _, ok := details["name"]; ok {
		t.Fatal("نام عوض نشده نباید در جزئیات باشد")
	}
	added := details["added_permissions"].([]string)
	removed := details["removed_permissions"].([]string)
	if len(added) != 1 || added[0] != "reception.read" || len(removed) != 1 || removed[0] != "users.read" {
		t.Fatalf("اختلاف مجوز نادرست است: %#v", details)
	}
}

// TestR1_04_RoleUpdatedOmitsUnchangedFields تغییر نام را بدون کلید مجوز می‌نویسد.
func TestR1_04_RoleUpdatedOmitsUnchangedFields(t *testing.T) {
	signer := testSigner()
	perms := []Permission{{ID: 1, Name: "users.read"}}
	role := signedRole(signer, 2, "قدیم", "شرح", perms)
	repo := &stubRepo{role: role, rolePerms: perms, byID: map[int]Permission{1: perms[0]}}
	rec := &testutil.FakeAuditRecorder{}
	svc := &Service{repo: repo, signer: signer, audit: rec}
	if _, err := svc.UpdateRole(context.Background(), 2, UpdateRoleRequest{
		Name: "جدید", Description: "شرح", PermissionIDs: []int{1},
	}, testActor(4, "narges")); err != nil {
		t.Fatalf("به‌روزرسانی: %v", err)
	}
	details := rec.Events[0].Event.Details.(map[string]any)
	name := details["name"].(fieldChange)
	if name.Before != "قدیم" || name.After != "جدید" {
		t.Fatalf("تغییر نام نادرست است: %#v", name)
	}
	if _, ok := details["added_permissions"]; ok {
		t.Fatal("مجوز عوض نشده نباید در جزئیات باشد")
	}
	if _, ok := details["description"]; ok {
		t.Fatal("توضیحات عوض نشده نباید در جزئیات باشد")
	}
}

// TestR1_02_RoleDeletedRecordsTargetName حذف نقش نام نقش را هدف می‌گذارد.
func TestR1_02_RoleDeletedRecordsTargetName(t *testing.T) {
	signer := testSigner()
	perms := []Permission{{ID: 1, Name: "users.read"}}
	role := signedRole(signer, 5, "موقت", "", perms)
	repo := &stubRepo{role: role, rolePerms: perms}
	rec := &testutil.FakeAuditRecorder{}
	svc := &Service{repo: repo, signer: signer, audit: rec}
	if err := svc.DeleteRole(context.Background(), 5, testActor(4, "narges")); err != nil {
		t.Fatalf("حذف: %v", err)
	}
	got := rec.Events[0].Event
	if got.Type != audit.EventRoleDeleted || got.Target.Name != "موقت" || got.Actor.Username != "narges" {
		t.Fatalf("حذف نقش نادرست است: %+v", got)
	}
}

// TestR1_21_UserCreatedRecordsTarget ایجاد کاربر هدف را نام کاربری می‌گذارد و رمز را در رویداد نمی‌آورد.
func TestR1_21_UserCreatedRecordsTarget(t *testing.T) {
	enc, err := encryption.NewEncryptor("12345678901234567890123456789012")
	if err != nil {
		t.Fatal(err)
	}
	signer := testSigner()
	role := signedRole(signer, 1, "پذیرش", "", nil)
	repo := &stubRepo{role: role}
	rec := &testutil.FakeAuditRecorder{}
	svc := &Service{
		repo: repo, signer: signer, audit: rec,
		encryptSvc:   encryption.NewUserEncryptionService(enc),
		pwdValidator: allowPassword{},
	}
	resp, err := svc.CreateUser(context.Background(), CreateUserRequest{
		Username: "ali", Password: "Secret@123", Name: "علی", Family: "رضایی", RoleID: 1,
	}, testActor(4, "narges"))
	if err != nil {
		t.Fatalf("ایجاد کاربر: %v", err)
	}
	got := rec.Events[0].Event
	if got.Type != audit.EventUserCreated || got.Target.Name != "ali" || got.Actor.Username != "narges" || resp.Username != "ali" {
		t.Fatalf("ایجاد کاربر نادرست است: %+v", got)
	}
	raw, _ := json.Marshal(got)
	if strings.Contains(string(raw), "Secret@123") {
		t.Fatal("رمز نباید در رویداد ایجاد کاربر باشد")
	}
}

// TestR1_21_UserUpdatedDiffOmitsPasswordAndUnchanged اختلاف فیلدها رمز و فیلد ثابت را ندارد.
func TestR1_21_UserUpdatedDiffOmitsPasswordAndUnchanged(t *testing.T) {
	secret := "super-secret-hash"
	code := "security-code-secret"
	before := User{
		ID: 1, Username: "ali", Name: "علی", Family: "رضایی",
		PhoneNumber: "09121234567", Address: "تهران", PasswordHash: secret, SecurityCode: code,
		RoleID: 1, UserType: UserTypeUser, IsActive: true,
	}
	same := before
	same.PasswordHash = "another-hash"
	same.SecurityCode = "another-code"
	rec := &testutil.FakeAuditRecorder{}
	svc := &Service{audit: rec}
	if err := svc.recordUserUpdate(context.Background(), testActor(4, "narges"), before, same, "پذیرش", "پذیرش"); err != nil {
		t.Fatal(err)
	}
	if len(rec.Events) != 0 {
		t.Fatal("بدون تغییر فیلد کسب‌وکار نباید رویداد ثبت شود")
	}
	after := before
	after.Name = "رضا"
	after.PhoneNumber = "09120000000"
	after.RoleID = 2
	after.PasswordHash = "another-hash"
	after.SecurityCode = "another-code"
	if err := svc.recordUserUpdate(context.Background(), testActor(4, "narges"), before, after, "پذیرش", "پزشک"); err != nil {
		t.Fatal(err)
	}
	got := rec.Events[0].Event
	if got.Type != audit.EventUserUpdated || got.Actor.Username != "narges" || got.Target.Name != "ali" {
		t.Fatalf("رویداد کاربر نادرست است: %+v", got)
	}
	raw, _ := json.Marshal(got.Details)
	text := string(raw)
	for _, hidden := range []string{secret, "another-hash", code, "another-code", "PasswordHash", "SecurityCode", "password"} {
		if strings.Contains(text, hidden) {
			t.Fatalf("جزئیات داده حساس دارد (%s): %s", hidden, text)
		}
	}
	details := got.Details.(map[string]any)
	phone := details["phone"].(fieldChange)
	if phone.Before != "*******4567" || phone.After != "*******0000" {
		t.Fatalf("تلفن ماسک نشده است: %#v", phone)
	}
	role := details["role"].(fieldChange)
	if role.Before != "پذیرش" || role.After != "پزشک" {
		t.Fatalf("نقش باید نام باشد نه شناسه: %#v", role)
	}
	if _, ok := details["family"]; ok {
		t.Fatal("نام خانوادگی عوض نشده نباید در جزئیات باشد")
	}
	if !strings.Contains(got.Description, "نام") || !strings.Contains(got.Description, "تلفن") || !strings.Contains(got.Description, "نقش") {
		t.Fatalf("توضیح فیلدها: %s", got.Description)
	}
}

// TestR1_21_PasswordChangedHasNoSecret تغییر رمز رویداد جدا است و خود رمز را ندارد.
func TestR1_21_PasswordChangedHasNoSecret(t *testing.T) {
	user := &User{ID: 1, Username: "ali", PasswordHash: "hashed-password"}
	rec := &testutil.FakeAuditRecorder{}
	svc := &Service{audit: rec}
	if err := svc.recordPasswordChanged(context.Background(), testActor(1, "ali"), user); err != nil {
		t.Fatal(err)
	}
	got := rec.Events[0].Event
	if got.Type != audit.EventPasswordChanged || got.Details != nil || got.Target.Name != "ali" {
		t.Fatalf("تغییر رمز نادرست است: %+v", got)
	}
	raw, _ := json.Marshal(got)
	if strings.Contains(string(raw), "hashed-password") {
		t.Fatal("هش رمز نباید در رویداد باشد")
	}
}
