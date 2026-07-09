-- نقش‌ها و مجوزها
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'Roles')
BEGIN
    CREATE TABLE Roles (
        ID INT IDENTITY(1,1) PRIMARY KEY,
        Name NVARCHAR(100) NOT NULL UNIQUE,
        Description NVARCHAR(255) NULL,
        IntegrityHash NVARCHAR(128) NOT NULL,
        CreatedAt DATETIME2 NOT NULL DEFAULT SYSUTCDATETIME()
    );
END

IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'Permissions')
BEGIN
    CREATE TABLE Permissions (
        ID INT IDENTITY(1,1) PRIMARY KEY,
        Name NVARCHAR(100) NOT NULL UNIQUE,
        Description NVARCHAR(255) NULL,
        IntegrityHash NVARCHAR(128) NOT NULL,
        CreatedAt DATETIME2 NOT NULL DEFAULT SYSUTCDATETIME()
    );
END

IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'RolePermissions')
BEGIN
    CREATE TABLE RolePermissions (
        RoleID INT NOT NULL,
        PermissionID INT NOT NULL,
        IntegrityHash NVARCHAR(128) NOT NULL,
        PRIMARY KEY (RoleID, PermissionID),
        FOREIGN KEY (RoleID) REFERENCES Roles(ID),
        FOREIGN KEY (PermissionID) REFERENCES Permissions(ID)
    );
END

-- جدول کاربران (PersonnelAccount)
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'Users')
BEGIN
    CREATE TABLE Users (
        ID INT IDENTITY(1,1) PRIMARY KEY,
        Username NVARCHAR(100) NOT NULL UNIQUE,
        PasswordHash NVARCHAR(255) NOT NULL,
        Address NVARCHAR(500) NULL,
        Name NVARCHAR(100) NOT NULL,
        Family NVARCHAR(100) NOT NULL,
        PhoneNumber NVARCHAR(20) NULL,
        MedicalCode NVARCHAR(50) NULL,
        RoleID INT NOT NULL,
        IsActive BIT NOT NULL DEFAULT 1,
        IsLocked BIT NOT NULL DEFAULT 0,
        SecurityCode NVARCHAR(512) NOT NULL,
        IntegrityHash NVARCHAR(128) NOT NULL,
        PasswordChangedAt DATETIME2 NOT NULL DEFAULT SYSUTCDATETIME(),
        LastLoginAt DATETIME2 NULL,
        TwoFactorEnabled BIT NOT NULL DEFAULT 0,
        TwoFactorSecret NVARCHAR(512) NULL,
        CreatedAt DATETIME2 NOT NULL DEFAULT SYSUTCDATETIME(),
        UpdatedAt DATETIME2 NOT NULL DEFAULT SYSUTCDATETIME(),
        FOREIGN KEY (RoleID) REFERENCES Roles(ID)
    );
    CREATE INDEX IX_Users_Username ON Users(Username);
    CREATE INDEX IX_Users_RoleID ON Users(RoleID);
END

-- داده‌های اولیه نقش‌ها
IF NOT EXISTS (SELECT 1 FROM Roles WHERE Name = 'Admin')
    INSERT INTO Roles (Name, Description, IntegrityHash) VALUES ('Admin', N'مدیر سیستم', '');

IF NOT EXISTS (SELECT 1 FROM Roles WHERE Name = 'Doctor')
    INSERT INTO Roles (Name, Description, IntegrityHash) VALUES ('Doctor', N'پزشک', '');

IF NOT EXISTS (SELECT 1 FROM Roles WHERE Name = 'Reception')
    INSERT INTO Roles (Name, Description, IntegrityHash) VALUES ('Reception', N'پذیرش', '');
