# ردیابی الزام افتا

جدول الزام پروفایل حفاظتی، ایراد مشاور، کد، تست و وضعیت. این فایل تازه است و هنوز ماتریس کامل الزام‌ها را ندارد.

| الزام PP | ایراد مشاور | فایل کد | تست | وضعیت |
|---|---|---|---|---|
| — | — | `internal/platform/testutil`، `internal/platform/security/audit/logger.go`، `scripts/afta-check.ps1`، `frontend/src/test/renderWithProviders.tsx` | `TestWriteError_AppErrorReturnsCodeAndMessage`، `TestWriteError_UnknownErrorIsGeneric500`، `TestComputeRowHash`، `TestIntegration_AutoMigrate_AllModelsCreate`، `PageHeader` | منتظر اجرای تست DB |
| — | — | `.github/workflows/ci.yml` | `govulncheck ./...`، `npm audit --omit=dev --audit-level=high` | غیرمسدود (`continue-on-error`). فاز ۵: `xlsx@0.18.5` جایگزین شود و هر دو بررسی blocking شوند |
