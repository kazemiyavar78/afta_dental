/* وضعیت زنجیره ممیزی.
   تعداد رویداد به تفکیک نوع، تعداد پیوندهای شکسته PrevHash با LAG، و ردیف AuditChainStates.
   فقط SELECT است و afta-check آن را در تراکنش rollback اجرا می‌کند. */

SELECT EventType, COUNT(*) AS EventCount
FROM SecurityEvents
GROUP BY EventType
ORDER BY EventCount DESC;

SELECT COUNT(*) AS BrokenLinkCount
FROM (
    SELECT
        PrevHash,
        LAG(RowHash) OVER (ORDER BY ID) AS PrevRowHash
    FROM SecurityEvents
) AS links
WHERE PrevRowHash IS NOT NULL
  AND PrevHash <> PrevRowHash;

SELECT
    ID,
    LastVerifiedID,
    LastVerifiedHash,
    AnchorID,
    AnchorHash,
    LastCheckAt,
    LastResult,
    BrokenAtID,
    IntegrityHash
FROM AuditChainStates;
