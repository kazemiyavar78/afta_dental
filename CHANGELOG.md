# تغییرات

قالب این فایل [Keep a Changelog](https://keepachangelog.com/fa/1.1.0/) است. این محصول هنوز نسخه منتشرشده ندارد.

## [منتشرنشده]

### اضافه‌شده

- test(platform): زیرساخت تست مشترک (`internal/platform/testutil`)، تست integration برای AutoMigrate، helper فرانت `renderWithProviders` و اسکریپت `scripts/afta-check.ps1`. منطق برنامه عوض نشده است. وضعیت تست پایگاه: منتظر اجرای کاربر.
- ci: GitHub Actions برای `go vet`، تست واحد Go، `govulncheck`، و تست و build فرانت بدون پایگاه داده. `govulncheck` و `npm audit` تا فاز ۵ blocking نیستند چون `xlsx@0.18.5` آسیب‌پذیری شناخته‌شده دارد و باید در آن فاز جایگزین و سپس blocking شود.
- docs(afta): نقشه راه به `docs/afta/roadmap.md` منتقل شد. `docs/afta/traceability.md` ایرادهای R1-01 تا R1-24، یافته‌های D-11 به بعد و شکاف‌های PP را با همان ستون‌ها دارد. اسکلت `user-guide.md` و `install-guide.md` و جدول `settings.md` اضافه شد. README با نسخه Go و قرارداد env و تست هم‌خوان شد.

### اصلاح‌شده

- fix(git): الگوی `logs` در `.gitignore` و `frontend/.gitignore` به `/logs/` محدود شد تا فقط پوشه logs کنار همان فایل نادیده گرفته شود. `internal/logs` و `frontend/src/modules/logs` که به‌خاطر الگوی قبلی هرگز commit نشده بودند به مخزن برگشتند (D-10). `*.log` همچنان نادیده گرفته می‌شود. `CheckWorkHours` دیگر مقدار ساعت شروع و پایان را در stdout چاپ نمی‌کند (D-09). `scripts/afta-discovery.ps1` ماسک و اتصال پایگاه را از `scripts/lib` می‌گیرد، پارامتر تهی را `DBNull` می‌فرستد و نتیجه پرس‌وجو را به‌صورت یک `DataTable` برمی‌گرداند. وضعیت D-01: انجام‌شده — منتظر مشاور؛ خروجی discovery ساعت ۲۱:۲۳ وجود `LogCleanupSummaries` و `UploadCounters` را در پایگاه تأیید کرد.
- fix(platform): خطای bind ورودی JSON و query به‌جای HTTP 500 عمومی، پاسخ ۴۰۰ با کد `VALIDATION_ERROR` و پیام «اطلاعات ورودی نامعتبر یا ناقص است.» می‌دهد. نام فیلد و مقدار در پاسخ نیست و مقدار فیلد در لاگ سرور نیست (D-08). `UpdateSettingByID` دیگر نام و مقدار تنظیم را در stdout چاپ نمی‌کند؛ خطای ذخیره فقط نام کلید و خطا را لاگ می‌کند (D-09).
- fix(migrate): AutoMigrate تنها مسیر schema است. `LogCleanupSummary` و `UploadCounter` به فهرست راه‌اندازی اضافه شدند. خطای هر جدول با نام جدول برمی‌گردد و سرور متوقف می‌شود (FPT_FLS.1، D-01). فایل‌های `migrations/` حذف شدند چون هرگز اجرا نشده بودند. `Funds.Balance` همچنان float است و به فاز ۵ موکول شد (D-02). وضعیت تست پایگاه: منتظر اجرای کاربر.
