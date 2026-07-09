-- جدول تلاش‌های ناموفق ورود (ذخیره در دیتابیس، نه کش حافظه)
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'LoginAttempts')
BEGIN
    CREATE TABLE LoginAttempts (
        Ip NVARCHAR(45) PRIMARY KEY,
        FailedCount INT NOT NULL DEFAULT 0,
        LastAttemptAt DATETIME2 NOT NULL DEFAULT SYSUTCDATETIME(),
        LockedUntil DATETIME2 NULL
    );
    CREATE INDEX IX_LoginAttempts_LastAttemptAt ON LoginAttempts(LastAttemptAt);
END
