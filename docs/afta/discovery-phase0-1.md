# کشف فاز ۰ و ۱

تاریخ بررسی کد: ۲۰۲۶-۱۱-۰۵. فقط خواندن کد و اجرای تست بدون پایگاه‌داده. هیچ کوئری به SQL Server زده نشده است. هر جا مقدار زنده لازم است، نوشته شده «منتظر خروجی اسکریپت».

ریشهٔ مخزن: `afta_tpdenta`. ماژول: `github.com/tpdenta/afta-reception`.

---

## الف) تست و CI

### ۱. فهرست تست‌ها

همهٔ تست‌های Go در حافظه هستند و پایگاه‌داده باز نمی‌کنند. ۱۶ تابع `Test`.

| فایل | تعداد | چه چیزی را تست می‌کند |
|---|---|---|
| `internal/platform/security/audit/manager_test.go` | ۱ | `TestComputeRowHash`: یک HMAC دستی ساخته می‌شود و طول ۶۴ نویسهٔ هگز چک می‌شود. تابع `Manager.computeRowHash` صدا زده نمی‌شود. |
| `internal/platform/security/encryption/encryption_test.go` | ۱ | رمز و رمزگشایی AES. |
| `internal/platform/security/integrity/integrity_test.go` | ۱ | امضا و تأیید HMAC. |
| `internal/platform/security/loginguard/guard_test.go` | ۲ | تجزیهٔ ساعت روز؛ فیلدهای مدل `LoginAttempt`. |
| `internal/platform/security/passwordpolicy/validator_test.go` | ۲ | وجود حرف بزرگ؛ پیش‌فرض حداقل طول ۸. |
| `internal/platform/security/session/service_test.go` | ۱ | فیلدهای مدل نشست. |
| `internal/platform/security/session/mssql_uuid_test.go` | ۴ | جابه‌جایی endian و Scan/Value برای UUID مخصوص SQL Server. |
| `internal/platform/security/settings/model_test.go` | ۱ | `FindAll(true)` کلیدهای خارج از دیکشنری فارسی را حذف می‌کند. |
| `internal/platform/security/settings/service_test.go` | ۱ | امضای تنظیمات. |
| `internal/user/service_test.go` | ۳ | bcrypt؛ فیلدهای HMAC کاربر؛ پاک شدن association کهنهٔ نقش در به‌روزرسانی. |

فرانت (Vitest)، ۱۳ تست:

| فایل | تعداد | موضوع |
|---|---|---|
| `frontend/src/platform/reports/utils/html.test.ts` | ۳ | escapeHtml، CSS صفحه، ادغام پیش‌فرض صفحه |
| `frontend/src/platform/reports/utils/filename.test.ts` | ۳ | پسوند نام فایل، کوتاه‌کردن و حذف نویسهٔ نامعتبر شیت |
| `frontend/src/platform/reports/formatters/value.test.ts` | ۳ | قالب سلول، خط تیره برای خالی، خواندن فیلد ردیف |
| `frontend/src/platform/reports/formatters/number.test.ts` | ۴ | عدد فارسی، پول، درصد، null/NaN |

ماژول‌های پذیرش، بیمار، کیف پول، گزارش بک‌اند و هندلرها فایل تست ندارند.

### ۲. وابستگی به DB و نتیجهٔ اجرا

هیچ‌کدام از ۱۶ تست Go و ۱۳ تست فرانت به SQL Server وصل نمی‌شوند. helper دیتابیس، `httptest` و `sqlmock` در مخزن نیست.

اجرای همین مرحله (بدون DB):

- `go test ./... -count=1`: همهٔ بسته‌های دارای تست پاس شدند. شکست: ۰.
- `npm test` در `frontend` (`vitest run`): ۴ فایل، ۱۳ تست، همه پاس. شکست: ۰.
- `go vet ./...`: خروج ۰. برای کش ماژول از `GOMODCACHE` کاربر و `GOPROXY=off` استفاده شد چون پروکسی در محیط ایزوله ۴۰۳ می‌داد.

`go test ./...` و `npm test` داخل `scripts/afta-discovery.ps1` بخش `=== Q2 ===` هم هست. منتظر خروجی اسکریپت برای تکرار روی ماشین کاربر.

### ۳. helper مشترک، دیتابیس تستی، mock

وجود ندارد. هر فایل تست داده را داخل همان فایل می‌سازد. هیچ `testdb`، `testify`، `sqlmock` یا سرور HTTP آزمایشی در کد تست دیده نشد.

### ۴. CI

`.github/workflows` وجود ندارد. `gitlab-ci.yml` و `Jenkinsfile` هم در ریشهٔ مخزن نیست. CI دیگری در مخزن پیدا نشد.

### ۵. قوانین Cursor

`.cursor/rules` و `.cursorrules` وجود ندارد. قوانین در `cursor/rules/` هستند:

- `afta.mdc` و `afta-core.mdc` تقریباً تکراری‌اند. هر دو محدودهٔ ماژول، الگوی handler/service/repo، و ارجاع به `docs/afta/traceability.md` را می‌گویند. اختلاف در بخش ۳ است. فایل `docs/afta/traceability.md` هنوز وجود ندارد.
- `docs-sync.mdc` می‌گوید پس از تغییر رفتار، سند زمینهٔ پروژه هم‌خوان شود.

---

## ب) پایگاه داده و migration

### ۶. AutoMigrate در برابر `migrations/0001` تا `0006`

سرور فقط `migrate.Migrate` را صدا می‌زند (`cmd/server/main.go` خط ۸۶ تا ۱۱۶). آن تابع هر مدل را با `db.AutoMigrate` می‌سازد (`internal/platform/migrate/migrate.go` خط ۱۱ تا ۱۴) و خطای یک جدول را نادیده می‌گیرد و ادامه می‌دهد. فایل‌های SQL از Go فراخوانی نمی‌شوند. `golang-migrate` وابستگی `go.mod` نیست. README هنوز CLI آن را مستند کرده است.

فهرست AutoMigrate: User، OrganizationPackage، OrganizationPackageRelation، Organization، Patient، ServiceItem، ExcludedService، Fund، Tariff، Reception، ReceptionService، ReceptionPhoto، SpecialCode، Regulation، RegulationService، CashFund، PatientWallet، BankAccount، CreditAccount، Transaction، SecuritySetting، Session، SecurityEvent، Role، Permission، RolePermission، LoginAttempt.

`LogCleanupSummary` و `UploadCounter` در این فهرست نیستند.

SQL فقط این جدول‌ها را می‌سازد: SecuritySettings (0001)، Sessions (0002)، SecurityEvents (0003)، Roles / Permissions / RolePermissions / Users (0004)، LoginAttempts (0005)، UploadCounters و نمونه‌های Receptions / Organizations / Funds / Tariffs و LogCleanupSummaries (0006).

تفاوت‌های قطعی از روی کد و SQL:

| جدول | تفاوت |
|---|---|
| Users | SQL ستون `UserType` ندارد (`migrations/0004_users.up.sql` خط ۳۹ تا ۶۰). مدل Go ستون `UserType` با `not null;default:User` دارد (`internal/user/model.go` خط ۲۹). بقیهٔ ستون‌های امنیتی (از جمله PhoneNumber و TwoFactor*) در هر دو هستند. |
| Receptions | SQL چهار ستون نمونه است: PatientName، DoctorID اجباری، ReceptionDate، Status (`0006` خط ۱۶ تا ۲۴). مدل Go `gorm.Model` به‌علاوه PatientID، بیمه‌ها، DoctorID تهی‌پذیر و ده‌ها ستون دیگر است (`internal/reception/model.go` خط ۳۰ به بعد). |
| Organizations | SQL فقط ID، Name، CreatedAt. مدل Go `gorm.Model` به‌علاوه IsTakmili، IsActive، IsFree، IntegrityHash، PackageID، CenterPackageID است (`internal/organization/model.go` خط ۱۰ به بعد). |
| Funds | نزدیک است: ID، Name، Balance، CreatedAt در هر دو (`internal/fund/model.go` خط ۱۰ تا ۱۵ و SQL خط ۴۰). |
| Tariffs | SQL فقط Name و Amount. مدل Go OrganizationID، ServiceID، Amount، TariffAmount، OrganizationShare و فیلدهای دیگر روی `gorm.Model` است (`internal/tariff/model.go` خط ۸ به بعد). |
| LogCleanupSummaries | فقط SQL (`0006` خط ۶۲ تا ۶۹). مدل Go ستون `TableName` را با فیلد `TargetTable` نگاشت می‌کند (`internal/platform/security/logretention/worker.go` خط ۱۸ تا ۲۷) ولی AutoMigrate نمی‌شود. |
| UploadCounters | فقط SQL. مدل در `uploadguard` هست و AutoMigrate نمی‌شود. |
| بقیهٔ جدول‌های دامنه | patients، Services، ReceptionServices، کیف پول، ضابطه، کد خاص، بستهٔ تعرفه فقط از AutoMigrate می‌آیند. در SQL نیستند. |

`gorm.Model` بدون تگ ستون، با NamingStrategy پیش‌فرض GORM به `id`، `created_at`، `updated_at`، `deleted_at` (snake_case) تبدیل می‌شود. ستون‌های دارای `column:` همان نام PascalCase می‌مانند. اگر پایگاه با SQL ساخته شده باشد و بعد AutoMigrate روی همان جدول نمونه اجرا شود، شکل نهایی نامشخص است تا اسکریپت ستون‌ها را بخواند.

وجود `schema_migrations` و فهرست واقعی جدول/ستون: منتظر خروجی اسکریپت (`=== Q6 ===`).

### ۷. LogCleanupSummaries

از روی کد: SQL `0006` جدول را می‌سازد. سرور آن را AutoMigrate نمی‌کند. worker پس از حذف، `db.Create` می‌زند و خطا را دور می‌ریزد (`worker.go` خط ۱۲۱ تا ۱۲۷). اگر جدول نباشد، پاکسازی لاگ همچنان رویداد `LOG_CLEANUP` را ثبت می‌کند و درج خلاصه بی‌صدا شکست می‌خورد.

وجود جدول در پایگاه و ۲۰ خط آخر لاگ worker: منتظر خروجی اسکریپت (`=== Q7 ===`). در مخزن فایل لاگ سرور نیست؛ سرور با `log.Printf` روی stdout می‌نویسد. اسکریپت فقط اگر `AFTA_SERVER_LOG` یا یکی از نام‌های `server.log` / `app.log` / `afta.log` / `backend.log` وجود داشته باشد، دم فایل را می‌خواند.

---

## ج) لاگ ممیزی

### ۸. `SecurityEvent` و هش

`internal/platform/security/audit/model.go` خط ۱۰ تا ۱۹:

| فیلد | نوع | ستون |
|---|---|---|
| ID | int64 | ID |
| UserID | *int | UserID |
| Ip | string | Ip |
| EventType | string | EventType |
| Description | string | Description |
| CreatedAt | time.Time | CreatedAt |
| PrevHash | string | PrevHash |
| RowHash | string | RowHash |

`PrevHash` برابر `RowHash` آخرین ردیف موجود است. اگر ردیفی نباشد، رشتهٔ خالی است (`manager.go` خط ۵۰ تا ۶۵، `GetLastRowHash`).

`RowHash` برابر HMAC-SHA256 با کلید ممیزی است. متن، با جداکنندهٔ `|`:

`prevHash | userID | ip | eventType | description | createdAt.UTC().RFC3339Nano`

اگر `userID == nil`، بخش کاربر رشتهٔ خالی است (`computeRowHash` خط ۳۰ تا ۴۶). زمان داخل `LogEvent` با `time.Now().UTC()` ساخته می‌شود، نه زمان کلاینت.

### ۹. همهٔ `LogEvent`

فیلد کاربر همیشه `*int` است: شناسهٔ عددی کنشگر، یا شناسهٔ خود کاربر در ورود. نام کاربری در این ستون ذخیره نمی‌شود. `nil` در SQL برابر NULL است. توضیح، متن وضعیت جدید است؛ مقدار قبلی ذخیره نمی‌شود. خطای `LogEvent` همه‌جا با `_ =` نادیده گرفته می‌شود.

ثابت‌های `ROLE_CHANGE` و `DOWNLOAD` در `audit/types.go` خط ۱۰ و ۱۲ تعریف شده‌اند و به `LogEvent` داده نمی‌شوند.

| فایل:خط | نوع | کاربر | توضیح |
|---|---|---|---|
| user/service.go:107 | USER_DATA_CHANGE | شناسهٔ کنشگر | ایجاد کاربر + نام کاربری |
| user/service.go:122 | LOGIN_FAILED | nil | کاربر یافت نشد + نام کاربری واردشده |
| user/service.go:131 | DATA_TAMPERING | شناسهٔ همان کاربر | دستکاری دادهٔ کاربر |
| user/service.go:141 | LOGIN_FAILED | شناسهٔ کاربر | حساب قفل یا غیرفعال |
| user/service.go:155 | LOGIN_FAILED | شناسهٔ کاربر | خارج از ساعات کاری |
| user/service.go:162 | LOGIN_FAILED | شناسهٔ کاربر | رمز اشتباه |
| user/service.go:168 | DATA_TAMPERING | شناسهٔ کاربر | SecurityCode نامعتبر هنگام ورود |
| user/service.go:184 | LOGIN_SUCCESS | شناسهٔ کاربر | ورود موفق |
| user/service.go:192 | LOGOUT | شناسهٔ کاربر | خروج کاربر ID=… |
| user/service.go:216 | DATA_TAMPERING | شناسهٔ ادمین خواننده | خواندن کاربر با یکپارچگی نقض‌شده؛ IP خالی |
| user/service.go:333 | USER_DATA_CHANGE | کنشگر | به‌روزرسانی کاربر + نام کاربری |
| user/service.go:376 | USER_DATA_CHANGE | خود کاربر | «تغییر رمز عبور» |
| user/service.go:423 | DATA_TAMPERING | شناسهٔ کاربر بررسی‌شده | دستکاری کاربر |
| user/service.go:443 | DATA_TAMPERING | nil | دستکاری نقش |
| user/service.go:732 | USER_DATA_CHANGE | کنشگر | ایجاد نقش + نام |
| user/service.go:751 | DATA_TAMPERING | کنشگر | نقض یکپارچگی نقش |
| user/service.go:789 | USER_DATA_CHANGE | کنشگر | بروزرسانی نقش + نام |
| user/service.go:810 | DATA_TAMPERING | کنشگر | نقض یکپارچگی نقش |
| user/service.go:829 | USER_DATA_CHANGE | کنشگر | حذف نقش + نام |
| user/service.go:860 | DATA_TAMPERING | کنشگر | انتصاب نقش دستکاری‌شده رد شد |
| user/service.go:874 | DATA_TAMPERING | شناسهٔ کاربر | دستکاری نقش؛ ورود/دسترسی رد شد |
| wallet/service.go:119,214,266,346,385,531,557 | USER_DATA_CHANGE | کنشگر | تراکنش نقدی، کارت‌به‌کارت، کارتخوان، اعتبار، وضعیت پرداخت، بدهی خدمات |
| wallet/service.go:629,666,697,773,810,835 | USER_DATA_CHANGE | کنشگر | ایجاد/ویرایش/حذف حساب بانکی و اعتباری |
| organization/service.go:84 | DATA_TAMPERING | کنشگر | نقض یکپارچگی سازمان |
| organization/service.go:188,260,281 | USER_DATA_CHANGE | کنشگر | ایجاد/بروزرسانی/حذف سازمان |
| reception/service.go:209,314,420,440,1285 | USER_DATA_CHANGE | کنشگر | ایجاد، ویرایش، حذف نرم، بازیابی، پایان پذیرش |
| reception/service.go:1333 | UPLOAD | کنشگر | آپلود عکس پذیرش |
| patient/service.go:115 | DATA_TAMPERING | کنشگر | نقض یکپارچگی بیمار (نام و نام خانوادگی در توضیح) |
| patient/service.go:164,307,329 | USER_DATA_CHANGE | کنشگر | ایجاد/بروزرسانی/حذف بیمار |
| services/service.go:91,332,374 | DATA_TAMPERING | کنشگر | نقض یکپارچگی خدمت یا خدمت خارج‌شده |
| services/service.go:142,237,258,351,383 | USER_DATA_CHANGE | کنشگر | ایجاد/بروزرسانی/حذف خدمت و افزودن/حذف خارج‌شده |
| tariff/service.go:216,312 | USER_DATA_CHANGE | کنشگر | ذخیره و بازمحاسبهٔ تعرفه |
| session/service.go:106 | SESSION_REVOKED | کنشگر | حذف نشست + UUID نشست |
| regulation/service.go:74,146,162 | USER_DATA_CHANGE | کنشگر | ایجاد/بروزرسانی/حذف ضابطه |
| specialcode/service.go:40 | DATA_TAMPERING | کنشگر | نقض یکپارچگی کد خاص |
| specialcode/service.go:74,171,187 | USER_DATA_CHANGE | کنشگر | ایجاد/بروزرسانی/حذف کد خاص |
| fund/service.go:27 | USER_DATA_CHANGE | کنشگر | ایجاد صندوق. ویرایش/حذف صندوق رویداد ندارد |
| organizationPackage/service.go:79 | DATA_TAMPERING | کنشگر | نقض یکپارچگی بسته |
| organizationPackage/service.go:116,174,195 | USER_DATA_CHANGE | کنشگر | ایجاد/بروزرسانی/حذف بسته |
| settings/service.go:45 | DATA_TAMPERING | nil و IP خالی | دستکاری تنظیم هنگام خواندن |
| settings/service.go:83 و 105 | SECURITY_SETTING_CHANGE | کنشگر؛ برای worker عدد ۰ | تغییر تنظیم + نام کلید + مقدار جدید |
| settings/service.go:143 | DATA_TAMPERING | nil | دستکاری تنظیم در بررسی دوره‌ای |
| logretention/worker.go:129 | LOG_CLEANUP | nil و IP برابر `system` | تعداد حذف و حجم |
| logretention/worker.go:169 | DATA_TAMPERING | nil و IP برابر `system` | شکست Hash-Chain از checkpoint |
| middleware/csrf.go:42 و 55 | CSRF_VIOLATION | اشاره‌گر به نتیجهٔ type-assert؛ اگر کاربر در context نباشد مقدار ۰ | توکن ناقص، یا توکن نامطابق با نشست |
| middleware/security_check.go:48 | DATA_TAMPERING | شناسهٔ کاربر جاری | SecurityCode نامعتبر و پایان نشست |

تعداد رویداد به تفکیک نوع، و تعداد UserID خالی یا فقط عددی: منتظر خروجی اسکریپت (`=== Q9 ===` و `=== Q10 ===`). اسکریپت متن رکورد را چاپ نمی‌کند.

### ۱۰. چرا شناسهٔ کاربر خالی یا عدد است

ستون `UserID` نام کاربری نیست. `nil` یعنی NULL و در UI خالی دیده می‌شود. عدد، شناسهٔ جدول Users است. شناسهٔ ۱ در کد هاردکد نشده؛ اگر اولین کاربر ادمین باشد، کنش‌های او ۱ است. مسیرهایی که کنشگر را ثبت نمی‌کنند:

- ورود با نام کاربری ناموجود: `nil` (`user/service.go:122`).
- دستکاری نقش در اسکن دوره‌ای: `nil` (`user/service.go:443`).
- خواندن یا اسکن تنظیم خراب: `nil` (`settings/service.go:45` و `:143`).
- worker پاکسازی و شکست زنجیره: `nil` (`worker.go:129` و `:169`).
- CSRF وقتی `ContextKeyUserID` عدد نیست: `uid, _ := userID.(int)` صفر می‌ماند و `&uid` ذخیره می‌شود (`csrf.go:41` و `:54`). این NULL نیست؛ عدد ۰ است.
- worker هنگام نوشتن checkpoint تنظیم، `UpdateSetting(..., 0, "system")` می‌زند (`worker.go:176`) و رویداد با UserID برابر ۰ ثبت می‌شود (`settings/service.go:104`).

### ۱۱. فیلتر IP در `GET /api/logs`

هندلر فقط کوئری را به ریپازیتوری می‌دهد (`internal/logs/handler.go` خط ۱۶ تا ۳۳). فیلتر این است:

```130:132:internal/platform/security/audit/model.go
	if filter.IP != "" {
		q = q.Where("SecurityEvents.Ip = ?", filter.IP)
	}
```

مقایسه برابری دقیق است. `LIKE`، نرمال‌سازی IPv4/IPv6، حذف فاصله، و تطبیق پیشوند وجود ندارد. یک نویسهٔ اضافه نتیجه را خالی می‌کند.

دو رفتار هم‌جوار هم جستجو را خالی نشان می‌دهد:

- فیلتر نام کاربری `INNER JOIN Users` است (`model.go` خط ۱۳۳ تا ۱۳۵). ردیف‌هایی که `UserID` آن‌ها NULL است از نتیجه حذف می‌شوند.
- صفحهٔ لاگ از DatePicker آنت‌دیزاین استفاده می‌کند، نه `JalaliDatePicker`. `App` داخل `ConfigProvider` با locale فارسی است و `JalaliProvider` (antd-jalali-v5) روی locale فارسی `dayjs.calendar('jalali')` سراسری می‌گذارد. `.format('YYYY-MM-DD')` می‌تواند سال جلالی بفرستد و سرور آن را میلادی پارس کند؛ بازه با تاریخ ذخیره‌شدهٔ UTC هم‌پوشانی ندارد.

### ۱۲. مرتب‌سازی، جستجو، صفحه‌بندی

فیلتر IP، نام کاربری و تاریخ سمت سرور است، ولی فقط همان برابری‌های بند ۱۱. مرتب‌سازی سرور ثابت است: `ORDER BY SecurityEvents.ID DESC` (`model.go` خط ۱۴۵). پارامتر صفحه، اندازهٔ صفحه، یا ستون مرتب‌سازی در `FindEvents` نیست. `DataTable` صفحه‌بندی را روی آرایهٔ کامل سمت کلاینت انجام می‌دهد (پیش‌فرض ۱۰ ردیف). ستون‌های صفحهٔ لاگ sorter سرور ندارند.

### ۱۳. حذف نشست و خروج

| مسیر | رویداد |
|---|---|
| `POST` خروج: `handler.go` خط ۵۶ تا ۶۷ اول `DeleteSession` سپس `HandleLogout` | `SESSION_REVOKED` با UUID نشست (`session/service.go:106`) و بعد `LOGOUT` (`user/service.go:192`). اگر شناسهٔ نشست در context نباشد، فقط LOGOUT. |
| حذف نشست توسط کاربر یا مدیر: `handler.go` خط ۴۲۲ و `DeleteSession` | فقط `SESSION_REVOKED`. |
| رسیدن به سقف نشست: `CreateSession` خط ۴۴ تا ۴۸ `DeleteOldestByUserID` | هیچ رویدادی. |
| `DeleteAllUserSessions` (`session/service.go:121`) | هیچ رویدادی. |
| شکست SecurityCode در میدل‌ور (`security_check.go:45` تا ۴۹) | حذف همهٔ نشست‌ها، کوکی با `secure=true` ثابت، `DATA_TAMPERING`. `LOGOUT` ثبت نمی‌شود. |
| شکست SecurityCode داخل خود ورود (`user/service.go:167`) | فقط `DATA_TAMPERING`. نشستی ساخته نشده که حذف شود. |
| مهلت بیکاری، تغییر IP یا مرورگر، تغییر نقش یا تنظیمات | چنین ابطالی در کد نیست. |

### ۱۴. قفل پس از ورود ناموفق

قفل روی IP مشتری در جدول `LoginAttempts` است (`loginguard/guard.go` خط ۱۳ تا ۲۳)، نه روی `User.IsLocked`.

`RecordFailedLogin` آستانه را از `MaximumNumberOfFailedLogin` (پیش‌فرض ۵) و مدت را از `MaximumTimeOfUserIpBeingLocked` (پیش‌فرض ۳۰ دقیقه) می‌خواند. MERGE وقتی شمارنده به آستانه برسد `LockedUntil` را ست می‌کند (خط ۴۸ تا ۶۲). خود این تابع رویداد ممیزی نمی‌نویسد. رویداد `LOGIN_FAILED` را مسیر ورود می‌نویسد، فقط برای «کاربر نیست» و «رمز اشتباه» (`user/service.go:122` و `:162`). حساب `IsLocked` یا غیرفعال، `LOGIN_FAILED` می‌نویسد ولی `RecordFailedLogin` را صدا نمی‌زند (خط ۱۴۰ تا ۱۴۳).

درخواست بعدی همان IP، پیش از هندلر، در `LoginRateLimiterMiddleware` رد می‌شود (`middleware/ratelimit.go` خط ۵۲ تا ۶۰) و دوباره داخل `HandleUserLogin` (`service.go:115`). هر دو `ErrIPLocked` برمی‌گردانند. شمارنده زیاد نمی‌شود و رویداد تازه‌ای ثبت نمی‌شود. این میدل‌ور روی کل گروه `/api` است، نه فقط `/api/login`؛ IP قفل‌شده به بقیهٔ API هم نمی‌رسد.

### ۱۵. نقش، مجوز، کاربر، تنظیمات

- ایجاد نقش، بروزرسانی نقش (شامل تعویض مجوزها) و حذف نقش: `USER_DATA_CHANGE` با نام نقش (`user/service.go:732`، `:789`، `:829`). فهرست مجوزهای قبل و بعد در توضیح نیست.
- ویرایش کاربر: یک سطر `USER_DATA_CHANGE` با نام کاربری (`service.go:333`). فیلدهای عوض‌شده نوشته نمی‌شوند.
- تغییر رمز: متن ثابت «تغییر رمز عبور» (`service.go:376`).
- تنظیم امنیتی: `SECURITY_SETTING_CHANGE` با نام کلید و مقدار جدید (`settings/service.go:83` و `:105`). مقدار قبلی در رویداد نیست. مقدار قبلی فقط تا وقتی ردیف DB خوانده شود در حافظه است و در لاگ کپی نمی‌شود.

`ROLE_CHANGE` استفاده نمی‌شود.

---

## د) تنظیمات امنیتی

### ۱۶. کلیدها

منبع: `internal/platform/security/settings/keys.go`. اعتبارسنجی هنگام ذخیره وجود ندارد؛ `UpdateSetting` / `UpdateSettingByID` هر رشته‌ای را امضا و upsert می‌کنند (`service.go` خط ۶۱ به بعد). UI یک Input آزاد است.

| کلید | برچسب فارسی | پیش‌فرض | نوع | کجا خوانده می‌شود | اعتبارسنجی خواندن |
|---|---|---|---|---|---|
| MaximumNumberOfSessionsPerUser | حداکثر تعداد نشست همزمان | 3 | عدد | `session/service.go:34` | خطا ⇒ ۳ |
| MaximumNumberOfFailedLogin | بیشترین تعداد ورود اشتباه | 5 | عدد | `loginguard/guard.go:132` | خطا یا ≤۰ ⇒ ۵ |
| MaximumTimeOfUserIpBeingLocked | بیشترین مدت زمان قفل بودن اکانت | 30 | دقیقه | `guard.go:136` | خطا یا ≤۰ ⇒ ۳۰ |
| MaximumSizeOfLogTables | حداکثر تعداد لاگ | 100 | مگابایت در worker، نه «تعداد» | `worker.go:90` | خطا یا ≤۰ ⇒ ۱۰۰ |
| LastVerifiedEventId | آخرین رویداد ثبت شده | 0 | شناسهٔ عدد | `worker.go:161` | `ParseInt` نامعتبر نادیده گرفته می‌شود و checkpoint صفر می‌ماند |
| WorkHours_Start | ساعت شروع کاری | 08:00 | ساعت | `guard.go:168` | تجزیهٔ ساعت؛ شکست یعنی رد ساعات کاری طبق همان تابع |
| WorkHours_End | ساعت پایان کاری | 18:00 | ساعت | `guard.go:169` | همان |
| IsSaturdayBlocked | تعطیلی شنبه | false | بولین رشته‌ای | `guard.go:207` | فقط برابر دقیق `true` |
| IsSundayBlocked | تعطیلی یکشنبه | true | بولین | همان | فقط `true` |
| IsMondayBlocked | تعطیلی دوشنبه | false | بولین | همان | فقط `true` |
| IsTuesdayBlocked | تعطیلی سه شنبه | false | بولین | همان | فقط `true` |
| IsWednesdayBlocked | تعطیلی چهارشنبه | false | بولین | همان | فقط `true` |
| IsThursdayBlocked | تعطیلی پنجشنبه | false | بولین | همان | فقط `true` |
| IsFridayBlocked | تعطیلی جمعه | true | بولین | همان | فقط `true` |
| Password_MinLength | کمترین مقدار رمزعبود | 8 | عدد | `passwordpolicy/validator.go:24` | ≤۰ ⇒ ۸. مقدار ۵۰۰۰ معتبرِ کد است و همان پیام حداقل را می‌سازد |
| Password_RequireUppercase | حروف بزرگ استفاده شود | true | بولین | `validator.go:32` | فقط برابر `true` الزام می‌سازد |
| Password_RequireLowercase | حروف کوچک مجاز است | true | بولین | `validator.go:37` | فقط `true` |
| Password_RequireDigit | عدد مجاز است | true | بولین | `validator.go:42` | فقط `true` |
| Password_RequireSpecial | حروف علامت دار مجاز است | true | بولین | `validator.go:47` | فقط `true` |
| MaximumUploadsPerWindow | بیشترین تعداد آپلود | 10 | عدد | `uploadguard/guard.go:33` | خطای تبدیل نادیده گرفته می‌شود |
| UploadWindowMinutes | بیشترین زمان آپلود | 60 | دقیقه | `uploadguard/guard.go:37` | همان |
| OrgIntegrityHashMigrationV2 | در دیکشنری نیست؛ UI نشان نمی‌دهد | 0 | پرچم یک‌بار | `cmd/server/main.go:193` | بدون اعتبارسنجی |
| OrgIntegrityHashMigrationV3 | در دیکشنری نیست | 0 | پرچم یک‌بار | `main.go:207` | بدون اعتبارسنجی |

`GetAllSettings(true)` نام کلید را با برچسب فارسی عوض می‌کند و کلیدهای خارج از دیکشنری را حذف می‌کند (`settings/model.go` خط ۷۲ تا ۸۰). مقدار فعلی DB: منتظر خروجی اسکریپت (`=== Q16 ===`). اسکریپت `IntegrityHash` را چاپ نمی‌کند.

### ۱۷. آخرین رویداد ثبت‌شده

کلید `LastVerifiedEventId` است (`keys.go` خط ۹). نویسنده: `verifyAuditChain` بعد از `VerifyChain` موفق، با `UpdateSetting(..., actor 0, ip "system")` (`worker.go` خط ۱۶۰ تا ۱۷۷). خواننده: همان worker در دور بعد، و صفحهٔ تنظیمات چون کلید در دیکشنری فارسی است.

اگر کاربر مقدار را دستی عوض کند، دور بعد زنجیره از همان عدد بررسی می‌شود. اگر عدد نباشد، `ParseInt` خطا را دور می‌ریزد و بررسی از ۰ شروع می‌شود.

نقص کد، مستقل از مقدار DB: وقتی `fromID > 1` است، `VerifyChain` به‌جای هش ردیف بلافاصله قبل، `FindFromID(fromID-1)` را می‌گیرد و `RowHash` **آخرین** ردیف آن مجموعه را به‌عنوان `prevHash` می‌گذارد (`manager.go` خط ۸۰ تا ۸۷). پس از اولین تأیید کامل، دورهای بعدی می‌توانند `DATA_TAMPERING` اشتباه بدهند و checkpoint جلو نرود.

### ۱۸. حداکثر حجم جداول لاگ

در UI هست، با برچسب «حداکثر تعداد لاگ» (`keys.go` خط ۶۵). worker فقط حجم `SecurityEvents` را با `GetTableSizeMB` می‌سنجد و اگر از عدد (به‌عنوان مگابایت) بزرگ‌تر باشد، قدیمی‌ترین ۱۰٪ را حذف می‌کند (`worker.go` خط ۸۹ تا ۱۱۹). جدول دیگری را با این کلید پاک نمی‌کند. درج `LogCleanupSummaries` خطایش دور ریخته می‌شود.

حجم و تعداد ردیف فعلی: منتظر خروجی اسکریپت (`=== Q18 ===`).

---

## هـ) ایجاد کاربر

### ۱۹. مسیر و خطاها

`POST` ایجاد کاربر → `Handler.CreateUser` (`internal/user/handler.go` خط ۷۴ تا ۸۹) → `Service.CreateUser` (`service.go` خط ۵۴ تا ۱۱۰) → `repo.Create`.

متن‌هایی که خود سرویس برمی‌گرداند (API فقط `UserMsg` و `code` را می‌فرستد، `internal/platform/middleware/common.go` خط ۹۴ تا ۱۰۵):

| محل | کد | متن کاربر |
|---|---|---|
| bind نامعتبر (فیلد required خالی) | INTERNAL، چون خطای gin از نوع `*apperror.AppError` نیست | پیام عمومی خطای داخلی، HTTP 500 |
| سیاست رمز | WEAK_PASSWORD | اولین جملهٔ اعتبارسنج. حداقل طول: «رمز عبور باید حداقل N کاراکتر باشد.» (`validator.go:29`). حرف بزرگ/کوچک/رقم/ویژه هر کدام جملهٔ جدا دارند (خط ۳۴ تا ۴۹) |
| نقش ناموجود | INVALID_ROLE | نقش انتخاب‌شده معتبر نیست. |
| خطای DB نقش یا مجوز | DB_ERROR | خطا در خواندن نقش. / خطا در خواندن مجوزهای نقش. |
| نقش با هش خراب | INVALID_ROLE | نقش انتخاب‌شده به دلیل نقض یکپارچگی قابل انتصاب نیست. |
| bcrypt، از جمله رمز بلندتر از ۷۲ بایت | HASH_ERROR | خطا در پردازش رمز عبور. |
| SecurityCode | ENCRYPT_ERROR | خطا در تولید کد امنیتی. |
| درج DB | DB_ERROR | خطا در ایجاد کاربر. |

جملهٔ «حداقل ۵۰۰۰ کاراکتر» از خود کد به‌عنوان عدد ثابت نیامده است. فقط اگر `Password_MinLength` به ۵۰۰۰ پارس شود، `validator.go:29` همان قالب را با N=5000 می‌سازد. پیش‌فرض کد ۸ است. فرانت هم `MIN_PASSWORD_LENGTH = 8` را جداگانه دارد و این سقف سرور را نشان نمی‌دهد. مقدار واقعی تنظیم: منتظر خروجی اسکریپت.

`CreateUser` همیشه `UserTypeUser` می‌گذارد (`service.go:79`). بدنهٔ `CreateUserRequest` فیلد `user_type` ندارد (`dto.go` خط ۴ تا ۱۳)؛ اگر فرانت بفرستد، سرور آن را نادیده می‌گیرد.

نتیجهٔ HTTP ایجاد `afta_probe_<زمان>`: منتظر خروجی اسکریپت (`=== Q19 ===`). اسکریپت کاربر آزمایشی را در پایان حذف می‌کند و ردیف ممیزی را پاک نمی‌کند.

### ۲۰. `UserType` و `Role`

`RoleID` کلید نقش دسترسی است و به جدول Roles وصل است. مجوزها از RolePermissions خوانده می‌شوند. فرم ایجاد این را می‌فرستد و سرور الزامی‌اش می‌داند.

`UserType` برچسب شغلی است: `User`، `Doctor`، `Assistant`، `Specialist` (`model.go` خط ۹ تا ۱۶). در ایجاد همیشه `User` ذخیره می‌شود. در ویرایش، `UpdateUserRequest.UserType` پذیرفته می‌شود (`dto.go` خط ۲۴). `FindDoctors` نوع Doctor و Specialist را می‌گیرد و `FindAssistants` نوع Assistant را (`repo.go` خط ۱۰۴ و ۱۱۰). این ستون در فرمول SecurityCode نیست و در HMAC کاربر هم نیست (`BuildUserIntegrityFields`، `repo.go` خط ۲۴۹ تا ۲۶۴).

---

## و) فرانت‌اند

### ۲۱. `frontend/src/platform/components`

شمارش، تعداد فایل واردکننده بیرون از خود پوشهٔ کامپوننت است.

| کامپوننت | کار | استفاده‌کننده |
|---|---|---|
| PageHeader | عنوان صفحه و اکشن‌ها | ۲۲ صفحه |
| DataTable | جدول آنت‌دیزاین با صفحه‌بندی کلاینت | ۱۸ فایل (۱۵ صفحه + ۳ مودال بیمار/کیف‌پول/تاریخچه) |
| FormLayout | قاب فرم | ۰؛ فقط `index.ts` خودش |
| ConfirmDialog | تأیید حذف/عمل | ۱۱ |
| ErrorBoundary | خطای رندر | ۱ (`App.tsx`) |
| JalaliDatePicker | تاریخ جلالی | ۸ |
| MultiSelectWithActions | چندانتخابی با عملیات | ۳ صفحهٔ گزارش |
| tooths/tooth.tsx (ToothSelection) | نمودار دندان | ۱ (`ToothChartModal`) |
| toothQuadrant.ts و toothChartSvg.ts | داده و SVG داخلی نمودار | فقط خود نمودار |

### ۲۲. جدول خام، فرم بدون react-hook-form، استایل inline

`Table` خام آنت‌دیزاین:

- `frontend/src/modules/tariff/pages/TariffListPage.tsx` (دو جدول، حدود خط ۳۸۸ و ۴۳۷) علاوه بر DataTable
- `frontend/src/platform/components/DataTable/DataTable.tsx`
- `frontend/src/platform/reports/components/ReportTable.tsx` (صفحه‌های گزارش از ReportViewer به این می‌رسند)

فرم آنت‌دیزاین بدون `react-hook-form`:

- `modules/settings/pages/SettingsPage.tsx`
- `modules/profile/pages/ProfilePage.tsx`
- `modules/reports/pages/DoctorPerformancePage.tsx`
- `modules/reports/pages/RevenueByOrganizationPage.tsx`
- `modules/wallet/components/PatientWalletLedgerModal.tsx`
- `modules/reception/components/DoctorSelection.tsx`
- `modules/reception/components/InsuranceSelection.tsx`
- `modules/reception/components/PatientInfo.tsx`
- `platform/reports/components/ReportPageSetup.tsx`
- `platform/components/FormLayout/FormLayout.tsx`
- `modules/reception/pages/ReceptionWorkspacePage.tsx` فرم آنت‌دیزاین نیست؛ state محلی/zustand است و react-hook-form ندارد

`style={{` در این فایل‌ها هست: صفحه‌های گزارش CreditAccountPayments، UserPerformance، DoctorPerformance، RevenueByOrganization؛ AppShell و AuthLayout و RequireAuth و LoginPage؛ Patients، Roles، Services، Regulations، SpecialCodes، Tariff، OrganizationPackages، Profile؛ مودال‌ها و فرم‌های wallet و reception (از جمله ServiceRow، ServicesTable، ReceptionTotalsBar، PatientInfo، InsuranceSelection، DoctorSelection)؛ ReportHeader، ReportFooter، ReportSummary، ReportViewer، ReportPageSetup؛ PageHeader و MultiSelectWithActions. رنگ هگز هم در همین مجموعه و در `index.css` پراکنده است. فهرست فایلِ `style={{` همان نتیجهٔ جست‌وجوی بالا است.

### ۲۳. Enter به ورودی بعدی

سراسری نیست. فقط در پذیرش: `ServiceRow.tsx` خط ۲۰۰ و ۲۴۶ (`onPressEnter` به فیلد بعد) و `InsuranceSelection.tsx` خط ۲۱۰ (Enter کد خاص را حل می‌کند). `useReceptionHotkeys` میانبر صفحه است، نه پیمایش عمومی فرم.

### ۲۴. موتور گزارش و فیش ۷۰ میلی‌متری

`frontend/src/platform/reports`: `constants.ts` پیش‌فرض A4 عمودی با حاشیهٔ ۱۲ میلی‌متر (medium). presetها ۸ و ۱۲ و ۱۸ میلی‌متر و اندازهٔ A4/A5/Letter/Legal/سفارشی. سربرگ و پاورقی کامپوننت‌های `ReportHeader` و `ReportFooter` هستند. چاپ با `window.open` و `window.print` در `exporters/print.ts`. PDF در `exporters/pdf.tsx` با فونت محلی Vazirmatn.

فیش پذیرش جزء این موتور نیست. `frontend/src/modules/reception/printReceptionReceipt.ts` خط ۱۰۴ تا ۱۰۵ `@page { size: 70mm auto; }` و عرض ۷۰ میلی‌متر دارد. `receptionPrintActions.ts` آن را صدا می‌زند و همان HTML را برای چاپ باز می‌کند.

### ۲۵. فونت و asset از اینترنت

در `frontend/src` ارجاع به fonts.google، cdn، unpkg، jsdelivr یا `@import` فونت خارجی پیدا نشد. `index.css` فقط نام `'Vazirmatn'` را در `font-family` می‌گذارد و `@font-face` ندارد. PDF به `/fonts/Vazirmatn-Regular.ttf` و `Vazirmatn-Bold.ttf` اشاره می‌کند (`constants.ts` خط ۳۵ تا ۳۷). فایل ttf/woff/otf داخل مخزن نیست. `xmlns` روی SVG بارگذاری شبکه نیست.

---

## ز) بک‌اند

### ۲۶. بزرگ‌ترین فایل‌ها و تابع‌ها

شمارش خط فایل (بدون فایل تولیدشده):

1. `internal/reception/service.go` — ۱۳۵۲
2. `internal/wallet/service.go` — ۱۰۳۵
3. `internal/user/service.go` — ۹۸۳
4. `internal/report/service.go` — ۸۲۴
5. `internal/report/repo.go` — ۶۷۸
6. `internal/user/handler.go` — ۴۴۸
7. `internal/report/dto.go` — ۴۰۳
8. `internal/report/user_performance.go` — ۴۰۰
9. `internal/services/service.go` — ۳۸۶
10. `internal/organization/service.go` — ۳۴۸
11. `internal/user/repo.go` — ۳۳۸
12. `cmd/server/main.go` — ۳۳۶
13. `internal/tariff/service.go` — ۳۳۵
14. `internal/patient/service.go` — ۳۳۲
15. `internal/wallet/repo.go` — ۳۳۱

بلندترین تابع‌ها با شمارش بدنه (کامنت و رشته حذف شده):

1. `main` — `cmd/server/main.go` حدود ۴۳ تا ۳۰۰
2. `RevenueByOrganization` — `report/service.go` حدود ۱۱۱ تا ۲۶۱
3. `DoctorPerformanceByOrganization` — حدود ۵۰۱ تا ۶۰۹
4. `UpdateReception` — `reception/service.go` حدود ۲۱۶ تا ۳۲۲
5. `CreateReception` — حدود ۱۱۸ تا ۲۱۳
6. `seed-admin` main — حدود ۲۲ تا ۱۱۶
7. `EndReception` — حدود ۱۲۰۱ تا ۱۲۹۴
8. `AdjustCardToCard` — `wallet/service.go` حدود ۱۲۵ تا ۲۱۷
9. `toResponse` پذیرش — حدود ۸۴۴ تا ۹۳۶
10. `DoctorPerformanceByService` — حدود ۶۱۲ تا ۶۹۵
11. `CreditAccountPayments` — `credit_account_payments.go` حدود ۱۴ تا ۹۷
12. `ChargeFromCreditAccount` — wallet حدود ۲۷۳ تا ۳۴۹
13. `HandleUserLogin` — `user/service.go` حدود ۱۱۴ تا ۱۸۸
14. `UserPerformanceByReception` — حدود ۲۵۹ تا ۳۳۳
15. `SaveTariffsForOrganization` — `tariff/service.go` حدود ۱۵۹ تا ۲۲۵

### ۲۷. منطق تکراری قابل انتقال به `internal/platform`

- پیچیدن `gorm.ErrRecordNotFound` و خطاهای DB در `apperror` در هر سرویس جدا تکرار شده است.
- `_ = audit.LogEvent` در همه‌جا؛ شکست ثبت ممیزی گم می‌شود.
- ساخت و بررسی HMAC/رمز برای هر موجودیت جداگانه است، با الگوی یکسان.
- صفحه‌بندی سرور مشترک وجود ندارد؛ هر لیست یا همه‌چیز را برمی‌گرداند یا در فرانت برش می‌خورد.
- الگوی هندلر (bind، خواندن actor، ClientIP، WriteError) کپی شده است.
- تبدیل رشتهٔ تنظیم به bool/int در چند بسته با قواعد کمی متفاوت است.
- `FormLayout` استفاده نمی‌شود و صفحه‌ها فرم آنت‌دیزاین را دوباره چیده‌اند.

### ۲۸. نسخهٔ Go

`go version` روی همین ماشین: `go1.26.4 windows/amd64`. `go.mod` خط ۳: `go 1.26.4`. یکی هستند.

---

## ح) آمادگی پیامک مرحلهٔ دوم

### ۲۹. موبایل کاربر

فیلد `PhoneNumber` روی `User` است (`internal/user/model.go` خط ۲۷)، `size:20`، بدون `not null`. SQL `0004` آن را `NVARCHAR(20) NULL` کرده (خط ۴۶). `CreateUserRequest.PhoneNumber` تگ `binding:"required"` ندارد (`dto.go` خط ۱۰). در فرانت اختیاری است و الگوی `^09\d{9}$` یا خالی دارد.

در SecurityCode هست: `sensitiveFields` نام، نام خانوادگی، تلفن، هش رمز، آدرس، RoleID و کد نظام را به هم می‌چسباند (`encryption/user_encryption.go` خط ۳۰ تا ۳۲). در HMAC کاربر هم هست (`repo.go` خط ۲۵۶). `TwoFactorEnabled` و `TwoFactorSecret` فقط در HMAC هستند (خط ۲۶۲ تا ۲۶۳)، نه در SecurityCode.

### ۳۰. پیامک، OTP، TwoFactor

غیر از ستون‌های `TwoFactorEnabled` و `TwoFactorSecret` (`model.go` خط ۳۷ تا ۳۸) و همان دو فیلد در `BuildUserIntegrityFields`، پیاده‌سازی SMS، OTP، ارسال کد، یا بررسی راز مرحلهٔ دوم در کد Go و فرانت پیدا نشد. جست‌وجوی `TwoFactor` فقط همان تعریف و HMAC را برگرداند.

### ۳۱. جریان `POST /api/login`

1. میدل‌ور سراسری: Recovery، هدر امنیتی، لاگر درخواست، CORS، rate limit کلی (`main.go` خط ۱۵۹ تا ۱۶۳).
2. گروه `/api`: `LoginRateLimiterMiddleware` قفل IP را برای هر مسیر API چک می‌کند (`ratelimit.go:52`). اگر قفل باشد، `ErrIPLocked` و قطع.
3. `SessionMiddleware` مسیر عمومی `/api/login` را رد می‌کند.
4. نقش و مجوز، سپس CSRF. مسیر عمومی از CSRF می‌گذرد.
5. `SecurityCheckMiddleware` و `AuthorizationMiddleware`.
6. `Handler.Login` (`handler.go:33`): JSON را bind می‌کند. خطای bind به HTTP 500 عمومی تبدیل می‌شود.
7. `HandleUserLogin` (`service.go:114`): قفل IP، یافتن کاربر، یکپارچگی کاربر، یکپارچگی نقش، IsLocked/IsActive، ساعات کاری (ادمین با همهٔ مجوزها معاف است)، bcrypt، SecurityCode، صفر کردن تلاش IP، `CreateSession` (حذف خاموش قدیمی‌ترین نشست اگر از سقف گذشته باشد)، به‌روزرسانی `LastLoginAt`، `LOGIN_SUCCESS`.
8. کوکی httpOnly نشست و کوکی CSRF (نه httpOnly). JSON شامل `SessionID` است (`handler.go:46` تا ۵۱).

جای مرحلهٔ دوم: بعد از موفقیت رمز و SecurityCode (خط ۱۶۷) و قبل از `CreateSession` و `SetSessionCookie` (خط ۱۷۵ و handler خط ۴۶). امروز بین رمز درست و ساخت نشست هیچ وقفه یا چالش دومی نیست.

### ۳۲. کلاینت HTTP بیرونی

`http.Client` و `http.NewRequest` در کد Go برنامه پیدا نشد. `internal/cardreader` شبیه‌ساز دستگاه در همان فرایند است و به شبکه وصل نمی‌شود. الگوی مشترک برای فراخوانی سرویس خارجی (پیامک یا درگاه) وجود ندارد.

---

## ریسک‌ها و موارد غیرمنتظره

- دو منبع اسکیما با هم اختلاف دارند و سرور فقط AutoMigrate را اجرا می‌کند. اگر پایگاه توسعه با SQL ساخته شده باشد، ستون `UserType` و جدول‌های دامنه ممکن است با مدل Go یکی نباشند. تا خروجی Q6 این معلوم نیست.
- `VerifyChain` بعد از checkpoint بزرگ‌تر از ۱، هش آخرین ردیف را به‌جای هش ردیف قبلی برمی‌دارد. تأیید دوره‌ای می‌تواند هشدار دستکاری اشتباه بدهد.
- برچسب «حداکثر تعداد لاگ» در worker به‌عنوان مگابایت فقط برای `SecurityEvents` به کار می‌رود. خلاصهٔ پاکسازی اگر جدول نباشد بی‌صدا حذف می‌شود و خودِ حذف لاگ، زنجیرهٔ هش را از وسط می‌بُرد چون ردیف‌های قدیمی حذف می‌شوند و `PrevHash` ردیف بعدی به ردیف حذف‌شده اشاره دارد.
- CSRF بدون کاربر، UserID برابر ۰ ذخیره می‌کند. ورود ناموفقِ کاربر ناموجود NULL ذخیره می‌کند. UI هر دو را شبیه «بدون کاربر واقعی» نشان می‌دهد.
- `LoginRateLimiterMiddleware` روی کل `/api` است. قفل IP، کار پذیرش را هم می‌بندد.
- حذف نشست به‌خاطر سقف، و `DeleteAllUserSessions`، رویداد ندارند. توضیح `SESSION_REVOKED` خودِ UUID نشست را در متن لاگ می‌نویسد.
- تغییر تنظیم، نقش و کاربر مقدار قبلی را لاگ نمی‌کند. تغییر تنظیم نشست‌های باز را باطل نمی‌کند.
- `UpdateSettingByID` نام کلید و مقدار را با `fmt.Println` روی stdout می‌نویسد (`settings/service.go` حدود خط ۶۷).
- `Password_MinLength` بدون سقف ذخیره می‌شود. مقدار ۵۰۰۰ همان خطای «حداقل N کاراکتر» را می‌سازد؛ رمز طولانی‌تر از ۷۲ بایت بعداً در bcrypt با «خطا در پردازش رمز عبور» می‌میرد. خطای bind فیلد خالی HTTP 500 عمومی است.
- ایجاد کاربر نوع شغلی را دور می‌ریزد. پزشک و دستیار فقط اگر بعداً `user_type` ویرایش شود در لیست پزشک/دستیار می‌آیند.
- `GetAllSettings(true)` کلید را با متن فارسی جایگزین می‌کند. دو کلید مهاجرت سازمان در UI نیستند ولی هنگام استارت نوشته می‌شوند.
- بولین‌ها فقط با رشتهٔ دقیق `true` فعال می‌شوند. کلمهٔ فارسی یا `True` یعنی «خاموش».
- صفحهٔ لاگ تاریخ را ممکن است جلالی بفرستد و سرور میلادی بخواند؛ همراه با برابری دقیق IP، فیلتر «کار نمی‌کند».
- فونت Vazirmatn در مخزن نیست؛ PDF و CSS فقط به نام یا مسیر محلی اشاره می‌کنند.
- `FormLayout` مرده است. تست یکپارچه، CI، و کلاینت HTTP بیرونی وجود ندارد. ستون‌های دو مرحله‌ای فقط برای HMAC آماده‌اند.
- README مسیر migrate را طوری توضیح می‌دهد که کد فعلی اجرا نمی‌کند.
