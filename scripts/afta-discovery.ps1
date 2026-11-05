# Discovery checks for afta_tpdenta. Run from the repository root.
#
# Required before Q19 (set in this PowerShell process; do not put them in the repo):
#   $env:AFTA_PROBE_ADMIN_USER = '<admin username>'
#   $env:AFTA_PROBE_ADMIN_PASS = '<admin password>'
#
# Optional:
#   $env:AFTA_BASE_URL   = 'http://127.0.0.1:8080'   # default: http://127.0.0.1:$SERVER_PORT from .env, else 8080
#   $env:AFTA_SERVER_LOG = 'C:\path\to\server.log'   # if the server log is outside the repo
#
# Run:
#   powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\afta-discovery.ps1
#
# The script reads .env but never prints secret values. It does not drop tables.
# The only write is creating user afta_probe_<timestamp> and deleting that user at the end.
# An audit row created by the server is left in place (deleting it would break the hash chain).

$ErrorActionPreference = 'Continue'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

$Root = Resolve-Path (Join-Path $PSScriptRoot '..')
Set-Location $Root
. (Join-Path $PSScriptRoot 'lib\redact.ps1')
. (Join-Path $PSScriptRoot 'lib\db.ps1')

$stamp = Get-Date -Format 'yyyyMMdd-HHmm'
$outDir = Join-Path $Root 'docs\afta\check-output'
New-Item -ItemType Directory -Force -Path $outDir | Out-Null
$outFile = Join-Path $outDir ("discovery-{0}.txt" -f $stamp)
$utf8 = New-Object System.Text.UTF8Encoding $false
$writer = New-Object System.IO.StreamWriter($outFile, $false, $utf8)

function Write-Out {
    param([string]$Text = '')
    $writer.WriteLine($Text)
    $writer.Flush()
    Write-Host $Text
}

function Invoke-Section {
    param([string]$Title, [scriptblock]$Body)
    Write-Out ''
    Write-Out ("=== {0} ===" -f $Title)
    try {
        & $Body
    } catch {
        Write-Out ('ERROR: ' + (Redact $_.Exception.Message))
    }
}

function Import-DotEnv {
    param([string]$Path)
    if (-not (Test-Path $Path)) {
        Write-Out 'DOTENV: .env not found'
        return
    }
    $count = 0
    foreach ($line in [System.IO.File]::ReadAllLines($Path)) {
        $trim = $line.Trim()
        if ($trim.Length -eq 0 -or $trim.StartsWith('#')) { continue }
        $idx = $trim.IndexOf('=')
        if ($idx -lt 1) { continue }
        $key = $trim.Substring(0, $idx).Trim()
        $val = $trim.Substring($idx + 1).Trim()
        if ($val.StartsWith('"') -and $val.EndsWith('"') -and $val.Length -ge 2) {
            $val = $val.Substring(1, $val.Length - 2)
        } elseif ($val.StartsWith("'") -and $val.EndsWith("'") -and $val.Length -ge 2) {
            $val = $val.Substring(1, $val.Length - 2)
        }
        $existing = [Environment]::GetEnvironmentVariable($key, 'Process')
        if ([string]::IsNullOrEmpty($existing)) {
            [Environment]::SetEnvironmentVariable($key, $val, 'Process')
        }
        $count++
    }
    Write-Out ("DOTENV: loaded {0} keys (values not printed)" -f $count)
    Write-Out ('DATABASE_URL set: ' + (-not [string]::IsNullOrEmpty($env:DATABASE_URL)))
    Write-Out ('AFTA_PROBE_ADMIN_USER set: ' + (-not [string]::IsNullOrEmpty($env:AFTA_PROBE_ADMIN_USER)))
    Write-Out ('AFTA_PROBE_ADMIN_PASS set: ' + (-not [string]::IsNullOrEmpty($env:AFTA_PROBE_ADMIN_PASS)))
}

function Invoke-AftaNonQuery {
    param([string]$Sql, [hashtable]$Parameters)
    Add-Type -AssemblyName System.Data
    $cs = Convert-AftaConnectionString $env:DATABASE_URL
    $conn = New-Object System.Data.SqlClient.SqlConnection $cs
    $conn.Open()
    try {
        $cmd = $conn.CreateCommand()
        $cmd.CommandText = $Sql
        $cmd.CommandTimeout = 60
        if ($Parameters) {
            foreach ($key in $Parameters.Keys) {
                $value = $Parameters[$key]
                if ($null -eq $value) { $value = [DBNull]::Value }
                [void]$cmd.Parameters.AddWithValue("@$key", $value)
            }
        }
        return $cmd.ExecuteNonQuery()
    } finally {
        $conn.Close()
    }
}

function Invoke-AftaQuery {
    param([string]$Sql, [hashtable]$Parameters)
    Add-Type -AssemblyName System.Data
    $cs = Convert-AftaConnectionString $env:DATABASE_URL
    $conn = New-Object System.Data.SqlClient.SqlConnection $cs
    $conn.Open()
    try {
        $cmd = $conn.CreateCommand()
        $cmd.CommandText = $Sql
        $cmd.CommandTimeout = 60
        if ($Parameters) {
            foreach ($key in $Parameters.Keys) {
                $value = $Parameters[$key]
                if ($null -eq $value) { $value = [DBNull]::Value }
                [void]$cmd.Parameters.AddWithValue("@$key", $value)
            }
        }
        $adapter = New-Object System.Data.SqlClient.SqlDataAdapter $cmd
        $table = New-Object System.Data.DataTable
        [void]$adapter.Fill($table)
        # کاما جلوی باز شدن ردیف‌های DataTable در خروجی PowerShell را می‌گیرد.
        return ,$table
    } finally {
        $conn.Close()
    }
}

function Write-DataTable {
    param($Table)
    if ($null -eq $Table) {
        Write-Out '(no table)'
        return
    }
    if ($Table -isnot [System.Data.DataTable]) {
        Write-Out ("(not a DataTable: {0})" -f $Table.GetType().FullName)
        return
    }
    $cols = @()
    foreach ($c in $Table.Columns) { $cols += $c.ColumnName }
    Write-Out ($cols -join "`t")
    foreach ($row in $Table.Rows) {
        $vals = @()
        foreach ($c in $cols) { $vals += [string]$row[$c] }
        Write-Out ($vals -join "`t")
    }
    Write-Out ("row_count=" + $Table.Rows.Count)
}

function Write-ServerLogTail {
    param([string]$Label)
    Write-Out ("-- server log ({0}) --" -f $Label)
    $files = @()
    if (-not [string]::IsNullOrWhiteSpace($env:AFTA_SERVER_LOG)) {
        $files += $env:AFTA_SERVER_LOG
    }
    foreach ($name in @('server.log', 'app.log', 'afta.log', 'backend.log')) {
        $hit = Get-ChildItem -Path $Root -Filter $name -Recurse -File -ErrorAction SilentlyContinue |
            Where-Object { $_.FullName -notmatch 'node_modules|check-output' }
        foreach ($h in $hit) { $files += $h.FullName }
    }
    $files = $files | Select-Object -Unique
    if (-not $files -or $files.Count -eq 0) {
        Write-Out 'NO_LOG_FILE: no server log in the repo. Set AFTA_SERVER_LOG to the process log if you have one.'
        return
    }
    foreach ($file in $files) {
        Write-Out ("file=" + $file)
        if (-not (Test-Path $file)) {
            Write-Out 'missing'
            continue
        }
        $lines = Get-Content -Path $file -Tail 400 -ErrorAction SilentlyContinue
        $matched = @($lines | Where-Object {
            $_ -match 'پاکسازی|یکپارچگی|Hash-Chain|SecurityEvents|LogCleanup|worker|migrate'
        })
        if ($matched.Count -eq 0) {
            Write-Out 'no worker-related lines in the last 400 lines'
            continue
        }
        $take = $matched | Select-Object -Last 20
        foreach ($line in $take) { Write-Out (Redact $line) }
    }
}

function Invoke-Api {
    param(
        [string]$Method,
        [string]$Url,
        [string]$JsonBody,
        $WebSession,
        [hashtable]$Headers
    )
    $params = @{
        Uri             = $Url
        Method          = $Method
        UseBasicParsing = $true
        TimeoutSec      = 60
    }
    if ($WebSession) { $params.WebSession = $WebSession }
    if ($Headers) { $params.Headers = $Headers }
    if ($JsonBody) {
        $params.ContentType = 'application/json; charset=utf-8'
        $params.Body = [System.Text.Encoding]::UTF8.GetBytes($JsonBody)
    }
    try {
        $resp = Invoke-WebRequest @params
        Write-Out ("HTTP " + [int]$resp.StatusCode)
        Write-Out (Redact $resp.Content)
        return $resp
    } catch {
        $ex = $_.Exception
        $resp = $ex.Response
        if ($resp) {
            $code = [int]$resp.StatusCode
            $body = ''
            try {
                $stream = $resp.GetResponseStream()
                $reader = New-Object System.IO.StreamReader($stream)
                $body = $reader.ReadToEnd()
                $reader.Close()
            } catch {
                $body = $ex.Message
            }
            Write-Out ("HTTP " + $code)
            Write-Out (Redact $body)
        } else {
            Write-Out ('ERROR: ' + (Redact $ex.Message))
        }
        return $null
    }
}

function New-PolicyPassword {
    param([int]$Length, [string]$Upper, [string]$Lower, [string]$Digit, [string]$Special)
    $chars = New-Object System.Collections.Generic.List[char]
    if ($Upper -eq 'true') { $chars.Add([char]'A') }
    if ($Lower -eq 'true') { $chars.Add([char]'a') }
    if ($Digit -eq 'true') { $chars.Add([char]'1') }
    if ($Special -eq 'true') { $chars.Add([char]'!') }
    if ($chars.Count -eq 0) { $chars.Add([char]'a') }
    while ($chars.Count -lt $Length) { $chars.Add([char]'b') }
    return -join $chars
}

function Get-SettingValue {
    param([string]$Name, [string]$Fallback)
    $table = Invoke-AftaQuery -Sql "SELECT SettingValue FROM SecuritySettings WHERE SettingName = @Name" -Parameters @{ Name = $Name }
    if ($table.Rows.Count -eq 0) { return $Fallback }
    return [string]$table.Rows[0]['SettingValue']
}

Write-Out ("discovery start " + (Get-Date -Format 'yyyy-MM-dd HH:mm:ss'))
Write-Out ("root=" + $Root)
Import-DotEnv (Join-Path $Root '.env')

Invoke-Section 'Q2' {
    Write-Out '-- go test ./... --'
    & go test ./... -count=1 2>&1 | ForEach-Object { Write-Out (Redact ([string]$_)) }
    Write-Out ("go_test_exit=" + $LASTEXITCODE)
    Write-Out '-- npm test --'
    Push-Location (Join-Path $Root 'frontend')
    try {
        & npm test 2>&1 | ForEach-Object { Write-Out (Redact ([string]$_)) }
        Write-Out ("npm_test_exit=" + $LASTEXITCODE)
    } finally {
        Pop-Location
    }
}

Invoke-Section 'Q6' {
    Write-Out '-- tables --'
    $tables = Invoke-AftaQuery -Sql "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE = 'BASE TABLE' ORDER BY TABLE_NAME"
    Write-DataTable $tables
    Write-Out '-- columns --'
    $cols = Invoke-AftaQuery -Sql @"
SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE, CHARACTER_MAXIMUM_LENGTH, IS_NULLABLE, ORDINAL_POSITION
FROM INFORMATION_SCHEMA.COLUMNS
ORDER BY TABLE_NAME, ORDINAL_POSITION
"@
    Write-DataTable $cols
    Write-Out '-- schema_migrations --'
    $exists = Invoke-AftaQuery -Sql "SELECT CASE WHEN OBJECT_ID(N'schema_migrations', N'U') IS NULL THEN 'MISSING' ELSE 'EXISTS' END AS status"
    Write-DataTable $exists
    if ([string]$exists.Rows[0]['status'] -eq 'EXISTS') {
        $vers = Invoke-AftaQuery -Sql "SELECT * FROM schema_migrations"
        Write-Out 'schema_migrations columns (values of version only if that column exists):'
        $names = @()
        foreach ($c in $vers.Columns) { $names += $c.ColumnName }
        Write-Out ("columns=" + ($names -join ','))
        if ($names -contains 'version') {
            foreach ($row in $vers.Rows) { Write-Out ("version=" + $row['version']) }
        } else {
            Write-Out 'version column not present; row values omitted'
        }
    }
}

Invoke-Section 'Q7' {
    $st = Invoke-AftaQuery -Sql "SELECT CASE WHEN OBJECT_ID(N'LogCleanupSummaries', N'U') IS NULL THEN 'MISSING' ELSE 'EXISTS' END AS status"
    Write-DataTable $st
    if ([string]$st.Rows[0]['status'] -eq 'EXISTS') {
        $cc = Invoke-AftaQuery -Sql @"
SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_NAME = 'LogCleanupSummaries'
ORDER BY ORDINAL_POSITION
"@
        Write-DataTable $cc
        $n = Invoke-AftaQuery -Sql "SELECT COUNT(*) AS n FROM LogCleanupSummaries"
        Write-DataTable $n
    }
    Write-ServerLogTail 'Q7'
}

Invoke-Section 'Q9' {
    $byType = Invoke-AftaQuery -Sql "SELECT EventType, COUNT(*) AS n FROM SecurityEvents GROUP BY EventType ORDER BY n DESC"
    Write-DataTable $byType
}

Invoke-Section 'Q10' {
    $users = Invoke-AftaQuery -Sql @"
SELECT
  COUNT(*) AS total,
  SUM(CASE WHEN UserID IS NULL THEN 1 ELSE 0 END) AS empty_user_id,
  SUM(CASE WHEN UserID IS NOT NULL THEN 1 ELSE 0 END) AS numeric_user_id,
  SUM(CASE WHEN UserID = 0 THEN 1 ELSE 0 END) AS user_id_zero,
  SUM(CASE WHEN UserID = 1 THEN 1 ELSE 0 END) AS user_id_one
FROM SecurityEvents
"@
    Write-DataTable $users
}

Invoke-Section 'Q16' {
    $settings = Invoke-AftaQuery -Sql "SELECT SettingName, SettingValue FROM SecuritySettings ORDER BY SettingName"
    Write-DataTable $settings
}

Invoke-Section 'Q18' {
    $count = Invoke-AftaQuery -Sql "SELECT COUNT(*) AS row_count FROM SecurityEvents"
    Write-DataTable $count
    $size = Invoke-AftaQuery -Sql @"
SELECT
  SUM(reserved_page_count) * 8.0 / 1024.0 AS reserved_kb,
  SUM(used_page_count) * 8.0 / 1024.0 AS used_kb
FROM sys.dm_db_partition_stats
WHERE object_id = OBJECT_ID(N'SecurityEvents')
"@
    Write-DataTable $size
}

Invoke-Section 'Q19' {
    Write-Out '-- password settings --'
    $pwd = Invoke-AftaQuery -Sql @"
SELECT SettingName, SettingValue
FROM SecuritySettings
WHERE SettingName IN (
  'Password_MinLength',
  'Password_RequireUppercase',
  'Password_RequireLowercase',
  'Password_RequireDigit',
  'Password_RequireSpecial'
)
ORDER BY SettingName
"@
    Write-DataTable $pwd

    $minStr = Get-SettingValue 'Password_MinLength' '8'
    $reqUpper = Get-SettingValue 'Password_RequireUppercase' 'true'
    $reqLower = Get-SettingValue 'Password_RequireLowercase' 'true'
    $reqDigit = Get-SettingValue 'Password_RequireDigit' 'true'
    $reqSpecial = Get-SettingValue 'Password_RequireSpecial' 'true'
    $minLen = 0
    [void][int]::TryParse($minStr, [ref]$minLen)
    if ($minLen -le 0) { $minLen = 8 }
    Write-Out ("parsed_min_length=" + $minLen)

    $role = Invoke-AftaQuery -Sql "SELECT TOP 1 ID, Name FROM Roles ORDER BY ID"
    Write-DataTable $role
    if ($role.Rows.Count -eq 0) {
        Write-Out 'NO_ROLE: skipped API create'
        return
    }
    $roleId = [int]$role.Rows[0]['ID']

    if ([string]::IsNullOrEmpty($env:AFTA_PROBE_ADMIN_USER) -or [string]::IsNullOrEmpty($env:AFTA_PROBE_ADMIN_PASS)) {
        Write-Out 'SKIP_API: set AFTA_PROBE_ADMIN_USER and AFTA_PROBE_ADMIN_PASS before running'
        return
    }

    $port = if ([string]::IsNullOrWhiteSpace($env:SERVER_PORT)) { '8080' } else { $env:SERVER_PORT }
    $base = if ([string]::IsNullOrWhiteSpace($env:AFTA_BASE_URL)) { "http://127.0.0.1:$port" } else { $env:AFTA_BASE_URL.TrimEnd('/') }
    Write-Out ("base_url=" + $base)

    $probeUser = 'afta_probe_' + (Get-Date -Format 'yyyyMMddHHmmss')
    Write-Out ("probe_username=" + $probeUser)

    $session = New-Object Microsoft.PowerShell.Commands.WebRequestSession
    $loginUser = ($env:AFTA_PROBE_ADMIN_USER -replace '\\','\\' -replace '"','\"')
    $loginPass = ($env:AFTA_PROBE_ADMIN_PASS -replace '\\','\\' -replace '"','\"')
    $loginJson = '{"username":"' + $loginUser + '","password":"' + $loginPass + '"}'
    Write-Out '-- login --'
    $loginResp = Invoke-Api -Method POST -Url ($base + '/api/login') -JsonBody $loginJson -WebSession $session
    if (-not $loginResp) {
        Write-Out 'login failed; create skipped'
        return
    }

    $uri = [Uri]($base + '/')
    $csrf = $null
    foreach ($cookie in $session.Cookies.GetCookies($uri)) {
        if ($cookie.Name -eq 'csrf_token') { $csrf = $cookie.Value }
    }
    if ([string]::IsNullOrEmpty($csrf)) {
        Write-Out 'CSRF cookie missing; create skipped'
        return
    }
    Write-Out 'csrf_cookie=present'

    function Invoke-ProbeCreate {
        param(
            [int]$Length,
            [string]$Upper,
            [string]$Lower,
            [string]$Digit,
            [string]$Special
        )
        $password = New-PolicyPassword -Length $Length -Upper $Upper -Lower $Lower -Digit $Digit -Special $Special
        $bodyObj = @{
            username     = $probeUser
            password     = $password
            name         = 'probe'
            family       = 'discovery'
            address      = ''
            phone_number = '09120000000'
            user_type    = 'Doctor'
            role_id      = $roleId
        }
        $json = ($bodyObj | ConvertTo-Json -Compress)
        Write-Out ("-- create password_len=" + $Length + " user_type_sent=Doctor role_id=" + $roleId + " --")
        Invoke-Api -Method POST -Url ($base + '/api/users') -JsonBody $json -WebSession $session -Headers @{ 'X-CSRF-Token' = $csrf } | Out-Null
    }

    if ($minLen -gt 12) {
        Invoke-ProbeCreate -Length 12 -Upper $reqUpper -Lower $reqLower -Digit $reqDigit -Special $reqSpecial
    }
    Invoke-ProbeCreate -Length $minLen -Upper $reqUpper -Lower $reqLower -Digit $reqDigit -Special $reqSpecial

    Write-Out '-- cleanup probe user --'
    if ($probeUser -notlike 'afta_probe_*') {
        Write-Out 'REFUSING cleanup: username is not an afta_probe_ user'
    } else {
        try {
            $nSess = Invoke-AftaNonQuery -Sql "DELETE FROM Sessions WHERE PersonnelAccountID IN (SELECT ID FROM Users WHERE Username = @Username)" -Parameters @{ Username = $probeUser }
            Write-Out ("sessions_deleted=" + $nSess)
        } catch {
            Write-Out ('sessions_delete_error: ' + (Redact $_.Exception.Message))
        }
        try {
            $nUser = Invoke-AftaNonQuery -Sql "DELETE FROM Users WHERE Username = @Username" -Parameters @{ Username = $probeUser }
            Write-Out ("users_deleted=" + $nUser)
        } catch {
            Write-Out ('users_delete_error: ' + (Redact $_.Exception.Message))
        }
    }
    Write-Out 'audit rows for this create were not deleted'
    Write-ServerLogTail 'Q19'
}

Write-Out ''
Write-Out ("discovery end " + (Get-Date -Format 'yyyy-MM-dd HH:mm:ss'))
Write-Out ("output=" + $outFile)
$writer.Flush()
$writer.Close()
Write-Host ("OUTPUT " + $outFile)
