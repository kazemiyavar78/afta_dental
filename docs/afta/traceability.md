# ردیابی الزام افتا

جدول الزام پروفایل حفاظتی، ایراد مشاور، کد، تست و وضعیت. این فایل تازه است و هنوز ماتریس کامل الزام‌ها را ندارد.

| الزام PP | ایراد مشاور | فایل کد | تست | وضعیت |
|---|---|---|---|---|
| — | — | `internal/platform/testutil`، `internal/platform/security/audit/logger.go`، `scripts/afta-check.ps1`، `frontend/src/test/renderWithProviders.tsx` | `TestWriteError_AppErrorReturnsCodeAndMessage`، `TestWriteError_UnknownErrorIsGeneric500`، `TestComputeRowHash`، `TestIntegration_AutoMigrate_AllModelsCreate`، `PageHeader` | منتظر اجرای تست DB |
| FPT_FLS.1 | D-01 نبودن `LogCleanupSummaries` و `UploadCounters` در AutoMigrate، و ادامهٔ سرور پس از خطای schema | `internal/platform/migrate/migrate.go`، `cmd/server/main.go` | `TestCombineMigrateErrors_EmptyIsNil`، `TestCombineMigrateErrors_NamesEveryTable`، `TestModelTableName_UsesTableNameThenTypeName`، `TestIntegration_AutoMigrate_AllModelsCreate`، `TestIntegration_LogCleanupAndUploadCounterTablesExist` | منتظر تست DB |
| — | D-02 Funds.Balance float | `internal/fund/model.go` | — | فاز ۵؛ نیاز به migration داده. نوع ستون الان float می‌ماند |
| — | — | `.github/workflows/ci.yml` | `govulncheck ./...`، `npm audit --omit=dev --audit-level=high` | غیرمسدود (`continue-on-error`). فاز ۵: `xlsx@0.18.5` جایگزین شود و هر دو بررسی blocking شوند |
| — | D-08 خطای bind ورودی HTTP 500 | `internal/platform/apperror/apperror.go`، `internal/platform/middleware/bind.go`، handlerهای `ShouldBindJSON` و `ShouldBindQuery` | `TestD08_BindJSON_EmptyOrIncompleteReturnsValidationError`، `TestD08_BindJSON_ValidBodyFillsStruct`، `TestD08_BindQuery_InvalidValueDoesNotLeakFieldValue` | انجام‌شده — منتظر مشاور |
| — | D-09 چاپ مقدار تنظیم در stdout | `internal/platform/security/settings/service.go` | بررسی `fmt.Print` در `internal/`؛ مورد `UpdateSettingByID` حذف شد. `loginguard/guard.go` هنوز دو `fmt.Println` برای ساعت کاری دارد | انجام‌شده — منتظر مشاور |
