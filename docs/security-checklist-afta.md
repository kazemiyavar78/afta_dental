# چک‌لیست امنیتی افتا — مرجع بررسی

این سند هر تسک سند افتا را به کد واقعی پروژه لینک می‌کند.

---

## تسک ۱ — SecuritySettings

| مورد | مسیر |
|---|---|
| جدول Key-Value | `migrations/0001_security_settings.up.sql` |
| مدل | `internal/platform/security/settings/model.go` |
| سرویس GetSettingValue / UpdateSetting | `internal/platform/security/settings/service.go` |
| ثابت‌های کلید | `internal/platform/security/settings/keys.go` |
| IntegrityHash روی Update | `settings/service.go` → `UpdateSetting` |
| ثبت رویداد امنیتی | `settings/service.go` → `audit.EventSecuritySettingChange` |
| Endpoint Admin | `internal/user/handler.go` → `UpdateSecuritySetting` |
| تست Smoke | `internal/platform/security/settings/service_test.go` |

---

## تسک ۲ — Sessions

| مورد | مسیر |
|---|---|
| جدول Sessions | `migrations/0002_sessions.up.sql` |
| مدل و Repository | `internal/platform/security/session/model.go` |
| سرویس ایجاد/حذف | `internal/platform/security/session/service.go` |
| کوکی HttpOnly+Secure+SameSite | `internal/platform/middleware/session.go` |
| SessionMiddleware | `internal/platform/middleware/session.go` |
| حذف فیزیکی | `session/model.go` → `Delete` (DELETE نه soft) |
| Endpoint مدیریت | `internal/user/handler.go` → `ListSessions`, `DeleteSession` |
| تست Smoke | `internal/platform/security/session/service_test.go` |

---

## تسک ۳ — SecurityEvents + Hash-Chain

| مورد | مسیر |
|---|---|
| جدول | `migrations/0003_security_events.up.sql` |
| مدل (PrevHash, RowHash, CreatedAt) | `internal/platform/security/audit/model.go` |
| LogEvent با Hash-Chain | `internal/platform/security/audit/manager.go` |
| VerifyChain | `internal/platform/security/audit/manager.go` |
| Checkpoint | `settings/keys.go` → `LastVerifiedEventId` |
| انواع رویداد (const) | `internal/platform/security/audit/types.go` |
| تست Smoke | `internal/platform/security/audit/manager_test.go` |

---

## تسک ۴ — رمزنگاری AES-256

| مورد | مسیر |
|---|---|
| AES-256-GCM | `internal/platform/security/encryption/encryption.go` |
| UserEncryptionService | `internal/platform/security/encryption/user_encryption.go` |
| SecurityCode روی User | `internal/user/service.go` → `CreateUser`, `UpdateUser` |
| SecurityCheckMiddleware | `internal/platform/middleware/security_check.go` |
| تست Smoke | `internal/platform/security/encryption/encryption_test.go` |

---

## تسک ۵ — محدودیت آپلود

| مورد | مسیر |
|---|---|
| جدول UploadCounters | `migrations/0006_upload_counters_and_samples.up.sql` |
| MERGE اتمیک | `internal/platform/security/uploadguard/guard.go` |
| کلیدهای تنظیم | `settings/keys.go` → `MaximumUploadsPerWindow`, `UploadWindowMinutes` |

---

## تسک ۶ — پیچیدگی رمز عبور

| مورد | مسیر |
|---|---|
| کلیدهای Password_* | `settings/keys.go` |
| PasswordValidator | `internal/platform/security/passwordpolicy/validator.go` |
| استفاده در CreateUser/ChangePassword | `internal/user/service.go` |
| تست Smoke | `internal/platform/security/passwordpolicy/validator_test.go` |

---

## تسک ۷ — محدودیت نشست هم‌زمان

| مورد | مسیر |
|---|---|
| کلید MaximumNumberOfSessionsPerUser | `settings/keys.go` |
| حذف قدیمی‌ترین | `session/service.go` → `CreateSession` |

---

## تسک ۸ — ساعات کاری

| مورد | مسیر |
|---|---|
| کلیدهای Is{Day}Blocked + WorkHours | `settings/keys.go` |
| بررسی در لاگین | `loginguard/guard.go` → `CheckWorkHours` |
| Admin مستثنی | `loginguard/guard.go` → `roleName == "Admin"` |

---

## تسک ۹ — محدودیت تلاش ناموفق (دیتابیس)

| مورد | مسیر |
|---|---|
| جدول LoginAttempts | `migrations/0005_login_attempts.up.sql` |
| MERGE اتمیک | `loginguard/guard.go` → `RecordFailedAttempt` |
| Reset پس از ورود موفق | `user/service.go` → `HandleUserLogin` |
| LoginRateLimiterMiddleware | `internal/platform/middleware/ratelimit.go` |
| تست Smoke | `internal/platform/security/loginguard/guard_test.go` |

---

## تسک ۱۰ — مدیریت حجم لاگ

| مورد | مسیر |
|---|---|
| Background Worker | `internal/platform/security/logretention/worker.go` |
| sp_spaceused معادل | `audit/model.go` → `GetTableSizeMB` |
| حذف ۱۰٪ قدیمی‌ترین | `logretention/worker.go` → `cleanupSecurityEvents` |
| LogCleanupSummaries | `migrations/0006_upload_counters_and_samples.up.sql` |
| Verify Integrity دوره‌ای | `logretention/worker.go` → `verifyIntegrity` |

---

## موارد اضافه (توافق چت)

### Integrity Hash (HMAC)

| مورد | مسیر |
|---|---|
| پکیج Sign/Verify | `internal/platform/security/integrity/integrity.go` |
| کلید جدا از AES | `config/config.go` → `INTEGRITY_HMAC_KEY` |
| Users | `internal/user/repo.go` → `SignUserIntegrityHash` |
| Roles | `internal/user/repo.go` → `SignRoleIntegrityHash` |
| SecuritySettings | `settings/service.go` |
| Verify در خواندن + Worker | `settings/service.go`, `logretention/worker.go` |
| تست Smoke | `integrity/integrity_test.go` |

### CSRF Double-Submit

| مورد | مسیر |
|---|---|
| CSRFMiddleware | `internal/platform/middleware/csrf.go` |
| فقط mutating | `csrf.go` → GET/HEAD/OPTIONS رد می‌شوند |
| رویداد نقض | `audit.EventCSRFViolation` |

### ساختار ماژولار

| مورد | مسیر |
|---|---|
| handler → service → repo | `internal/user/`, `internal/reception/` |
| نمونه ماژول‌ها | `internal/reception/`, `organization/`, `fund/`, `tariff/` |
| platform مشترک | `internal/platform/` |

---

## چک‌لیست پذیرش نهایی

- [x] هیچ رمز/کلید Hardcode نشده — `config/config.go`, `.env.example`
- [x] Password به‌صورت Hash (bcrypt) — `user/service.go`
- [x] حذف فیزیکی نشست و لاگ — `session/model.go`, `audit/model.go`
- [x] رویدادهای امنیتی Trigger شده — `audit/types.go` + فراخوانی در serviceها
- [x] CSRF فقط روی mutating — `middleware/csrf.go`
- [x] Integrity Hash روی جداول ذکرشده
- [x] Hash-Chain با VerifyChain
- [x] ماژول پذیرش از همان میدلورها
- [x] کامنت‌ها فارسی
