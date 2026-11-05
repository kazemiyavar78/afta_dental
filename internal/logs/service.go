package logs

import (
	"context"
	"errors"

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

// Service فهرست لاگ، فراداده فیلتر و وضعیت زنجیره را می‌سازد.
type Service struct {
	events eventRepo
	chain  chainReader
}

// NewService سرویس لاگ را می‌سازد.
// db اتصال خواندن رویدادهاست و chain وضعیت AuditChainStates را می‌دهد.
// خروجی سرویس آماده هندلر است.
func NewService(db *gorm.DB, chain chainReader) *Service {
	return &Service{events: eventRepo{db: db}, chain: chain}
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

// Integrity آخرین نتیجه بررسی زنجیره را برمی‌گرداند.
// ctx لغو درخواست است. خروجی زمان بررسی، نتیجه و شناسه ردیف شکسته‌شده است.
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
	resp := IntegrityResponse{LastResult: state.LastResult, BrokenAtID: state.BrokenAtID}
	if !state.LastCheckAt.IsZero() {
		at := state.LastCheckAt.UTC()
		resp.LastCheckAt = &at
	}
	return resp, nil
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
	}
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
