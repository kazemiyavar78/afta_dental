# سیستم پذیرش بیماران — بک‌اند (طب پرداز)

[![CI](https://github.com/kazemiyavar78/afta_dental/actions/workflows/ci.yml/badge.svg)](https://github.com/kazemiyavar78/afta_dental/actions/workflows/ci.yml)

بک‌اند سیستم پذیرش کلینیک دندان‌پزشکی با رعایت الزامات امنیتی **افتا**.

## استک تکنولوژی

| مؤلفه | انتخاب |
|---|---|
| زبان | Go 1.22+ |
| HTTP Framework | Gin |
| پایگاه داده | SQL Server |
| ORM | `gorm.io/gorm` + `gorm.io/driver/sqlserver` |
| Migration | `golang-migrate/migrate` (فایل‌های SQL خام) |
| هش رمز | bcrypt |
| رمزنگاری داده حساس | AES-256-GCM |
| یکپارچگی | HMAC-SHA256 (کلید جدا از AES) |

> **تصمیم ORM:** از GORM با درایور sqlserver استفاده شده تا هم خوانایی کد و هم سازگاری با migrationهای SQL خام حفظ شود. AutoMigrate استفاده **نمی‌شود**.

## ساختار پروژه

```
cmd/server/          ← نقطه ورود
internal/
  platform/          ← زیرساخت مشترک (config, db, middleware, security)
  user/              ← ماژول کاربران (کامل)
  reception/         ← نمونه
  organization/      ← نمونه
  fund/              ← نمونه
  tariff/            ← نمونه
migrations/          ← فایل‌های SQL
docs/                ← مستندات استقرار و چک‌لیست افتا
```

## پیش‌نیازها

- Go 1.22+
- Node.js 22.17.0 (فایل `.nvmrc` و فیلد `engines` در `frontend/package.json`)
- SQL Server 2019+
- [golang-migrate CLI](https://github.com/golang-migrate/migrate)

## راه‌اندازی

### ۱. کلون و تنظیم env

```bash
cp .env.example .env
# مقادیر DATABASE_URL و کلیدهای امنیتی را ویرایش کنید
```

**مهم:** هر کلید امنیتی باید جدا و یکتا باشد:
- `AES_KEY` — دقیقاً ۳۲ بایت
- `INTEGRITY_HMAC_KEY`
- `AUDIT_HMAC_KEY`
- `CSRF_HMAC_KEY`

### ۲. ایجاد دیتابیس

```sql
CREATE DATABASE afta_reception;
```

### ۳. اجرای Migration

```bash
migrate -path migrations -database "sqlserver://user:pass@host:1433?database=afta_reception&encrypt=disable" up
```

### ۴. اجرای سرور

```bash
go run ./cmd/server
```

سرور روی پورت `8080` (یا `SERVER_PORT`) اجرا می‌شود.

### ۵. ایجاد کاربر Admin اولیه

```bash
go run ./cmd/seed-admin
# یا با متغیرهای سفارشی:
# SEED_ADMIN_USERNAME=admin SEED_ADMIN_PASSWORD=Admin@1234 go run ./cmd/seed-admin
```

> پس از `SeedDefaults`، تنظیمات امنیتی پیش‌فرض خودکار درج می‌شوند. لیست کامل در بخش [تنظیمات پیش‌فرض](#تنظیمات-پیش‌فرض-securitysettings) آمده است.

## API Endpoints

| Method | Path | نقش | توضیح |
|---|---|---|---|
| POST | `/api/login` | عمومی | ورود |
| POST | `/api/logout` | احراز شده | خروج |
| GET | `/api/me` | احراز شده | اطلاعات کاربر فعلی |
| POST | `/api/users` | Admin | ایجاد کاربر |
| GET | `/api/users` | Admin | لیست کاربران |
| PUT | `/api/users/:id` | Admin/خود | به‌روزرسانی |
| POST | `/api/change-password` | احراز شده | تغییر رمز |
| GET/PUT | `/api/security/settings` | Admin | تنظیمات امنیتی |
| GET/DELETE | `/api/sessions` | احراز شده | مدیریت نشست |
| POST/GET | `/api/reception` | Reception+ | نمونه پذیرش |
| GET/POST | `/api/credit-accounts` | `credit_account.read` / `credit_account.create` | لیست و ایجاد حساب اعتباری |
| GET | `/api/credit-accounts/active` | `credit_charge.create` | حساب‌های اعتباری فعال برای پرداخت |
| GET/PUT/DELETE | `/api/credit-accounts/:id` | `credit_account.read/update/delete` | مشاهده، ویرایش و حذف حساب اعتباری |
| POST | `/api/wallet/credit-charge` | `credit_charge.create` | شارژ پرونده مالی از محل کسر اعتبار |
| PATCH | `/api/wallet/transactions/:id/credit-paid` | `credit_charge.update` | تیک پرداخت‌شدن تراکنش کسر از اعتبار |
| GET | `/api/reports/credit-account-payments` | `reports.read` | گزارش پرداخت‌های حساب اعتباری |

## حساب‌های اعتباری

حساب اعتباری روش پرداخت کیف پول است. فیلدها: نام اعتبار، توضیحات، سقف اعتبار (ریال) و وضعیت فعال/غیرفعال.

| مجوز | کاربرد |
|---|---|
| `credit_account.read` | دیدن حساب‌ها |
| `credit_account.create` | ایجاد حساب |
| `credit_account.update` | ویرایش حساب |
| `credit_account.delete` | حذف حساب بدون پرداخت |
| `credit_charge.create` | شارژ پرونده با فرم انتخاب حساب، مبلغ و توضیحات |
| `credit_charge.update` | تیک زدن پرداخت‌شدن تراکنش کسر از اعتبار |

سقف اعتبار با هر شارژ کم می‌شود. حساب غیرفعال در پرداخت جدید انتخاب نمی‌شود. حذف فقط وقتی مجاز است که هنوز پرداختی ثبت نشده باشد. ویرایش سقف کمتر از مبلغ استفاده‌شده رد می‌شود.

تراکنش کسر از اعتبار در ابتدا پرداخت‌نشده است. در پرونده مالی و در گزارش، با مجوز `credit_charge.update` می‌توان تیک «پرداخت شده» را زد یا برداشت. این تیک سقف اعتبار و موجودی کیف پول را تغییر نمی‌دهد.

گزارش `GET /api/reports/credit-account-payments` پرداخت‌ها را با فیلتر از تاریخ، تا تاریخ و نام حساب برمی‌گرداند و وضعیت پرداخت‌شدن را هم نشان می‌دهد.

## CSRF

الگوی **Double-Submit Cookie**:
1. پس از لاگین، کوکی `csrf_token` (غیر HttpOnly) ست می‌شود
2. فرانت باید مقدار را در هدر `X-CSRF-Token` برای POST/PUT/PATCH/DELETE بفرستد
3. GET/HEAD/OPTIONS نیاز به CSRF ندارند

## نقش‌ها (RBAC)

فاز اول: یک `RoleID` مستقیم روی کاربر. نقش‌های پیش‌فرض: `Admin`, `Doctor`, `Reception`.

> در آینده قابل گسترش به چندنقشی (`UserRoles`) بدون تغییر معماری امنیتی.

## تنظیمات پیش‌فرض SecuritySettings

| کلید | مقدار پیشنهادی |
|---|---|
| MaximumNumberOfSessionsPerUser | 3 |
| MaximumNumberOfFailedLogin | 5 |
| MaximumTimeOfUserIpBeingLocked | 30 (دقیقه) |
| MaximumSizeOfLogTables | 100 (MB) |
| WorkHours_Start | 08:00 |
| WorkHours_End | 18:00 |
| IsSundayBlocked | true |
| IsFridayBlocked | true |
| Password_MinLength | 8 |
| Password_RequireUppercase | true |
| Password_RequireLowercase | true |
| Password_RequireDigit | true |
| Password_RequireSpecial | true |
| MaximumUploadsPerWindow | 10 |
| UploadWindowMinutes | 60 |

## تست

تست واحد (بدون SQL Server):

```bash
go test ./...
```

تست integration (اگر `AFTA_TEST_DATABASE_URL` خالی باشد رد می‌شود):

```bash
go test -tags integration ./...
```

بررسی کامل (vet، هر دو نوع تست Go، تست و build فرانت، audit وابستگی‌ها، SQL خواندنی):

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\afta-check.ps1
```

CI روی push و pull request به `main` همین بررسی‌های بدون پایگاه داده را در [`.github/workflows/ci.yml`](.github/workflows/ci.yml) اجرا می‌کند: بک‌اند روی `windows-latest` (`go vet`، `go test ./... -count=1` بدون tag `integration`، `govulncheck`) و فرانت روی `ubuntu-latest` (`npm ci`، `npm test`، `npm run build`، `npm audit --omit=dev --audit-level=high`). هیچ secret و متغیر پایگاه داده در workflow نیست. تا فاز ۵، `govulncheck` و `npm audit` به‌خاطر آسیب‌پذیری شناخته‌شدهٔ `xlsx@0.18.5` با `continue-on-error` اجرا می‌شوند و در آن فاز باید blocking شوند.

## استقرار IIS

مستندات مختصر: [docs/iis-deployment.md](docs/iis-deployment.md)

## چک‌لیست افتا

مرجع بررسی: [docs/security-checklist-afta.md](docs/security-checklist-afta.md)
