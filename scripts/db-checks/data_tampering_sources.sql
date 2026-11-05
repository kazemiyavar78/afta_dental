/* منبع رویدادهای DATA_TAMPERING.
   گروه‌بندی بر اساس ۶۰ نویسه اول توضیح، با تعداد و اولین و آخرین زمان.
   فقط SELECT است و afta-check آن را در تراکنش rollback اجرا می‌کند.
   هدف این است که معلوم شود هشدارهای تکراری قدیمی از کدام بررسی آمده‌اند. */

SELECT TOP (20)
    LEFT(Description, 60) AS DescriptionPrefix,
    COUNT(*) AS EventCount,
    MIN(CreatedAt) AS FirstCreatedAt,
    MAX(CreatedAt) AS LastCreatedAt
FROM SecurityEvents
WHERE EventType = 'DATA_TAMPERING'
GROUP BY LEFT(Description, 60)
ORDER BY EventCount DESC;
