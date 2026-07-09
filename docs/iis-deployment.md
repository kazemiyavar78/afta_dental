# استقرار پشت IIS (Windows Server 2019)

## معماری

```
[Client HTTPS] → [IIS + TLS Termination] → [Reverse Proxy] → [Go Windows Service :8080] → [SQL Server]
```

## مراحل خلاصه

### ۱. کامپایل باینری Go

```powershell
go build -o afta-reception.exe ./cmd/server
```

### ۲. نصب به‌عنوان Windows Service

از ابزارهایی مثل `nssm` یا `sc create` استفاده کنید:

```powershell
nssm install AftaReception "C:\apps\afta-reception\afta-reception.exe"
nssm set AftaReception AppDirectory "C:\apps\afta-reception"
nssm set AftaReception AppEnvironmentExtra DATABASE_URL=... AES_KEY=...
nssm start AftaReception
```

### ۳. تنظیم IIS Reverse Proxy

1. نصب **URL Rewrite** و **Application Request Routing (ARR)**
2. ایجاد سایت HTTPS با گواهی TLS
3. قانون Reverse Proxy:

```xml
<rule name="ReverseProxyToGo" stopProcessing="true">
  <match url="(.*)" />
  <action type="Rewrite" url="http://localhost:8080/{R:1}" />
  <serverVariables>
    <set name="HTTP_X_FORWARDED_PROTO" value="https" />
  </serverVariables>
</rule>
```

### ۴. هدرهای امنیتی

Go خود هدرهای امنیتی را ست می‌کند. IIS نیز می‌تواند HSTS اضافه کند.

### ۵. کوکی‌ها

- `session_id`: HttpOnly + Secure + SameSite=Strict
- `csrf_token`: Secure + SameSite=Strict (غیر HttpOnly)

در محیط production، `SECURE_COOKIES=true` تنظیم شود.

## نکات

- Connection String و کلیدها فقط در env سرویس Windows، نه در web.config
- لاگ‌های IIS و Go جدا نگهداری شوند
- SQL Server روی همان سرور یا شبکه داخلی
