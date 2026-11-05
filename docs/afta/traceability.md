# ردیابی الزام افتا

ماتریس الزام پروفایل حفاظتی، ایراد گزارش ۱ مشاور، یافته discovery و کد. شرح ایرادهای Report1 از `docs/afta/consultant/Report1.pdf` خلاصه شده است. فاز و الزام PP همان ایرادها از بخش ۲ `docs/afta/roadmap.md` است. ردیف‌های D-11 به بعد از بخش‌های ۳-ب و ۳-ج همان نقشه هستند. شکاف‌های PP از بخش ۳ هستند.

وضعیت فقط یکی از این مقدارهاست: باز، در حال انجام، منتظر تست DB، انجام‌شده — منتظر مشاور، تأییدشده مشاور.

شناسه‌های D-01، D-02، D-08، D-09، D-10 و ردیف CI از قبل بودند. D-03 تا D-07 اختصاص داده نشده است. D-10 با وضعیت ثبت‌شده «انجام‌شده» مانده است. توضیح وضعیت قبلی D-02 (فاز ۵ و migration داده) به ستون فاز و یادداشت رفته و وضعیتش باز است، چون آن جمله در فهرست وضعیت‌ها نبود.

این یافته‌ها ردیف تازه نگرفتند چون شناسه دارند: خطای bind (D-08)، چاپ stdout (D-09)، جدول‌های `LogCleanupSummaries` و `UploadCounters` و مسیر AutoMigrate (D-01)، `Funds.Balance` از نوع float (D-02)، بیرون‌ماندن ماژول logs از git (D-10). یافته «نبود CI و helper تست» همان ردیف CI است. یافته «قوانین بیرون از `.cursor/rules`» با قرارگرفتن قوانین در `.cursor/rules` بسته شده و ردیف باز ندارد.

ستون‌ها در هر سه جدول یکی است.

## ردیف‌های قبلی

| شناسه | منبع | شرح کوتاه | الزام PP | فاز | فایل‌های کد | تست‌ها | وضعیت | یادداشت |
|---|---|---|---|---|---|---|---|---|
| CI | Discovery | زیرساخت تست واحد، helper فرانت و اسکریپت بررسی | — | ۰ | `internal/platform/testutil`، `internal/platform/security/audit/logger.go`، `scripts/afta-check.ps1`، `frontend/src/test/renderWithProviders.tsx` | `TestWriteError_AppErrorReturnsCodeAndMessage`، `TestWriteError_UnknownErrorIsGeneric500`، `TestComputeRowHash`، `TestIntegration_AutoMigrate_AllModelsCreate`، `PageHeader` | منتظر تست DB | وضعیت قبلی «منتظر اجرای تست DB» بود. تست integration هنوز منتظر اجرای کاربر است. `govulncheck` و `npm audit` تا فاز ۵ مسدودکننده نیستند (PP-19). |
| D-01 | Discovery | نبودن `LogCleanupSummaries` و `UploadCounters` در AutoMigrate، و ادامه سرور پس از خطای schema | FPT_FLS.1 | ۰ | `internal/platform/migrate/migrate.go`، `cmd/server/main.go` | `TestCombineMigrateErrors_EmptyIsNil`، `TestCombineMigrateErrors_NamesEveryTable`، `TestModelTableName_UsesTableNameThenTypeName`، `TestIntegration_AutoMigrate_AllModelsCreate`، `TestIntegration_LogCleanupAndUploadCounterTablesExist` | انجام‌شده — منتظر مشاور | خروجی discovery ساعت ۲۱:۲۳ (Q6) هر دو جدول را در پایگاه توسعه فهرست کرده است. |
| D-02 | Discovery | `Funds.Balance` از نوع float است | — | ۵ | `internal/fund/model.go` | — | باز | نیاز به migration داده. نوع ستون الان float می‌ماند. شاهد Q6. |
| D-08 | Discovery | خطای bind ورودی JSON و query به‌جای ۴۰۰، پاسخ HTTP 500 می‌دهد | — | ۰ | `internal/platform/apperror/apperror.go`، `internal/platform/middleware/bind.go`، handlerهای `ShouldBindJSON` و `ShouldBindQuery` | `TestD08_BindJSON_EmptyOrIncompleteReturnsValidationError`، `TestD08_BindJSON_ValidBodyFillsStruct`، `TestD08_BindQuery_InvalidValueDoesNotLeakFieldValue` | انجام‌شده — منتظر مشاور | پاسخ ۴۰۰ با کد `VALIDATION_ERROR`. نام فیلد و مقدار در پاسخ نیست. |
| D-09 | Discovery | چاپ نام و مقدار تنظیم، و ساعت کاری، در stdout | — | ۰ | `internal/platform/security/settings/service.go`، `internal/platform/security/loginguard/guard.go` | بررسی `fmt.Print` در `internal/` و `cmd/server` | انجام‌شده — منتظر مشاور | چاپ `UpdateSettingByID` و ساعت کاری حذف شد. `cmd/seed-admin` تا فاز ۶ استثناست (PP-14). |
| D-10 | Discovery | ماژول logs به‌خاطر الگوی `logs` در gitignore بیرون از git بود | — | ۰ | `.gitignore`، `frontend/.gitignore`، `internal/logs`، `frontend/src/modules/logs` | `go build ./...` و `npm run build` | انجام‌شده | الگوی `/logs/` فقط پوشه logs کنار همان فایل gitignore را نادیده می‌گیرد. |

## ایرادهای Report1

| شناسه | منبع | شرح کوتاه | الزام PP | فاز | فایل‌های کد | تست‌ها | وضعیت | یادداشت |
|---|---|---|---|---|---|---|---|---|
| R1-01 | Report1 | در لاگ امنیتی شناسه کاربر خالی یا فقط یک عدد است؛ هر ستون جستجو و مرتب‌سازی سمت سرور می‌خواهد؛ فیلتر IP کار نمی‌کند | FAU_GEN.2، FAU_SAR.3 | ۱ | `internal/logs`، `frontend/src/modules/logs`، جدول کلاینت‌ساید | — | باز | مشاور پرسیده شناسه کاربری همان نام کاربری است یا نه. هم‌پوشان با D-14 و D-16. |
| R1-02 | Report1 | نقش جدید ساخته شده ولی لاگ نمی‌گوید کدام کاربر آن را ساخته است | FAU_GEN.2، FMT_SMR | ۱ | ایجاد نقش در `internal/user` | — | باز | actor در رویداد ایجاد نقش نیست. |
| R1-03 | Report1 | هنگام ایجاد کاربر خطا داده می‌شود | FMT_MTD، FIA_PMG | ۳ | `internal/platform/security/passwordpolicy`، تنظیمات رمز | — | باز | تصویر خطا در متن استخراج‌شده از PDF نیامد. نقشه راه: مشاور حداقل طول را ۵۰۰۰ گذاشته و سامانه پذیرفته؛ پیام حداقل طول رفتار سیاست رمز است و ریشه، نبود اعتبارسنجی تنظیم است (R1-09، R1-10). |
| R1-04 | Report1 | پس از تغییر دسترسی نقش، توضیحات لاگ نمی‌گوید چه دسترسی‌هایی اضافه یا گرفته شده است | FMT_MSA، FMT_SMR | ۱ | به‌روزرسانی نقش | — | باز | diff مجوزها محاسبه نمی‌شود. |
| R1-05 | Report1 | صفحه نشست‌های کاربران برای همه ستون‌ها جستجو و مرتب‌سازی ندارد | FTA_TAH | ۱ | صفحه نشست‌ها | — | باز | همان محدودیت جدول کلاینت‌ساید. |
| R1-06 | Report1 | خروج به‌خاطر رسیدن نشست‌های فعال به سقف مجاز، لاگ امنیتی ندارد | FTA_MCS، FDP_ACF.4 | ۱ | مسیر حذف نشست قدیمی | — | باز | |
| R1-07 | Report1 | لاگ حذف نشست نام کاربری و این‌که چه کسی نشست چه کسی را بسته ندارد؛ خروج با دلیل، رویداد جدا نیست. شناسه نشست لازم نیست | FTA_SSL، FAU_GEN | ۱ | حذف نشست و خروج | — | باز | هر خروج باید دلیل مشخص داشته باشد. |
| R1-08 | Report1 | در لاگ، `Id=1` بی‌معناست و باید حذف شود؛ نام کاربری اضافه شود | FAU_SAR.2 | ۱ | نمایش رویداد امنیتی | — | باز | به‌جای شناسه عددی، نام کاربری. |
| R1-09 | Report1 | تنظیم بولی فقط True یا False بپذیرد و تنظیم عددی فقط بازه مجاز. مشاور پرسیده این تنظیم چیست و چه می‌کند | FMT_MTD | ۳ | `internal/platform/security/settings` | — | باز | نقشه راه: بیشتر تنظیمات schema ندارند. مقدار `True` با حرف بزرگ اعمال نمی‌شود (D-24). |
| R1-10 | Report1 | مقدار یک تنظیم باید عددی بین ۸ و ۲۵ باشد | FIA_PMG | ۳ | `Password_MinLength` | — | باز | نقشه راه این تنظیم را حداقل طول رمز می‌داند. در DB توسعه مقدار ۳ است (D-25). |
| R1-11 | Report1 | تنظیمات نامرتب‌اند؛ برای رمز باید برچسبی مثل «الزام وجود عدد در رمز عبور» باشد | AGD_OPE | ۳ | `SettingsDicFN` در `internal/platform/security/settings/keys.go` | — | باز | |
| R1-12 | Report1 | «مجاز است» در ساختار رمز غلط است؛ معنی درست «الزام استفاده از …» است | AGD_OPE | ۳ | برچسب‌های رمز در `keys.go` | — | باز | روشن‌بودن امروز یعنی استفاده اختیاری است، نه الزام. |
| R1-13 | Report1 | پس از هر تغییر تنظیم امنیتی همه نشست‌های فعال پاک شوند، برای هر کدام لاگ ثبت شود و کاربران دوباره وارد شوند | FIA_USB.3 | ۳ | ذخیره تنظیمات و نشست | — | باز | |
| R1-14 | Report1 | خطای دیگری هنگام ایجاد کاربر است و باید برطرف شود؛ تفاوت کاربرد دو فیلد فرم توضیح داده شود | FIA_ATD | ۴ | `internal/user/service.go`، فرم ایجاد کاربر | — | باز | تصویر خطا در PDF استخراج نشد. نقشه راه: `UserType` برای فهرست‌ها در برابر `Role` برای مجوز. فیلد نوع کاربر دور ریخته می‌شود (D-17). |
| R1-15 | Report1 | پس از احراز هویت ناموفق باید خود حساب قفل شود، نه فقط IP؛ اگر نام کاربری وجود ندارد IP قفل شود | FIA_AFL | ۴ | `internal/platform/security/loginguard` | — | باز | امروز قفل per-IP است. |
| R1-16 | Report1 | تنظیم «حداکثر حجم جداول لاگ» کجاست و آیا پیاده شده است | FAU_STG.3، FAU_STG.4 | ۳ | `internal/platform/security/settings/keys.go`، `internal/platform/security/logretention/worker.go` | — | باز | برچسب فعلی «حداکثر تعداد لاگ» است و worker آن را مگابایت `SecurityEvents` می‌خواند (D-19). |
| R1-17 | Report1 | لاگ تغییر تنظیمات فقط مقدار جدید را دارد؛ مقدار قبل و بعد هر دو لازم است | FMT_MTD، FAU_GEN | ۱ | ثبت رویداد تغییر تنظیم | — | باز | |
| R1-18 | Report1 | قفل‌شدن کاربر لاگ نمی‌شود و ورود ناموفق به‌خاطر قفل هم لاگ نمی‌شود | FIA_AFL، FAU_GEN | ۱ و ۴ | `internal/platform/security/loginguard` | — | باز | |
| R1-19 | Report1 | علاوه بر رمز، حداقل یک روش احراز هویت دومرحله‌ای لازم است و فعال‌بودنش از تنظیمات امنیتی روشن و خاموش شود | FIA_UAU.5 | ۴ | ستون‌های `TwoFactor*` در کاربر | — | باز | نقشه راه روش را پیامک می‌داند. منطق ورود دومرحله‌ای نیست. |
| R1-20 | Report1 | اگر کاربر بیش از مدت تنظیم‌شده هیچ API صدا نزند، نشست پاک شود، خارج شود و دوباره وارد شود؛ لاگ این رویداد ثبت شود | FTA_SSL.3 | ۴ | نشست | — | باز | سرور باید مرجع مهلت عدم فعالیت باشد. |
| R1-21 | Report1 | لاگ به‌روزرسانی کاربر باید بگوید کدام فیلد عوض شده است | FMT_MSA | ۱ | به‌روزرسانی کاربر | — | باز | diff فیلدها نیست. |
| R1-22 | Report1 | صفحه «لاگ‌های سیستمی» برای همه فراخوانی‌های API، با نام کاربری، زمان، نام API، ورودی، خروجی، IP و خطا، به‌همراه فیلتر و مرتب‌سازی | FAU_GEN.1 | ۲ | وجود ندارد | — | باز | آزمون اول FAU_GEN.1 در نقشه راه. |
| R1-23 | Report1 | تنظیمات آپلود دیده می‌شود ولی خود فرایند آپلود معلوم نیست در کدام بخش نرم‌افزار است | FDP_ITC | ۵ | آپلود عکس پذیرش، `internal/platform/security/uploadguard` | — | باز | `uploadguard` وصل نیست (D-27). |
| R1-24 | Report1 | وظیفه یک تنظیم روی صفحه تنظیمات نامشخص است | FAU_STG.2، FMT_MTD | ۳ | `LastVerifiedEventId` | — | باز | نقشه راه: نشانگر داخلی worker است، تنظیم کاربر نیست و تغییر دستی‌اش بررسی صحت لاگ را دور می‌زند. |

## یافته‌های Discovery

| شناسه | منبع | شرح کوتاه | الزام PP | فاز | فایل‌های کد | تست‌ها | وضعیت | یادداشت |
|---|---|---|---|---|---|---|---|---|
| D-11 | Discovery | `VerifyChain` بعد از checkpoint هش آخرین ردیف را به‌جای ردیف قبلی برمی‌دارد | FAU_STG.1 | ۱ | `internal/platform/security/audit/manager.go` | — | باز | هشدار دستکاری کاذب. checkpoint جلو نمی‌رود. شاهد داده در D-23. |
| D-12 | Discovery | پاک‌سازی لاگ قدیمی زنجیره هش را از وسط می‌برد | FAU_STG.1 | ۱ | `internal/platform/security/logretention/worker.go` | — | باز | پس از پاک‌سازی، بررسی صحت نامعتبر می‌شود. زنجیره لنگر (anchor) می‌خواهد. |
| D-13 | Discovery | خطای ثبت ممیزی در سرویس‌ها با `_ =` نادیده گرفته می‌شود | FAU_STG.4 | ۱ | سرویس‌هایی که `LogEvent` را دور می‌ریزند | — | باز | شکست ممیزی بی‌صدا می‌ماند. |
| D-14 | Discovery | CSRF بدون کاربر، `UserID=0` ذخیره می‌کند | FAU_GEN.2 | ۱ | `internal/platform/middleware/csrf.go` | — | باز | بخشی از علت R1-01. |
| D-15 | Discovery | قفل IP روی کل `/api` اعمال می‌شود | FIA_AFL | ۴ | `internal/platform/middleware/ratelimit.go` | — | باز | قفل یک IP کار پذیرش را می‌بندد. |
| D-16 | Discovery | فیلتر تاریخ لاگ ممکن است تاریخ شمسی بفرستد و سرور میلادی بخواند؛ فیلتر IP برابری دقیق است | FAU_SAR.3 | ۱ | `frontend/src/modules/logs`، `internal/platform/security/audit/model.go` | — | باز | بخشی از «فیلتر کار نمی‌کند» در R1-01. |
| D-17 | Discovery | ایجاد کاربر همیشه `UserType=User` می‌گذارد و فیلد ارسالی را دور می‌ریزد | FIA_ATD | ۴ | `internal/user/service.go` | — | باز | پزشک و دستیار در فهرست‌ها نمی‌آیند. بخشی از R1-14. |
| D-18 | Discovery | رمز بلندتر از ۷۲ بایت در bcrypt شکست می‌خورد | FIA_PMG | ۳ | `internal/user/service.go` | — | باز | سقف حداقل طول رمز باید حداکثر ۷۲ باشد. بازه ۸ تا ۲۵ مشاور با این سقف سازگار است. |
| D-19 | Discovery | برچسب «حداکثر تعداد لاگ» در عمل مگابایت است و فقط `SecurityEvents` را می‌سنجد | FAU_STG.3، FAU_STG.4 | ۳ | `internal/platform/security/logretention/worker.go`، `keys.go` | — | باز | مصداق R1-16. |
| D-20 | Discovery | `index.css` برای رابط `@font-face` ندارد و به فونت نصب‌شده روی سیستم وابسته است | AVA_VAN.1 | ۵ | `frontend/src/index.css`، `frontend/public/fonts` | — | باز | فایل‌های Vazirmatn محلی هستند. نمایش در کلاینت بدون آن فونت خراب می‌شود. |
| D-21 | Discovery | `FormLayout` استفاده نمی‌شود؛ فرم‌ها بدون react-hook-form، استایل inline و جدول خام مانده‌اند | — | ۱ | `frontend/src` | — | باز | نقشه راه آن را داخل فازهای ۱ و ۳ می‌گذارد. |
| D-22 | Discovery | فیش ۷۰ میلی‌متری پذیرش بیرون از موتور گزارش است | — | ۵ | `frontend` چاپ فیش پذیرش | — | باز | انتقال به `ReceiptTemplate`. |
| D-23 | Discovery | ۳۸۶۱ رویداد از ۴۰۹۶ رویداد لاگ امنیتی `DATA_TAMPERING` است | FAU_STG.2 | ۱ | جدول `SecurityEvents` | — | باز | Q9 و Q10. حدود ۳۸۷۷ رویداد بدون کاربر. `LastVerifiedEventId=6`. اثر باگ D-11. محیط ارزیابی باید پایگاه تمیز داشته باشد. |
| D-24 | Discovery | `IsSaturdayBlocked=True` و `IsFridayBlocked=False` با حرف بزرگ اعمال نمی‌شود | FMT_MTD | ۳ | `internal/platform/security/loginguard` | — | باز | Q16. کد فقط رشته دقیق `true` را می‌پذیرد. تعطیلی شنبه‌ای که مشاور گذاشته اعمال نمی‌شود. مصداق R1-09. |
| D-25 | Discovery | `Password_MinLength=3` در پایگاه توسعه ذخیره شده است | FIA_PMG | ۳ | `internal/platform/security/settings` | — | باز | Q16. کمتر از حداقل ۸ در FIA_PMG. سامانه مقدار غیرمجاز را پذیرفته. مصداق R1-10. |
| D-26 | Discovery | `MaximumTimeOfUserIpBeingLocked=1` دقیقه و `MaximumNumberOfFailedLogin=2` بدون بازه تعریف‌شده ذخیره شده‌اند | FMT_MTD | ۳ | `internal/platform/security/settings` | — | باز | Q16. مقدارها در کد پذیرفته می‌شوند ولی بازه مجاز تعریف نشده است. |
| D-27 | Discovery | `uploadguard` در برنامه فراخوانی نمی‌شود | FDP_ITC | ۵ | `internal/platform/security/uploadguard` | — | باز | پاسخ R1-23. جدول `UploadCounters` پس از D-01 در پایگاه هست؛ خود گارد به مسیر آپلود وصل نیست. |
| D-28 | Discovery | حجم جدول لاگ حدود ۱٫۷ مگابایت است و آستانه ۱۰۰ مگابایت است | FAU_STG.3، FAU_STG.4 | ۳ | `SecurityEvents` | — | باز | Q18. پاک‌سازی هرگز اجرا نشده. آزمون آستانه به مقدار کوچک در تست نیاز دارد. |
| D-29 | Discovery | فیلدهای بیمار در پایگاه متن ساده هستند | — | ۵ | مدل بیمار | — | باز | Q6: `national_code`، نام، تلفن، آدرس. الزام در PP-07. |
| D-30 | Discovery | مجوز بدون route و متغیر محیطی بی‌اثر مانده است | — | ۶ | کاتالوگ مجوز، `internal/platform/config` | — | باز | `fund.update`، `fund.delete`، `tariff.delete` مسیر ندارند. `LOGIN_RATE_LIMIT_PER_MINUTE` و `INTEGRITY_TICKER` خوانده می‌شوند و در اجرا اثر ندارند. دوگانگی migration در D-01 بسته شده است. |

## شکاف‌های PP

| شناسه | منبع | شرح کوتاه | الزام PP | فاز | فایل‌های کد | تست‌ها | وضعیت | یادداشت |
|---|---|---|---|---|---|---|---|---|
| PP-01 | PP | آغاز و توقف توابع ممیزی، و ثبت خواندن لاگ و تلاش ناموفق برای خواندن آن، ثبت نمی‌شود | FAU_GEN.1 | ۱ | راه‌اندازی سرور و API لاگ | — | باز | راه‌اندازی و خاموشی سرور رویداد ممیزی ندارند. |
| PP-02 | PP | انتخاب رویدادهای قابل ممیزی بر اساس نوع رویداد و کاربر وجود ندارد | FAU_SEL.1 | ۱ | تنظیم ممیزی | — | باز | |
| PP-03 | PP | هنگام عبور حجم لاگ از آستانه به مدیر هشدار داده نمی‌شود | FAU_STG.3 | ۳ | `logretention`، رابط مدیر | — | باز | اعلان در UI نیست. برچسب و واحد آستانه در D-19 است. |
| PP-04 | PP | نتیجه بررسی صحت لاگ به مدیر نشان داده نمی‌شود | FAU_STG.1، FAU_STG.2 | ۱ | زنجیره هش `audit` | — | باز | زنجیره هست. تشخیص حذف یا تغییر غیرمجاز به‌خاطر D-11 و D-12 قابل اتکا نیست. |
| PP-05 | PP | شکست بررسی صحت رکوردها گزارش و هشدار نمی‌شود | FDP_SDI.2، FDP_SDI.3 | ۵ | HMAC و `SecurityCode` | — | باز | امروز بیشتر محاسبه و ذخیره است تا هشدار. |
| PP-06 | PP | خروجی Excel و PDF سمت کلاینت است: مجوز جدا، لاگ ممیزی، و مشخصه کاربر و زمان روی فایل ندارد | FDP_ETC | ۵ | `frontend/src/platform/reports` | — | باز | |
| PP-07 | PP | فیلدهای حساس بیمار رمزنگاری ستونی و شاخص جستجو ندارند | — | ۵ | مدل بیمار | — | باز | plaintext در D-29 تأیید شده. تصمیم نهایی با مشاور است. |
| PP-08 | PP | وضعیت حساب (فعال، غیرفعال، قفل) و روش احراز هویت هر کاربر مدل نشده است | FIA_ATD.1 | ۴ | `internal/user` | — | باز | |
| PP-09 | PP | نشست اول از ایجاد نشست دوم باخبر نمی‌شود | FIA_USB.2 | ۴ | `internal/platform/security/session` | — | باز | |
| PP-10 | PP | نشست به IP و مرورگر بسته نیست؛ تغییر IP یا User-Agent نشست را نمی‌بندد و تغییر نقش ورود مجدد نمی‌خواهد | FIA_USB.3 | ۴ | نشست | — | باز | R1-13 هم انقضای نشست پس از تغییر تنظیم را می‌خواهد. |
| PP-11 | PP | پس از ورود، آخرین ورود موفق، آخرین تلاش ناموفق و تعداد تلاش‌های ناموفق نشان داده نمی‌شود | FTA_TAH.1، FTA_TAH.2 | ۴ | ورود | — | باز | |
| PP-12 | PP | منع نشست بر اساس IP یا محدوده IP وجود ندارد | FTA_TSE.1 | ۴ | ساعات کاری در `loginguard` | — | باز | محدودیت ساعت و روز هست. محدودیت IP نیست. |
| PP-13 | PP | TLS 1.2 به بالا، رد نگارش‌های قدیمی، مجموعه رمز مجاز، HSTS و اتصال رمزشده به SQL Server مشخص نشده است | FTP_TRP، FCS_TLSS، FPT_ITT | ۶ | استقرار IIS یا سرور Go | — | باز | راهنمای نصب هنوز این بخش را ندارد. |
| PP-14 | PP | پیکربندی امن تولید ناقص است: چاپ رمز در seed، صفحه خطا بدون افشای نسخه، و خاموش‌بودن حالت debug در Gin | — | ۶ | `cmd/seed-admin`، `cmd/server` | — | باز | چاپ رمز seed در D-09 تا فاز ۶ مانده است. |
| PP-15 | PP | شبیه‌ساز کارتخوان همیشه تأیید می‌کند و نباید در build تولیدی باشد | — | ۶ | `internal/cardreader` | — | باز | تحویل امن. |
| PP-16 | PP | محصول نسخه یکتا ندارد و نسخه فقط به مدیر نشان داده نمی‌شود | ALC_CMC.1 | ۶ | `frontend/package.json` | — | باز | نسخه فرانت `0.0.0` است. |
| PP-17 | PP | به‌روزرسانی دستی با درهم‌ساز منتشرشده وجود ندارد | FPT_TUD | ۶ | — | — | باز | |
| PP-18 | PP | `xlsx@0.18.5` آسیب‌پذیری شناخته‌شده دارد و دیگر نگهداری نمی‌شود | AVA_VAN.1 | ۵ | `frontend/package.json` | — | باز | نقشه راه فونت PDF از CDN را هم آورده. فایل فونت محلی است و کمبود `@font-face` در D-20 است. |
| PP-19 | PP | `govulncheck` و `npm audit` در CI شکست job را قطع نمی‌کنند | AVA_VAN.1 | ۵ | `.github/workflows/ci.yml` | — | باز | تا جایگزینی xlsx در فاز ۵ `continue-on-error` می‌مانند. ردیف CI خود زیرساخت را جدا نگه می‌دارد. |
