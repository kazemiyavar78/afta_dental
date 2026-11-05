package audit

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
)

// memoryFindings یافته‌ها را در حافظه نگه می‌دارد تا تست واحد به پایگاه وصل نشود.
type memoryFindings struct {
	next int
	rows []IntegrityFinding
}

func (m *memoryFindings) ListOpen(_ context.Context, entityType string) ([]IntegrityFinding, error) {
	out := make([]IntegrityFinding, 0)
	for _, row := range m.rows {
		if row.EntityType == entityType && row.ResolvedAt == nil {
			out = append(out, row)
		}
	}
	return out, nil
}

func (m *memoryFindings) ListAllOpen(context.Context) ([]IntegrityFinding, error) {
	out := make([]IntegrityFinding, 0)
	for _, row := range m.rows {
		if row.ResolvedAt == nil {
			out = append(out, row)
		}
	}
	return out, nil
}

func (m *memoryFindings) Insert(_ context.Context, row *IntegrityFinding) error {
	m.next++
	stored := *row
	stored.ID = m.next
	m.rows = append(m.rows, stored)
	row.ID = stored.ID
	return nil
}

func (m *memoryFindings) Touch(_ context.Context, id int, lastSeen time.Time) error {
	row := m.byID(id)
	if row == nil {
		return errors.New("یافته پیدا نشد")
	}
	row.LastSeenAt = lastSeen
	return nil
}

func (m *memoryFindings) Resolve(_ context.Context, id int, at time.Time) error {
	row := m.byID(id)
	if row == nil {
		return errors.New("یافته پیدا نشد")
	}
	seen := at
	row.ResolvedAt = &seen
	return nil
}

func (m *memoryFindings) Delete(_ context.Context, id int) error {
	kept := m.rows[:0]
	for _, row := range m.rows {
		if row.ID != id {
			kept = append(kept, row)
		}
	}
	m.rows = kept
	return nil
}

func (m *memoryFindings) Reopen(_ context.Context, id int) error {
	row := m.byID(id)
	if row == nil {
		return errors.New("یافته پیدا نشد")
	}
	row.ResolvedAt = nil
	return nil
}

func (m *memoryFindings) byID(id int) *IntegrityFinding {
	for i := range m.rows {
		if m.rows[i].ID == id {
			return &m.rows[i]
		}
	}
	return nil
}

func (m *memoryFindings) openCount(entityType string) int {
	n := 0
	for _, row := range m.rows {
		if row.EntityType == entityType && row.ResolvedAt == nil {
			n++
		}
	}
	return n
}

// memoryRecorder رویدادهای تست را نگه می‌دارد.
type memoryRecorder struct {
	events []Event
}

func (m *memoryRecorder) Record(_ context.Context, e Event) error {
	m.events = append(m.events, e)
	return nil
}

func (m *memoryRecorder) RecordTx(_ *gorm.DB, event Event) error {
	return m.Record(context.Background(), event)
}

func countType(events []Event, eventType string) int {
	n := 0
	for _, event := range events {
		if event.Type == eventType {
			n++
		}
	}
	return n
}

// TestD23_TwoBrokenRowsThenRepeatThenRestore دو ردیف خراب دو یافته و دو رویداد می‌سازد.
// دور دوم رویداد تازه ندارد. سالم شدن یکی ResolvedAt و INTEGRITY_RESTORED می‌گذارد.
// D-23 / FDP_SDI.2: همان ردیف در دورهای بعد هشدار تکراری نمی‌گیرد.
func TestD23_TwoBrokenRowsThenRepeatThenRestore(t *testing.T) {
	store := &memoryFindings{}
	rec := &memoryRecorder{}
	first := time.Date(2026, 11, 6, 8, 0, 0, 0, time.UTC)
	broken := []IntegritySubject{
		{EntityID: "4", EntityName: "ali"},
		{EntityID: "9", EntityName: "reza"},
	}
	if err := ApplyIntegrityFindings(context.Background(), store, rec, TargetUser, broken, first); err != nil {
		t.Fatalf("دور اول: %v", err)
	}
	if store.openCount(TargetUser) != 2 || countType(rec.events, EventDataTampering) != 2 {
		t.Fatalf("دور اول باید دو یافته و دو هشدار بسازد: findings=%d events=%d", store.openCount(TargetUser), len(rec.events))
	}
	for _, event := range rec.events {
		if event.Reason != ReasonIntegrityViolation || event.Actor.Username != "system" || event.Target.Type != TargetUser {
			t.Fatalf("هشدار باید دلیل و هدف همان ردیف را داشته باشد: %+v", event)
		}
	}
	if store.rows[0].FirstSeenAt != first || store.rows[1].EntityName != "reza" {
		t.Fatalf("یافته اول نادرست است: %+v", store.rows)
	}

	second := first.Add(15 * time.Minute)
	before := len(rec.events)
	if err := ApplyIntegrityFindings(context.Background(), store, rec, TargetUser, broken, second); err != nil {
		t.Fatalf("دور دوم: %v", err)
	}
	if len(rec.events) != before || store.openCount(TargetUser) != 2 {
		t.Fatalf("دور دوم نباید رویداد یا یافته تازه بسازد: events=%d open=%d", len(rec.events)-before, store.openCount(TargetUser))
	}
	if !store.rows[0].LastSeenAt.Equal(second) || !store.rows[0].FirstSeenAt.Equal(first) {
		t.Fatalf("فقط LastSeenAt باید جلو برود: %+v", store.rows[0])
	}

	third := second.Add(15 * time.Minute)
	if err := ApplyIntegrityFindings(context.Background(), store, rec, TargetUser, broken[:1], third); err != nil {
		t.Fatalf("دور رفع: %v", err)
	}
	if store.rows[1].ResolvedAt == nil || !store.rows[1].ResolvedAt.Equal(third) {
		t.Fatalf("ردیف سالم‌شده باید ResolvedAt داشته باشد: %+v", store.rows[1])
	}
	if store.rows[0].ResolvedAt != nil || store.openCount(TargetUser) != 1 {
		t.Fatalf("ردیف هنوز خراب باید باز بماند: %+v open=%d", store.rows[0], store.openCount(TargetUser))
	}
	if countType(rec.events, EventIntegrityRestored) != 1 || countType(rec.events, EventDataTampering) != 2 {
		t.Fatalf("باید یک رفع و همان دو هشدار قبلی باشد: %+v", rec.events)
	}
	restored := rec.events[len(rec.events)-1]
	if restored.Type != EventIntegrityRestored || restored.Target.ID != "9" || restored.Target.Name != "reza" || restored.Result != ResultSuccess {
		t.Fatalf("رویداد رفع نادرست است: %+v", restored)
	}
	if EventLabel(EventIntegrityRestored) != "رفع نقض یکپارچگی" || !strings.Contains(restored.Description, "رفع نقض یکپارچگی") {
		t.Fatalf("برچسب رفع نادرست است: %s %s", EventLabel(EventIntegrityRestored), restored.Description)
	}
}

// TestD23_ApplyDoesNotResolveAnotherEntityType یافته باز نقش را هنگام اسکن کاربر نمی‌بندد.
func TestD23_ApplyDoesNotResolveAnotherEntityType(t *testing.T) {
	store := &memoryFindings{}
	rec := &memoryRecorder{}
	now := time.Date(2026, 11, 6, 9, 0, 0, 0, time.UTC)
	if err := ApplyIntegrityFindings(context.Background(), store, rec, TargetRole, []IntegritySubject{{EntityID: "2", EntityName: "reception"}}, now); err != nil {
		t.Fatal(err)
	}
	if err := ApplyIntegrityFindings(context.Background(), store, rec, TargetUser, nil, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if store.openCount(TargetRole) != 1 || countType(rec.events, EventIntegrityRestored) != 0 {
		t.Fatalf("نقش باز باید بماند: %+v events=%+v", store.rows, rec.events)
	}
}
