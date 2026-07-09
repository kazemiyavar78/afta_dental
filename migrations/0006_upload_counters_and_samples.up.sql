-- شمارنده آپلود اتمیک (تسک ۵)
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'UploadCounters')
BEGIN
    CREATE TABLE UploadCounters (
        UserID INT NOT NULL,
        WindowStart DATETIME2 NOT NULL,
        UploadCount INT NOT NULL DEFAULT 0,
        PRIMARY KEY (UserID, WindowStart),
        FOREIGN KEY (UserID) REFERENCES Users(ID)
    );
END

-- جدول پذیرش (نمونه)
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'Receptions')
BEGIN
    CREATE TABLE Receptions (
        ID INT IDENTITY(1,1) PRIMARY KEY,
        PatientName NVARCHAR(200) NOT NULL,
        DoctorID INT NOT NULL,
        ReceptionDate DATETIME2 NOT NULL,
        Status NVARCHAR(50) NOT NULL DEFAULT 'pending',
        CreatedAt DATETIME2 NOT NULL DEFAULT SYSUTCDATETIME(),
        FOREIGN KEY (DoctorID) REFERENCES Users(ID)
    );
END

-- سازمان (نمونه)
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'Organizations')
BEGIN
    CREATE TABLE Organizations (
        ID INT IDENTITY(1,1) PRIMARY KEY,
        Name NVARCHAR(200) NOT NULL,
        CreatedAt DATETIME2 NOT NULL DEFAULT SYSUTCDATETIME()
    );
END

-- صندوق (نمونه)
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'Funds')
BEGIN
    CREATE TABLE Funds (
        ID INT IDENTITY(1,1) PRIMARY KEY,
        Name NVARCHAR(200) NOT NULL,
        Balance DECIMAL(18,2) NOT NULL DEFAULT 0,
        CreatedAt DATETIME2 NOT NULL DEFAULT SYSUTCDATETIME()
    );
END

-- تعرفه (نمونه)
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'Tariffs')
BEGIN
    CREATE TABLE Tariffs (
        ID INT IDENTITY(1,1) PRIMARY KEY,
        Name NVARCHAR(200) NOT NULL,
        Amount DECIMAL(18,2) NOT NULL,
        CreatedAt DATETIME2 NOT NULL DEFAULT SYSUTCDATETIME()
    );
END

-- لاگ خلاصه پاکسازی
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'LogCleanupSummaries')
BEGIN
    CREATE TABLE LogCleanupSummaries (
        ID INT IDENTITY(1,1) PRIMARY KEY,
        TableName NVARCHAR(100) NOT NULL,
        DeletedCount INT NOT NULL,
        FromDate DATETIME2 NULL,
        ToDate DATETIME2 NULL,
        CreatedAt DATETIME2 NOT NULL DEFAULT SYSUTCDATETIME()
    );
END
