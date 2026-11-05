/* بررسی‌های خواندنی فاز ۱ ممیزی.
   هر سه شمارش باید ۰ باشد.
   فقط SELECT است و afta-check آن را در تراکنش rollback اجرا می‌کند.
   پیشوند هشت‌نویسه‌ای نشست در Details مجاز است و اینجا بررسی نمی‌شود. */

/* رویداد نسخه ۲ که شناسه کاربر دارد ولی نام کاربری فاعل خالی است. */
SELECT COUNT(*) AS HashV2WithUserIdMissingActorUsername
FROM SecurityEvents
WHERE HashVersion = 2
  AND UserID IS NOT NULL
  AND (
      ActorUsername IS NULL
      OR LTRIM(RTRIM(ActorUsername)) = N''
  );

/* پایان نشست بدون دلیل. ردیف قدیمی حذف نشست EventType جدا دارد و اینجا نمی‌آید. */
SELECT COUNT(*) AS SessionTerminatedMissingReason
FROM SecurityEvents
WHERE EventType = N'SESSION_TERMINATED'
  AND (
      Reason IS NULL
      OR LTRIM(RTRIM(Reason)) = N''
  );

/* توضیح رویداد نسخه ۲ که الگوی UUID با گروه‌های ۸-۴-۴-۴-۱۲ دارد. */
SELECT COUNT(*) AS HashV2DescriptionWithUuid
FROM SecurityEvents
WHERE HashVersion = 2
  AND PATINDEX(
      N'%[0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f]-[0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f]-[0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f]-[0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f]-[0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f]%',
      Description
  ) > 0;
