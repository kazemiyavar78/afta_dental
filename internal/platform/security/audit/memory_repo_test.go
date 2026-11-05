package audit

import (
	"context"
	"sync"

	"gorm.io/gorm"
)

// memoryRepo زنجیره را در حافظه نگه می‌دارد تا تست واحد به پایگاه وصل نشود.
type memoryRepo struct {
	mu             sync.Mutex
	events         []SecurityEvent
	nextID         int64
	state          AuditChainState
	findAfterCalls int
}

// newMemoryRepo مخزن خالی تست را برمی‌گرداند.
func newMemoryRepo() *memoryRepo {
	return &memoryRepo{}
}

// AppendChained رویداد را با هش آخرین ردیف، یا لنگر اگر ردیفی نمانده باشد، اضافه می‌کند.
// build ردیف را از PrevHash می‌سازد. خروجی خطای build است.
func (f *memoryRepo) AppendChained(build func(prevHash string) (*SecurityEvent, error)) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	prev := ""
	if n := len(f.events); n > 0 {
		prev = f.events[n-1].RowHash
	} else {
		prev = f.state.AnchorHash
	}
	event, err := build(prev)
	if err != nil {
		return err
	}
	f.nextID++
	stored := *event
	stored.ID = f.nextID
	f.events = append(f.events, stored)
	return nil
}

// AppendChainedCtx همان AppendChained است و اگر ctx تمام شده باشد چیزی ذخیره نمی‌کند.
// ctx لغو تست است. build ردیف را می‌سازد.
func (f *memoryRepo) AppendChainedCtx(ctx context.Context, build func(prevHash string) (*SecurityEvent, error)) error {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	return f.AppendChained(build)
}

// AppendInTx در تست پایگاه ندارد و رویداد را در حافظه اضافه می‌کند.
// tx نادیده گرفته می‌شود. build همان سازنده ردیف است.
func (f *memoryRepo) AppendInTx(_ *gorm.DB, build func(prevHash string) (*SecurityEvent, error)) error {
	return f.AppendChained(build)
}

// FindAfterID ردیف‌های بعد از afterID را تا limit برمی‌گرداند و تعداد فراخوانی را می‌شمارد.
func (f *memoryRepo) FindAfterID(afterID int64, limit int) ([]SecurityEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.findAfterCalls++
	out := make([]SecurityEvent, 0)
	for _, event := range f.events {
		if event.ID <= afterID {
			continue
		}
		out = append(out, event)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

// Count تعداد ردیف‌های حافظه را برمی‌گرداند.
func (f *memoryRepo) Count() (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return int64(len(f.events)), nil
}

// GetTableSizeMB در تست واحد حجم واقعی ندارد و صفر برمی‌گرداند.
func (f *memoryRepo) GetTableSizeMB() (float64, error) {
	return 0, nil
}

// LoadChainState کپی وضعیت ذخیره‌شده را برمی‌گرداند.
func (f *memoryRepo) LoadChainState() (AuditChainState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	state := f.state
	if state.ID == 0 {
		state.ID = chainStateRowID
	}
	return state, nil
}

// SaveChainState کپی وضعیت را نگه می‌دارد.
func (f *memoryRepo) SaveChainState(state *AuditChainState) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if state == nil {
		return errNilChainState
	}
	f.state = *state
	f.state.ID = chainStateRowID
	return nil
}

// DeleteVerifiedPrefix ردیف‌های ID<=maxID را حذف می‌کند و وضعیت برگشتی apply را ذخیره می‌کند.
func (f *memoryRepo) DeleteVerifiedPrefix(maxID int64, limit int, apply func([]SecurityEvent) (*AuditChainState, error)) ([]SecurityEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	found := make([]SecurityEvent, 0)
	for _, event := range f.events {
		if event.ID > maxID {
			continue
		}
		found = append(found, event)
		if limit > 0 && len(found) >= limit {
			break
		}
	}
	if len(found) == 0 {
		return nil, nil
	}
	updated, err := apply(found)
	if err != nil || updated == nil {
		return nil, err
	}
	drop := make(map[int64]struct{}, len(found))
	for _, row := range found {
		drop[row.ID] = struct{}{}
	}
	kept := make([]SecurityEvent, 0, len(f.events)-len(found))
	for _, event := range f.events {
		if _, ok := drop[event.ID]; !ok {
			kept = append(kept, event)
		}
	}
	f.events = kept
	f.state = *updated
	f.state.ID = chainStateRowID
	return found, nil
}

// hasID می‌گوید ردیف با این شناسه هنوز در حافظه هست یا نه.
func (f *memoryRepo) hasID(id int64) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, event := range f.events {
		if event.ID == id {
			return true
		}
	}
	return false
}
