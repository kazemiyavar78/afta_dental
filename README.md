# سیستم پذیرش بیماران — بک‌اند (طب پرداز)

[![CI](https://github.com/kazemiyavar78/afta_dental/actions/workflows/ci.yml/badge.svg)](https://github.com/kazemiyavar78/afta_dental/actions/workflows/ci.yml)

بک‌اند سیستم پذیرش کلینیک دندان‌پزشکی با رعایت الزامات امنیتی **افتا**.

Cursor را مستقیم روی پوشه `afta_tpdenta` باز کنید تا `.cursor/rules` بار شود.

## استک تکنولوژی

| مؤلفه | انتخاب |
|---|---|
| زبان | Go 1.26.4 (`go.mod`) |
| HTTP Framework | Gin |
| پایگاه داده | SQL Server |
| ORM | `gorm.io/gorm` + `gorm.io/driver/sqlserver` |
| Migration | GORM AutoMigrate هنگام راه‌اندازی سرور |
| هش رمز | bcrypt |
| رمزنگاری داده حساس | AES-256-GCM |
| یکپارچگی | HMAC-SHA256 (کلید جدا از AES) |

> **تصمیم schema:** schema فقط با GORM AutoMigrate هنگام راه‌اندازی ساخته می‌شود؛ فایل‌های SQL قدیمی هرگز اجرا نشده بودند و حذف شدند (در تاریخچه git موجودند). اگر AutoMigrate حتی یک جدول را نسازد، سرور با پیام خطا متوقف می‌شود (FPT_FLS.1).

## ساختار پروژه

```
cmd/server/                 نقطه ورود API
cmd/seed-admin/             ساخت کاربر Admin اولیه
internal/platform/          زیرساخت (config, db, middleware, security)
internal/user/              کاربران، نقش، مجوز
internal/reception/         پذیرش
internal/organization/      سازمان و بیمه‌گر
internal/fund/              صندوق
internal/tariff/            تعرفه
internal/logs/              فهرست رویداد امنیتی
frontend/                   رابط React
docs/                       بافت پروژه، افتا، استقرار IIS
scripts/afta-check.ps1      بررسی واحد، integration، فرانت و SQL خواندنی
```

## پیش‌نیازها

- Go 1.26.4
- Node.js 22.17.0 (فایل `.nvmrc` و فیلد `engines` در `frontend/package.json`)
- SQL Server 2019+

## راه‌اندازی

### ۱. کلون و تنظیم env

فایل `.env` را در ریشه مخزن بسازید. این فایل commit نمی‌شود. در مخزن `.env.example` ریشه وجود ندارد. `frontend/.env.example` فقط نام `VITE_API_BASE_URL` را دارد. مقدار هیچ متغیری را در مخزن ننویسید.

| نام | کاربرد |
|---|---|
| `DATABASE_URL` | نشانی SQL Server برای GORM. الزامی. |
| `AES_KEY` | کلید AES-256-GCM. الزامی. طول ۳۲ بایت. جدا از کلیدهای HMAC. |
| `INTEGRITY_HMAC_KEY` | کلید HMAC صحت ردیف. الزامی. جدا از AES. |
| `AUDIT_HMAC_KEY` | کلید HMAC زنجیره رویداد امنیتی. الزامی. |
| `CSRF_HMAC_KEY` | کلید صدور توکن CSRF. الزامی. |
| `SERVER_HOST` | میزبان bind. اگر نباشد کد `0.0.0.0` می‌گذارد. |
| `SERVER_PORT` | درگاه bind. اگر نباشد کد `8080` می‌گذارد. |
| `APP_ENV` | `development` حالت توسعه (فقط API و Gin debug). مقدار دیگر، از جمله پیش‌فرض کد `production`، پوشه `frontend/dist` را سرو می‌کند. |
| `DEV_MODE` | `true` حالت توسعه را اجبار می‌کند. |
| `SECURE_COOKIES` | پرچم Secure کوکی نشست و CSRF. پیش‌فرض کد روشن است. |
| `ALLOWED_ORIGINS` | مبدأهای CORS، جداشده با ویرگول. |
| `RATE_LIMIT_PER_MINUTE` | سقف درخواست سراسری. پیش‌فرض کد ۱۲۰. |
| `LOGIN_RATE_LIMIT_PER_MINUTE` | در config خوانده می‌شود. محدودیت ورود علاوه بر آن از `loginguard` و `MaximumNumberOfFailedLogin` می‌آید. |
| `LOG_RETENTION_TICKER` | فاصله دورهای بعدی worker پاک‌سازی لاگ و صحت. مدت Go. پیش‌فرض کد `15m`. دور اول بلافاصله بعد از راه‌اندازی اجرا می‌شود. |
| `INTEGRITY_TICKER` | در config خوانده می‌شود. worker فقط `LOG_RETENTION_TICKER` را به کار می‌برد. |
| `SEED_ADMIN_USERNAME` | نام کاربر Admin در `cmd/seed-admin`. اگر نباشد کد نام جایگزین دارد. |
| `SEED_ADMIN_PASSWORD` | رمز همان کاربر. اگر نباشد کد رمز جایگزین دارد. در مخزن و لاگ نیاید. |
| `AFTA_TEST_DATABASE_URL` | نشانی SQL Server برای تست integration و `scripts/afta-check.ps1`. سرور آن را نمی‌خواند. |
| `VITE_API_BASE_URL` | پایه axios در فرانت. |
| `VITE_API_PROXY_TARGET` | مقصد پروکسی `/api` در سرور توسعه Vite. اگر نباشد کد `http://127.0.0.1:8080` می‌گذارد. |

### ۲. ایجاد دیتابیس

```sql
CREATE DATABASE afta_reception;
```

### ۳. ساخت schema

schema فقط با GORM AutoMigrate هنگام راه‌اندازی ساخته می‌شود؛ فایل‌های SQL قدیمی هرگز اجرا نشده بودند و حذف شدند (در تاریخچه git موجودند). دستور migrate جداگانه لازم نیست. اگر ساخت یک جدول خطا بدهد، فرایند سرور همان‌جا متوقف می‌شود.

### ۴. اجرای سرور

```bash
go run ./cmd/server
```

سرور روی پورت `8080` (یا `SERVER_PORT`) اجرا می‌شود.

### ۵. ایجاد کاربر Admin اولیه

```bash
go run ./cmd/seed-admin
```

نام و رمز از `SEED_ADMIN_USERNAME` و `SEED_ADMIN_PASSWORD` می‌آیند. اگر خالی باشند، خود دستور مقدار جایگزین کد را به کار می‌برد. اول یک بار سرور را اجرا کنید تا مجوزها ساخته شوند؛ seed بدون آن‌ها به خطا می‌خورد.

فرانت، از پوشه `frontend`:

```bash
npm install
npm run dev
```

سرور توسعه Vite روی درگاه ۵۱۷۳ است و `/api` را به بک‌اند پروکسی می‌کند. ساخت بسته تولید: `npm run build`.

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

تست integration. فایل‌های دارای `//go:build integration` داخل `go test ./...` نیستند. اگر `AFTA_TEST_DATABASE_URL` خالی باشد، `OpenTestDB` همان تست را `t.Skip` می‌کند. `WithTestTx` بدنه را در یک تراکنش اجرا می‌کند و در پایان، حتی اگر تست شکست بخورد، rollback می‌کند.

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
