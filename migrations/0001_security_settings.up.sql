-- جدول تنظیمات امنیتی Key-Value
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'SecuritySettings')
BEGIN
    CREATE TABLE SecuritySettings (
        ID INT IDENTITY(1,1) PRIMARY KEY,
        SettingName NVARCHAR(255) NOT NULL UNIQUE,
        SettingValue NVARCHAR(MAX) NOT NULL,
        IntegrityHash NVARCHAR(128) NOT NULL,
        UpdatedAt DATETIME2 NOT NULL DEFAULT SYSUTCDATETIME(),
        UpdatedByUserID INT NULL
    );
END
