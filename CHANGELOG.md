# تغییرات

## 2026-11-05

- fix(migrate): AutoMigrate تنها مسیر schema است. `LogCleanupSummary` و `UploadCounter` به فهرست راه‌اندازی اضافه شدند. خطای هر جدول با نام جدول برمی‌گردد و سرور متوقف می‌شود (FPT_FLS.1، D-01). فایل‌های `migrations/` حذف شدند چون هرگز اجرا نشده بودند. `Funds.Balance` همچنان float است و به فاز ۵ موکول شد (D-02). وضعیت تست پایگاه: منتظر اجرای کاربر.
- test(platform): زیرساخت تست مشترک (`internal/platform/testutil`)، تست integration برای AutoMigrate، helper فرانت `renderWithProviders` و اسکریپت `scripts/afta-check.ps1`. منطق برنامه عوض نشده است. وضعیت تست پایگاه: منتظر اجرای کاربر.
- ci: GitHub Actions برای `go vet`، تست واحد Go، `govulncheck`، و تست و build فرانت بدون پایگاه داده. `govulncheck` و `npm audit` تا فاز ۵ blocking نیستند چون `xlsx@0.18.5` آسیب‌پذیری شناخته‌شده دارد و باید در آن فاز جایگزین و سپس blocking شود.
