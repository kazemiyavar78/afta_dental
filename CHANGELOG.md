# تغییرات

## 2026-11-05

- test(platform): زیرساخت تست مشترک (`internal/platform/testutil`)، تست integration برای AutoMigrate، helper فرانت `renderWithProviders` و اسکریپت `scripts/afta-check.ps1`. منطق برنامه عوض نشده است. وضعیت تست پایگاه: منتظر اجرای کاربر.
- ci: GitHub Actions برای `go vet`، تست واحد Go، `govulncheck`، و تست و build فرانت بدون پایگاه داده. `govulncheck` و `npm audit` تا فاز ۵ blocking نیستند چون `xlsx@0.18.5` آسیب‌پذیری شناخته‌شده دارد و باید در آن فاز جایگزین و سپس blocking شود.
