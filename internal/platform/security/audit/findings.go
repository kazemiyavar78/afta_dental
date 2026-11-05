package audit

import (
	"context"
	"errors"
	"time"
)

// errNilIntegrityScan یعنی هم‌ترازی یافته بدون مخزن یا ثبت‌کننده صدا زده شده است.
var errNilIntegrityScan = errors.New("بررسی یکپارچگی بدون مخزن یا ثبت‌کننده است")

// IntegritySubject یک ردیف است که HMAC ذخیره‌شده با داده فعلی نمی‌خواند.
// EntityType یکی از user، role یا setting است. EntityID شناسه همان ردیف است.
type IntegritySubject struct {
	EntityType string
	EntityID   string
	EntityName string
}

// IntegrityFinding یک نقض باز یا بسته‌شده یکپارچگی است.
// ResolvedAt تهی یعنی یافته هنوز باز است و دور بعد نباید رویداد تازه بسازد.
// FDP_SDI.2: هشدار تکراری برای همان ردیف ثبت نمی‌شود.
type IntegrityFinding struct {
	ID          int        `gorm:"column:ID;primaryKey;autoIncrement"`
	EntityType  string     `gorm:"column:EntityType;size:50;not null;index:IX_IntegrityFindings_Entity,priority:1"`
	EntityID    string     `gorm:"column:EntityID;size:64;not null;index:IX_IntegrityFindings_Entity,priority:2"`
	EntityName  string     `gorm:"column:EntityName;size:200;not null"`
	FirstSeenAt time.Time  `gorm:"column:FirstSeenAt;not null"`
	LastSeenAt  time.Time  `gorm:"column:LastSeenAt;not null"`
	ResolvedAt  *time.Time `gorm:"column:ResolvedAt"`
}

// TableName نام جدول IntegrityFindings را برمی‌گرداند.
func (IntegrityFinding) TableName() string { return "IntegrityFindings" }

// FindingStore خواندن و نوشتن یافته‌های یکپارچگی است.
// ListOpen فقط ردیف‌های باز همان نوع را می‌دهد. ListAllOpen برای نمایش به مدیر است.
type FindingStore interface {
	ListOpen(ctx context.Context, entityType string) ([]IntegrityFinding, error)
	ListAllOpen(ctx context.Context) ([]IntegrityFinding, error)
	Insert(ctx context.Context, row *IntegrityFinding) error
	Touch(ctx context.Context, id int, lastSeen time.Time) error
	Resolve(ctx context.Context, id int, at time.Time) error
	Delete(ctx context.Context, id int) error
	Reopen(ctx context.Context, id int) error
}

// ApplyIntegrityFindings یافته‌های باز یک نوع را با ردیف‌های نامعتبر همین دور هم‌تراز می‌کند.
// entityType نوع همین بررسی است. invalid می‌تواند خالی باشد تا یافته‌های قبلی همان نوع بسته شوند.
// now زمان این دور است. یافته جدید یک DATA_TAMPERING می‌گیرد و یافته تکراری فقط LastSeenAt را جلو می‌برد.
// ردیفی که دیگر نامعتبر نیست ResolvedAt و یک INTEGRITY_RESTORED می‌گیرد.
// خروجی خطای مخزن یا ثبت ممیزی است. نوع‌های دیگر دست نمی‌خورند.
func ApplyIntegrityFindings(ctx context.Context, store FindingStore, rec Recorder, entityType string, invalid []IntegritySubject, now time.Time) error {
	if store == nil || rec == nil {
		return errNilIntegrityScan
	}
	ctx, now, err := prepareIntegrityScan(ctx, now)
	if err != nil {
		return err
	}
	open, err := store.ListOpen(ctx, entityType)
	if err != nil {
		return err
	}
	grouped := groupOpenFindings(open)
	seen, err := recordBrokenRows(ctx, store, rec, entityType, invalid, grouped, now)
	if err != nil {
		return err
	}
	return closeHealthyRows(ctx, store, rec, grouped, seen, now)
}

// prepareIntegrityScan زمینه و زمان این دور را آماده می‌کند.
// ctx تهی به زمینه پس‌زمینه تبدیل می‌شود. now صفر، زمان فعلی است.
// خروجی خطای لغو زمینه است و در آن صورت چیزی نوشته نمی‌شود.
func prepareIntegrityScan(ctx context.Context, now time.Time) (context.Context, time.Time, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return ctx, time.Time{}, err
	}
	if now.IsZero() {
		now = time.Now()
	}
	return ctx, now.UTC().Truncate(time.Microsecond), nil
}

// groupOpenFindings یافته‌های باز را بر اساس شناسه ردیف گروه می‌کند.
// rows خروجی ListOpen است. اگر برای یک شناسه چند ردیف باز مانده باشد همه در یک گروه می‌مانند.
func groupOpenFindings(rows []IntegrityFinding) map[string][]IntegrityFinding {
	grouped := make(map[string][]IntegrityFinding, len(rows))
	for _, row := range rows {
		grouped[row.EntityID] = append(grouped[row.EntityID], row)
	}
	return grouped
}

// recordBrokenRows برای هر ردیف نامعتبر تازه رویداد می‌سازد و تکراری را فقط لمس می‌کند.
// invalid فهرست همین دور است. grouped یافته‌های باز قبلی همان نوع است.
// خروجی شناسه‌هایی است که هنوز خراب‌اند تا ردیف سالم اشتباهی بسته نشود.
func recordBrokenRows(ctx context.Context, store FindingStore, rec Recorder, entityType string, invalid []IntegritySubject, grouped map[string][]IntegrityFinding, now time.Time) (map[string]struct{}, error) {
	seen := make(map[string]struct{}, len(invalid))
	for _, row := range invalid {
		row = normalizeSubject(entityType, row)
		if row.EntityID == "" {
			continue
		}
		if _, ok := seen[row.EntityID]; ok {
			continue
		}
		seen[row.EntityID] = struct{}{}
		if existing := grouped[row.EntityID]; len(existing) > 0 {
			if err := touchFindings(ctx, store, existing, now); err != nil {
				return nil, err
			}
			continue
		}
		if err := openFinding(ctx, store, rec, row, now); err != nil {
			return nil, err
		}
	}
	return seen, nil
}

// closeHealthyRows یافته بازی را که در این دور سالم شده می‌بندد.
// grouped همه یافته‌های باز این نوع است و seen شناسه‌های هنوز خراب است.
// برای هر شناسه سالم‌شده یک INTEGRITY_RESTORED ثبت می‌شود.
func closeHealthyRows(ctx context.Context, store FindingStore, rec Recorder, grouped map[string][]IntegrityFinding, seen map[string]struct{}, now time.Time) error {
	for id, existing := range grouped {
		if _, ok := seen[id]; ok || len(existing) == 0 {
			continue
		}
		if err := closeFinding(ctx, store, rec, existing, now); err != nil {
			return err
		}
	}
	return nil
}

// touchFindings فقط LastSeenAt ردیف‌های باز تکراری را جلو می‌برد.
// existing یافته‌های همان شناسه است. رویدادی ثبت نمی‌شود.
func touchFindings(ctx context.Context, store FindingStore, existing []IntegrityFinding, now time.Time) error {
	for _, row := range existing {
		if err := store.Touch(ctx, row.ID, now); err != nil {
			return err
		}
	}
	return nil
}

// openFinding یک یافته تازه درج می‌کند و یک DATA_TAMPERING می‌نویسد.
// row ردیف نامعتبر این دور است. اگر ثبت ممیزی شکست بخورد همان یافته حذف می‌شود تا دور بعد دوباره هشدار بدهد.
// Reason برابر integrity_violation و هدف همان ردیف است.
func openFinding(ctx context.Context, store FindingStore, rec Recorder, row IntegritySubject, now time.Time) error {
	stored := &IntegrityFinding{
		EntityType:  row.EntityType,
		EntityID:    row.EntityID,
		EntityName:  row.EntityName,
		FirstSeenAt: now,
		LastSeenAt:  now,
	}
	if err := store.Insert(ctx, stored); err != nil {
		return err
	}
	if err := rec.Record(ctx, integrityEvent(EventDataTampering, ResultFailure, row)); err != nil {
		if delErr := store.Delete(ctx, stored.ID); delErr != nil {
			return errors.Join(err, delErr)
		}
		return err
	}
	return nil
}

// closeFinding یافته‌های باز یک ردیف سالم‌شده را می‌بندد و INTEGRITY_RESTORED می‌نویسد.
// existing ردیف‌های باز همان شناسه است. اگر ثبت ممیزی شکست بخورد ResolvedAt دوباره تهی می‌شود.
func closeFinding(ctx context.Context, store FindingStore, rec Recorder, existing []IntegrityFinding, now time.Time) error {
	for _, row := range existing {
		if err := store.Resolve(ctx, row.ID, now); err != nil {
			return err
		}
	}
	subject := IntegritySubject{
		EntityType: existing[0].EntityType,
		EntityID:   existing[0].EntityID,
		EntityName: existing[0].EntityName,
	}
	if err := rec.Record(ctx, integrityEvent(EventIntegrityRestored, ResultSuccess, subject)); err != nil {
		if reopenErr := reopenFindings(ctx, store, existing); reopenErr != nil {
			return errors.Join(err, reopenErr)
		}
		return err
	}
	return nil
}

// reopenFindings ResolvedAt ردیف‌هایی را که رویداد بستن‌شان ثبت نشد تهی می‌کند.
// existing همان ردیف‌هایی است که Resolve شده‌اند. خروجی خطای اولین شکست است.
func reopenFindings(ctx context.Context, store FindingStore, existing []IntegrityFinding) error {
	for _, row := range existing {
		if err := store.Reopen(ctx, row.ID); err != nil {
			return err
		}
	}
	return nil
}

// normalizeSubject نوع بررسی و طول ستون‌ها را روی ردیف نامعتبر اعمال می‌کند.
// entityType نوع همین اسکن است و نوع داخل row را جایگزین می‌کند تا یافته با اسکن یکی بماند.
func normalizeSubject(entityType string, row IntegritySubject) IntegritySubject {
	row.EntityType = clip(entityType, maxTargetType)
	row.EntityID = clip(row.EntityID, maxTargetID)
	row.EntityName = clip(row.EntityName, maxTargetName)
	return row
}

// integrityEvent رویداد هشدار یا رفع یک ردیف را می‌سازد.
// eventType برابر DATA_TAMPERING یا INTEGRITY_RESTORED است. result موفقیت یا شکست است.
// هدف همان ردیف است. توضیح فارسی نوع و نام را دارد تا منبع هشدار از روی متن معلوم باشد.
func integrityEvent(eventType, result string, row IntegritySubject) Event {
	name := row.EntityName
	if name == "" {
		name = row.EntityID
	}
	kind := TargetTypeLabel(row.EntityType)
	verb := "نقض یکپارچگی "
	reason := ReasonIntegrityViolation
	if eventType == EventIntegrityRestored {
		verb = "رفع نقض یکپارچگی "
		reason = ""
	}
	return Event{
		Type:        eventType,
		Result:      result,
		Reason:      reason,
		Actor:       SystemActor(),
		Target:      Target{Type: row.EntityType, ID: row.EntityID, Name: row.EntityName},
		Description: verb + kind + " «" + name + "»",
	}
}
