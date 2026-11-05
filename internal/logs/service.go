package logs

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/listquery"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"gorm.io/gorm"
)

// chainReader فقط ردیف وضعیت زنجیره را می‌خواند.
// audit.Repository همین متد را دارد. بقیه متدهای مخزن ممیزی اینجا لازم نیستند.
type chainReader interface {
	LoadChainState() (audit.AuditChainState, error)
}

// openFindingReader یافته‌های بازی را می‌خواند که هنوز رفع نشده‌اند.
// audit.FindingStore همین متد را دارد.
type openFindingReader interface {
	ListAllOpen(ctx context.Context) ([]audit.IntegrityFinding, error)
}

// Service فهرست لاگ، فراداده فیلتر و وضعیت زنجیره را می‌سازد.
type Service struct {
	events   eventRepo
	chain    chainReader
	findings openFindingReader
}

// NewService سرویس لاگ را می‌سازد.
// db اتصال خواندن رویدادها و یافته‌هاست و chain وضعیت AuditChainStates را می‌دهد.
// اگر db تهی باشد یافته‌ها خالی می‌مانند تا تست بدون پایگاه همان وضعیت زنجیره را ببیند.
func NewService(db *gorm.DB, chain chainReader) *Service {
	svc := &Service{events: eventRepo{db: db}, chain: chain}
	if db != nil {
		svc.findings = audit.NewFindingStore(db)
	}
	return svc
}

// List یک صفحه از لاگ را با برچسب فارسی برمی‌گرداند.
// ctx لغو درخواست است. q فیلتر اعتبارسنجی‌شده است. onlyUserID لاگ خود کاربر است و برای فهرست مدیر تهی است.
// خروجی پاسخ استاندارد است. خطای پایگاه به پیام عمومی تبدیل می‌شود.
func (s *Service) List(ctx context.Context, q listquery.Query, onlyUserID *int) (listquery.Page[EventItem], error) {
	if s == nil || s.events.db == nil {
		return listquery.Page[EventItem]{}, apperror.ErrInternal
	}
	rows, total, err := s.events.list(ctx, q, onlyUserID)
	if err != nil {
		if errors.As(err, new(*apperror.AppError)) {
			return listquery.Page[EventItem]{}, err
		}
		return listquery.Page[EventItem]{}, apperror.New("DB_ERROR", "خطا در خواندن لاگ‌ها.", err.Error(), 500)
	}
	items := make([]EventItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, toItem(row))
	}
	return listquery.NewPage(items, total, q), nil
}

// Meta فهرست نوع رویداد، دلیل و نوع هدف را با برچسب فارسی برمی‌گرداند.
// ctx لغو درخواست است. خروجی برای کشوی فیلتر است و به پایگاه وصل نمی‌شود.
func (s *Service) Meta(ctx context.Context) (MetaResponse, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return MetaResponse{}, err
		}
	}
	return MetaResponse{
		EventTypes:  options(audit.ListEventTypes()),
		Reasons:     options(audit.ListReasons()),
		TargetTypes: options(audit.ListTargetTypes()),
	}, nil
}

// Integrity آخرین نتیجه بررسی زنجیره و یافته‌های باز را برمی‌گرداند.
// ctx لغو درخواست است. خروجی زمان بررسی، نتیجه، شناسه ردیف شکسته‌شده و فهرست نوع و نام است.
// FAU_STG.2: هش زنجیره در این پاسخ نیست.
func (s *Service) Integrity(ctx context.Context) (IntegrityResponse, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return IntegrityResponse{}, err
		}
	}
	if s == nil || s.chain == nil {
		return IntegrityResponse{}, apperror.ErrInternal
	}
	state, err := s.chain.LoadChainState()
	if err != nil {
		return IntegrityResponse{}, apperror.New("DB_ERROR", "خطا در خواندن وضعیت ممیزی.", err.Error(), 500)
	}
	rows, err := s.openFindings(ctx)
	if err != nil {
		return IntegrityResponse{}, apperror.New("DB_ERROR", "خطا در خواندن وضعیت ممیزی.", err.Error(), 500)
	}
	return chainIntegrity(state, rows), nil
}

// openFindings یافته‌های باز را می‌خواند.
// ctx لغو درخواست است. اگر مخزن نباشد خروجی خالی است تا پاسخ زنجیره بدون پایگاه بماند.
func (s *Service) openFindings(ctx context.Context) ([]audit.IntegrityFinding, error) {
	if s.findings == nil {
		return []audit.IntegrityFinding{}, nil
	}
	return s.findings.ListAllOpen(ctx)
}

// chainIntegrity وضعیت زنجیره و یافته‌ها را به پاسخ API تبدیل می‌کند.
// state ردیف AuditChainStates و rows یافته‌های باز است. خروجی هش ندارد.
func chainIntegrity(state audit.AuditChainState, rows []audit.IntegrityFinding) IntegrityResponse {
	items := make([]OpenFindingItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, OpenFindingItem{
			EntityType:      row.EntityType,
			EntityTypeLabel: audit.TargetTypeLabel(row.EntityType),
			EntityName:      row.EntityName,
			FirstSeenAt:     row.FirstSeenAt.UTC(),
		})
	}
	resp := IntegrityResponse{
		LastResult:        state.LastResult,
		BrokenAtID:        state.BrokenAtID,
		OpenFindingsCount: len(items),
		OpenFindings:      items,
	}
	if !state.LastCheckAt.IsZero() {
		at := state.LastCheckAt.UTC()
		resp.LastCheckAt = &at
	}
	return resp
}

// toItem یک ردیف اسکن‌شده را به پاسخ API تبدیل می‌کند.
// row ستون‌های جدول است. خروجی همان مقدارها به‌علاوه برچسب فارسی نوع و دلیل است.
func toItem(row eventRow) EventItem {
	return EventItem{
		ID:             row.ID,
		CreatedAt:      row.CreatedAt,
		ActorUsername:  row.ActorUsername,
		TargetName:     row.TargetName,
		TargetType:     row.TargetType,
		EventType:      row.EventType,
		EventTypeLabel: audit.EventLabel(row.EventType),
		Result:         row.Result,
		Reason:         row.Reason,
		ReasonLabel:    audit.ReasonLabel(row.Reason),
		IP:             row.IP,
		UserAgent:      row.UserAgent,
		Description:    row.Description,
		Details:        detailsObject(row.Details),
	}
}

// detailsObject فقط شیء JSON معتبر را برای پاسخ نگه می‌دارد.
// raw متن ستون Details است. خروجی nil یعنی خالی، نامعتبر، یا غیرشیء است و در JSON نمی‌آید.
// هش زنجیره در این ستون نیست. مقدار از قبل ماسک شده است.
func detailsObject(raw string) json.RawMessage {
	raw = strings.TrimSpace(raw)
	if raw == "" || !strings.HasPrefix(raw, "{") || !json.Valid([]byte(raw)) {
		return nil
	}
	return json.RawMessage(raw)
}

// options برچسب‌های کاتالوگ ممیزی را به شکل پاسخ فیلتر درمی‌آورد.
// items مقدار و برچسب فارسی‌اند. خروجی هرگز nil نیست تا JSON آرایه خالی بدهد.
func options(items []audit.CatalogItem) []Option {
	out := make([]Option, 0, len(items))
	for _, item := range items {
		out = append(out, Option{Value: item.Value, Label: item.Label})
	}
	return out
}
