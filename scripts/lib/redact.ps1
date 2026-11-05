# Shared redaction for scripts/afta-discovery.ps1 and scripts/afta-check.ps1.
# Dot-source this file; do not run it directly.

# متن را برای چاپ امن می‌کند: رمز، رشته اتصال، session و CSRF را با *** عوض می‌کند.
# ورودی متن خام است و خروجی همان متن بدون مقدارهای حساس است.
function Redact {
    param([string]$Text)
    if ([string]::IsNullOrEmpty($Text)) { return '' }
    $s = $Text
    # کل URL اتصال sqlserver/mssql تا کاربر، رمز و میزبان در خروجی نماند.
    $s = [regex]::Replace($s, '(?i)\b(?:sqlserver|mssql)://\S+', '***')
    $s = [regex]::Replace($s, '(?i)(password|pwd|pass)\s*=\s*[^;\s''"]+', '$1=***')
    $s = [regex]::Replace($s, '(?i)(password|pwd|pass)\s*=\s*"[^"]*"', '$1=***')
    $s = [regex]::Replace($s, "(?i)(password|pwd|pass)\s*=\s*'[^']*'", '$1=***')
    $s = [regex]::Replace($s, '(?i)(user id|uid)\s*=\s*[^;\s''"]+', '$1=***')
    $s = [regex]::Replace($s, '(?i)(Data Source|Server|Address|Addr|Network Address|Initial Catalog|Database)\s*=\s*[^;''\s]+', '$1=***')
    $s = [regex]::Replace($s, '://[^@\s/]+@', '://***@')
    $s = [regex]::Replace($s, '(?i)"(session_id|sessionId|csrf_token|password|two_factor_secret)"\s*:\s*"[^"]*"', '"$1":"***"')
    $s = [regex]::Replace($s, '(?i)(X-CSRF-Token|csrf_token|session_id|sessionId)\s*[:=]\s*\S+', '$1=***')
    # اگر همان خط دربارهٔ نشست است، UUID را هم ماسک کن.
    if ($s -match '(?i)session') {
        $s = [regex]::Replace($s, '(?i)\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b', '***')
    }
    return $s
}
