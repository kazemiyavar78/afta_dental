-- جدول رویدادهای امنیتی با Hash-Chain
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'SecurityEvents')
BEGIN
    CREATE TABLE SecurityEvents (
        ID BIGINT IDENTITY(1,1) PRIMARY KEY,
        UserID INT NULL,
        Ip NVARCHAR(45) NOT NULL,
        EventType NVARCHAR(100) NOT NULL,
        Description NVARCHAR(MAX) NOT NULL,
        CreatedAt DATETIME2 NOT NULL DEFAULT SYSUTCDATETIME(),
        PrevHash NVARCHAR(128) NOT NULL DEFAULT '',
        RowHash NVARCHAR(128) NOT NULL
    );
    CREATE INDEX IX_SecurityEvents_CreatedAt ON SecurityEvents(CreatedAt);
END
