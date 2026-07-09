# سیستم پذیرش بیماران — بک‌اند (طب پرداز)

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

```bash
go test ./...
```

## استقرار IIS

مستندات مختصر: [docs/iis-deployment.md](docs/iis-deployment.md)

## چک‌لیست افتا

مرجع بررسی: [docs/security-checklist-afta.md](docs/security-checklist-afta.md)
