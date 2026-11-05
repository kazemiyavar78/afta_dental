package session

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"gorm.io/gorm"
)

// memAudit ثبت‌کننده درون همین بسته است تا تست به testutil و چرخه import وابسته نشود.
type memAudit struct {
	events []audit.Event
}

// Record رویداد را نگه می‌دارد. ctx اگر لغو شده باشد خطا می‌دهد. e رویداد نشست است.
func (m *memAudit) Record(ctx context.Context, e audit.Event) error {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	m.events = append(m.events, e)
	return nil
}

// RecordTx همان Record است چون تست تراکنش پایگاه ندارد.
func (m *memAudit) RecordTx(_ *gorm.DB, e audit.Event) error {
	return m.Record(context.Background(), e)
}

type fixedLimit struct{ n int }

func (f fixedLimit) GetSettingInt(string) (int, error) { return f.n, nil }

type memSessions struct {
	items []Session
}

func (m *memSessions) Create(s *Session) error {
	m.items = append(m.items, *s)
	return nil
}

func (m *memSessions) FindByID(id uuid.UUID) (*Session, error) {
	for i := range m.items {
		if m.items[i].ID.UUID() == id {
			cp := m.items[i]
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *memSessions) FindByUserID(userID int) ([]Session, error) {
	var out []Session
	for _, item := range m.items {
		if item.PersonnelAccountID == userID {
			out = append(out, item)
		}
	}
	return out, nil
}

func (m *memSessions) Delete(id uuid.UUID) error {
	next := m.items[:0]
	for _, item := range m.items {
		if item.ID.UUID() != id {
			next = append(next, item)
		}
	}
	m.items = next
	return nil
}

func (m *memSessions) DeleteByUserID(userID int) error {
	next := m.items[:0]
	for _, item := range m.items {
		if item.PersonnelAccountID != userID {
			next = append(next, item)
		}
	}
	m.items = next
	return nil
}

func (m *memSessions) CountByUserID(userID int) (int64, error) {
	var n int64
	for _, item := range m.items {
		if item.PersonnelAccountID == userID {
			n++
		}
	}
	return n, nil
}

func (m *memSessions) DeleteOldestByUserID(userID int) (*Session, error) {
	idx := -1
	for i := range m.items {
		if m.items[i].PersonnelAccountID != userID {
			continue
		}
		if idx < 0 || m.items[i].CreationTime.Before(m.items[idx].CreationTime) {
			idx = i
		}
	}
	if idx < 0 {
		return nil, nil
	}
	removed := m.items[idx]
	m.items = append(m.items[:idx], m.items[idx+1:]...)
	return &removed, nil
}

func newCloseService(repo *memSessions, max int, rec *memAudit) *Service {
	return &Service{
		repo:     repo,
		settings: fixedLimit{n: max},
		audit:    rec,
		namer: func(userID int) (string, error) {
			if userID == 1 {
				return "ali", nil
			}
			return "other", nil
		},
	}
}

func actorOf(id int, name string) audit.Actor {
	return audit.Actor{UserID: &id, Username: name, IP: "203.0.113.4", UserAgent: "Browser"}
}

// TestR1_06_MaxSessionsRecordsSystemLogout برای هر نشست حذف‌شده به‌خاطر سقف، فاعل سیستم و دلیل max_sessions می‌نویسد.
func TestR1_06_MaxSessionsRecordsSystemLogout(t *testing.T) {
	oldID := NewMSSQLUUID()
	repo := &memSessions{items: []Session{{
		ID:                 oldID,
		Ip:                 "198.51.100.1",
		Browser:            "Old",
		CreationTime:       time.Now().Add(-time.Hour),
		PersonnelAccountID: 1,
	}}}
	rec := &memAudit{}
	svc := newCloseService(repo, 1, rec)
	if _, err := svc.CreateSession(context.Background(), 1, "203.0.113.8", "NewBrowser"); err != nil {
		t.Fatalf("ایجاد نشست: %v", err)
	}
	if len(rec.events) != 2 {
		t.Fatalf("باید پایان نشست و خروج ثبت شود، %d رویداد است", len(rec.events))
	}
	term, logout := rec.events[0], rec.events[1]
	if term.Type != audit.EventSessionTerminated || logout.Type != audit.EventLogout {
		t.Fatalf("ترتیب رویداد نادرست است: %s سپس %s", term.Type, logout.Type)
	}
	if term.Reason != audit.ReasonMaxSessions || logout.Reason != audit.ReasonMaxSessions {
		t.Fatalf("دلیل باید سقف نشست باشد: %s %s", term.Reason, logout.Reason)
	}
	if term.Actor.Username != "system" || term.Target.Name != "ali" || term.Target.Type != audit.TargetUser {
		t.Fatalf("فاعل سیستم و هدف صاحب نشست نیست: %+v", term)
	}
	if logout.Actor.Username != "ali" {
		t.Fatalf("خروج باید برای صاحب نشست باشد: %s", logout.Actor.Username)
	}
	assertNoSessionUUID(t, oldID, term.Description, logout.Description)
	prefix, _ := term.Details.(map[string]any)["session_prefix"].(string)
	if prefix != oldID.String()[:8] {
		t.Fatalf("پیشوند نشست باید ۸ نویسه اول باشد: %s", prefix)
	}
}

// TestR1_07_DeleteSessionRecordsActorTargetAndReason حذف توسط مدیر و توسط خود کاربر دو دلیل متفاوت دارد.
func TestR1_07_DeleteSessionRecordsActorTargetAndReason(t *testing.T) {
	repo := &memSessions{}
	rec := &memAudit{}
	svc := newCloseService(repo, 3, rec)
	owned := &Session{ID: NewMSSQLUUID(), Ip: "198.51.100.2", Browser: "Own", PersonnelAccountID: 1, CreationTime: time.Now()}
	repo.items = append(repo.items, *owned)

	if err := svc.DeleteSession(context.Background(), owned.ID.UUID(), actorOf(9, "admin"), true); err != nil {
		t.Fatalf("حذف مدیر: %v", err)
	}
	term := rec.events[0]
	logout := rec.events[1]
	if term.Actor.Username != "admin" || term.Target.Name != "ali" || term.Reason != audit.ReasonAdminTerminated {
		t.Fatalf("حذف مدیر نادرست است: actor=%s target=%s reason=%s", term.Actor.Username, term.Target.Name, term.Reason)
	}
	if logout.Type != audit.EventLogout || logout.Reason != audit.ReasonAdminTerminated || logout.Actor.Username != "ali" {
		t.Fatalf("خروج صاحب نشست نادرست است: %+v", logout)
	}
	assertNoSessionUUID(t, owned.ID, term.Description, logout.Description)

	self := &Session{ID: NewMSSQLUUID(), Ip: "198.51.100.3", Browser: "Self", PersonnelAccountID: 1, CreationTime: time.Now()}
	repo.items = append(repo.items, *self)
	rec.events = nil
	if err := svc.DeleteSession(context.Background(), self.ID.UUID(), actorOf(1, "ali"), false); err != nil {
		t.Fatalf("حذف خود کاربر: %v", err)
	}
	if rec.events[0].Reason != audit.ReasonUserLogout || rec.events[0].Actor.Username != "ali" {
		t.Fatalf("خروج خودش باید user_logout با فاعل خودش باشد: %+v", rec.events[0])
	}
	assertNoSessionUUID(t, self.ID, rec.events[0].Description, rec.events[1].Description)
}

// TestR1_07_DeleteAllSessionsUsesReason برای هر نشست یک جفت پایان و خروج با دلیل داده‌شده می‌سازد.
func TestR1_07_DeleteAllSessionsUsesReason(t *testing.T) {
	first := NewMSSQLUUID()
	second := NewMSSQLUUID()
	repo := &memSessions{items: []Session{
		{ID: first, PersonnelAccountID: 1, Ip: "1.1.1.1", Browser: "A"},
		{ID: second, PersonnelAccountID: 1, Ip: "1.1.1.2", Browser: "B"},
	}}
	rec := &memAudit{}
	svc := newCloseService(repo, 3, rec)
	if err := svc.DeleteAllUserSessions(context.Background(), 1, audit.SystemActor(), audit.ReasonSecurityCodeFailed); err != nil {
		t.Fatalf("حذف همه: %v", err)
	}
	if len(rec.events) != 4 {
		t.Fatalf("دو نشست باید چهار رویداد بسازند، %d ساخته شد", len(rec.events))
	}
	for i, ev := range rec.events {
		if ev.Reason != audit.ReasonSecurityCodeFailed || ev.Target.Name != "ali" {
			t.Fatalf("رویداد %d نادرست است: %+v", i, ev)
		}
		assertNoSessionUUID(t, first, ev.Description)
		assertNoSessionUUID(t, second, ev.Description)
	}
	if len(repo.items) != 0 {
		t.Fatal("نشست‌ها باید حذف شده باشند")
	}
}

// assertNoSessionUUID می‌گوید توضیح رویداد شناسه نشست را ندارد.
func assertNoSessionUUID(t *testing.T, id MSSQLUUID, descriptions ...string) {
	t.Helper()
	full := id.String()
	for _, text := range descriptions {
		if strings.Contains(text, full) || strings.Contains(strings.ToLower(text), strings.ToLower(full)) {
			t.Fatalf("شناسه نشست در توضیح آمده است: %s", text)
		}
	}
}
