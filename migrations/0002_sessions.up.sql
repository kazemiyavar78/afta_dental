-- جدول نشست‌های کاربر
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'Sessions')
BEGIN
    CREATE TABLE Sessions (
        Id UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
        Ip NVARCHAR(45) NOT NULL,
        Browser NVARCHAR(512) NOT NULL,
        CreationTime DATETIME2 NOT NULL DEFAULT SYSUTCDATETIME(),
        PersonnelAccountID INT NOT NULL
    );
    CREATE INDEX IX_Sessions_PersonnelAccountID ON Sessions(PersonnelAccountID);
    CREATE INDEX IX_Sessions_CreationTime ON Sessions(CreationTime);
END
