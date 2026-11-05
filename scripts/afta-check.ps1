# بررسی یکجای افتا. از ریشه مخزن اجرا کنید (اسکریپت خودش به ریشه برمی‌گردد):
#   powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\afta-check.ps1
#
# تست integration و SQLهای scripts/db-checks فقط وقتی به پایگاه وصل می‌شوند که
# AFTA_TEST_DATABASE_URL ست باشد، یا کاربر یک بار تأیید کند که DATABASE_URL
# داخل .env پایگاه توسعه است. مقدار اتصال چاپ نمی‌شود. production استفاده نمی‌شود.
# شکست یک بخش، بخش‌های بعدی را متوقف نمی‌کند.

$ErrorActionPreference = 'Continue'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

$Root = Resolve-Path (Join-Path $PSScriptRoot '..')
Set-Location $Root
. (Join-Path $PSScriptRoot 'lib\redact.ps1')
. (Join-Path $PSScriptRoot 'lib\db.ps1')

$stamp = Get-Date -Format 'yyyyMMdd-HHmm'
$outDir = Join-Path $Root 'docs\afta\check-output'
New-Item -ItemType Directory -Force -Path $outDir | Out-Null
$outFile = Join-Path $outDir ("check-{0}.txt" -f $stamp)
$utf8 = New-Object System.Text.UTF8Encoding $false
$script:Writer = New-Object System.IO.StreamWriter($outFile, $false, $utf8)
$script:Results = New-Object System.Collections.Generic.List[object]
$script:DbAllowed = $false

# یک خط را هم در فایل خروجی و هم در کنسول می‌نویسد. مقدار حساس قبل از نوشتن ماسک می‌شود.
function Write-Out {
    param([string]$Text = '')
    $safe = Redact $Text
    $script:Writer.WriteLine($safe)
    $script:Writer.Flush()
    Write-Host $safe
}

# فایل .env را به جدول محلی برمی‌گرداند و متغیر محیطی فرایند را عوض نمی‌کند.
function Read-DotEnvMap {
    param([string]$Path)
    $map = @{}
    if (-not (Test-Path -LiteralPath $Path)) { return $map }
    foreach ($line in [System.IO.File]::ReadAllLines($Path)) {
        $trim = $line.Trim()
        if ($trim.Length -eq 0 -or $trim.StartsWith('#')) { continue }
        $idx = $trim.IndexOf('=')
        if ($idx -lt 1) { continue }
        $key = $trim.Substring(0, $idx).Trim()
        $val = $trim.Substring($idx + 1).Trim()
        if ((($val.StartsWith('"') -and $val.EndsWith('"')) -or ($val.StartsWith("'") -and $val.EndsWith("'"))) -and $val.Length -ge 2) {
            $val = $val.Substring(1, $val.Length - 2)
        }
        if (-not $map.ContainsKey($key)) { $map[$key] = $val }
    }
    return $map
}

# نام پایگاه را از URL گو یا رشته ADO می‌خواند. اگر پیدا نشود رشته خالی برمی‌گرداند.
function Get-AftaDatabaseName {
    param([string]$Url)
    if ([string]::IsNullOrWhiteSpace($Url)) { return '' }
    if ($Url -match '(?i)(?:\?|&)(?:database|initial catalog)=([^&]+)') {
        return [Uri]::UnescapeDataString(($Matches[1] -replace '\+', ' ')).Trim()
    }
    if ($Url -match '(?i)(?:Initial Catalog|Database)\s*=\s*([^;]+)') {
        return $Matches[1].Trim()
    }
    return ''
}

# true یعنی این نام نباید هدف تست باشد: خالی، سیستمی، یا شبیه production.
function Test-AftaUnsafeDatabaseName {
    param([string]$Name)
    if ([string]::IsNullOrWhiteSpace($Name)) { return $true }
    if ($Name -match '^(?i)(master|model|msdb|tempdb)$') { return $true }
    if ($Name -match '(?i)(^|[^a-z0-9])(prod|production)([^a-z0-9]|$)') { return $true }
    return $false
}

# true یعنی می‌توان از کاربر سؤال تعاملی پرسید.
function Test-CanPrompt {
    try {
        if ([Console]::IsInputRedirected) { return $false }
    } catch {
        return $false
    }
    return $true
}

# توضیح کوتاه بدون چاپ نام یا رشته اتصال می‌دهد.
function Get-UnsafeDatabaseReason {
    param([string]$Name)
    if ([string]::IsNullOrWhiteSpace($Name)) { return 'نام پایگاه در اتصال مشخص نیست' }
    if ($Name -match '^(?i)(master|model|msdb|tempdb)$') { return 'پایگاه سیستمی SQL Server مجاز نیست' }
    return 'پایگاه شبیه production است و استفاده نمی‌شود'
}

# اگر AFTA_TEST_DATABASE_URL نباشد، یک بار برای پایگاه توسعه .env اجازه می‌گیرد.
function Resolve-TestDatabase {
    $explicit = [Environment]::GetEnvironmentVariable('AFTA_TEST_DATABASE_URL', 'Process')
    if (-not [string]::IsNullOrWhiteSpace($explicit)) {
        $name = Get-AftaDatabaseName $explicit
        if (Test-AftaUnsafeDatabaseName $name) {
            Write-Out ('AFTA_TEST_DATABASE_URL کنار گذاشته شد: ' + (Get-UnsafeDatabaseReason $name))
            Remove-Item Env:\AFTA_TEST_DATABASE_URL -ErrorAction SilentlyContinue
            $script:DbAllowed = $false
            return
        }
        $script:DbAllowed = $true
        Write-Out 'AFTA_TEST_DATABASE_URL از محیط گرفته شد. مقدار چاپ نمی‌شود.'
        return
    }

    $candidate = [Environment]::GetEnvironmentVariable('DATABASE_URL', 'Process')
    $source = 'محیط فرایند'
    if ([string]::IsNullOrWhiteSpace($candidate)) {
        $dotenv = Read-DotEnvMap (Join-Path $Root '.env')
        if ($dotenv.ContainsKey('DATABASE_URL')) {
            $candidate = [string]$dotenv['DATABASE_URL']
            $source = '.env'
        }
    }
    if ([string]::IsNullOrWhiteSpace($candidate)) {
        Write-Out 'AFTA_TEST_DATABASE_URL و DATABASE_URL خالی‌اند. بررسی پایگاه رد می‌شود.'
        $script:DbAllowed = $false
        return
    }

    $name = Get-AftaDatabaseName $candidate
    if (Test-AftaUnsafeDatabaseName $name) {
        Write-Out ('DATABASE_URL استفاده نمی‌شود: ' + (Get-UnsafeDatabaseReason $name))
        $script:DbAllowed = $false
        return
    }
    if (-not (Test-CanPrompt)) {
        Write-Out 'پایگاه توسعه پیدا شد ولی ورودی تعاملی نیست. AFTA_TEST_DATABASE_URL را ست کنید.'
        $script:DbAllowed = $false
        return
    }

    Write-Out ('AFTA_TEST_DATABASE_URL ست نیست. DATABASE_URL از {0} خوانده شد و چاپ نمی‌شود.' -f $source)
    Write-Host 'آیا تست integration و بررسی DB روی همین پایگاه توسعه اجرا شود؟'
    Write-Host 'هرگز روی production اجرا نکنید. y بله، n خیر:'
    $answer = ''
    try {
        $answer = Read-Host
    } catch {
        $answer = 'n'
    }
    if ($answer -match '^(?i)\s*(y|yes|بله)\s*$') {
        $env:AFTA_TEST_DATABASE_URL = $candidate
        $script:DbAllowed = $true
        Write-Out 'پایگاه توسعه برای همین اجرا انتخاب شد.'
        return
    }
    Write-Out 'اجرای بررسی پایگاه توسعه رد شد.'
    $script:DbAllowed = $false
}

# روی ویندوز npm.cmd قابل‌اعتمادتر از شیم پاورشل است. اگر نباشد همان npm برمی‌گردد.
function Get-NpmCommand {
    $cmd = Get-Command npm.cmd -ErrorAction SilentlyContinue
    if ($cmd) { return $cmd.Source }
    return 'npm'
}

# دستور خارجی را اجرا می‌کند، خروجی را ماسک می‌کند و کد خروج را برمی‌گرداند.
function Invoke-Native {
    param(
        [Parameter(Mandatory = $true)][string]$File,
        [string[]]$ArgumentList,
        [Parameter(Mandatory = $true)][string]$WorkingDirectory
    )
    Push-Location $WorkingDirectory
    try {
        & $File @ArgumentList 2>&1 | ForEach-Object { Write-Out ([string]$_) }
        if ($null -eq $LASTEXITCODE) { return 0 }
        return [int]$LASTEXITCODE
    } finally {
        Pop-Location
    }
}

# نتیجه پاس/شکست یک دستور را به شکل وضعیت بخش برمی‌گرداند.
function Get-CommandStatus {
    param([int]$Code, [string]$Note)
    if ($Code -eq 0) {
        return @{ Exit = 0; Status = 'PASS'; Note = $Note }
    }
    return @{ Exit = $Code; Status = 'FAIL'; Note = $Note }
}

# کامنت و رشتهٔ نقل‌قولی را برمی‌دارد تا جستجوی کلیدواژهٔ نوشتنی اشتباه نگیرد.
function Get-SqlBareText {
    param([string]$Sql)
    $s = [regex]::Replace($Sql, '(?s)/\*.*?\*/', ' ')
    $s = [regex]::Replace($s, '(?m)--.*?$', ' ')
    $s = [regex]::Replace($s, "'(?:''|[^'])*'", "''")
    return $s
}

# true یعنی متن فقط کوئری خواندنی است.
function Test-SqlReadOnly {
    param([string]$BareSql)
    $forbidden = 'insert|update|delete|drop|alter|merge|truncate|exec|execute|create|grant|revoke|deny|backup|restore|dbcc|into|xp_|openrowset|opendatasource|bulk'
    if ($BareSql -match "(?i)\b($forbidden)\b") { return $false }
    if ($BareSql -notmatch '(?i)\bselect\b') { return $false }
    return $true
}

# یک SELECT را در تراکنش می‌خواند و تراکنش را برمی‌گرداند. رشته اتصال چاپ نمی‌شود.
function Invoke-AftaSelect {
    param([string]$ConnectionString, [string]$Sql, [string]$Name)
    Add-Type -AssemblyName System.Data
    $conn = New-Object System.Data.SqlClient.SqlConnection $ConnectionString
    $conn.Open()
    try {
        $tran = $conn.BeginTransaction()
        try {
            $cmd = $conn.CreateCommand()
            $cmd.Transaction = $tran
            $cmd.CommandTimeout = 60
            $cmd.CommandText = $Sql
            $adapter = New-Object System.Data.SqlClient.SqlDataAdapter $cmd
            $set = New-Object System.Data.DataSet
            [void]$adapter.Fill($set)
            Write-Out ("result_sets={0} file={1}" -f $set.Tables.Count, $Name)
            foreach ($table in $set.Tables) {
                $cols = @()
                foreach ($c in $table.Columns) { $cols += $c.ColumnName }
                Write-Out ($cols -join "`t")
                foreach ($row in $table.Rows) {
                    $vals = @()
                    foreach ($c in $cols) { $vals += [string]$row[$c] }
                    Write-Out ($vals -join "`t")
                }
                Write-Out ("row_count=" + $table.Rows.Count)
            }
        } finally {
            $tran.Rollback()
        }
    } finally {
        $conn.Close()
    }
}

# فایل SQL را به بچ‌های GO تقسیم می‌کند و فقط بچ خواندنی را اجرا می‌کند.
function Invoke-AftaReadOnlyFile {
    param([string]$Path, [string]$ConnectionString)
    $raw = [System.IO.File]::ReadAllText($Path)
    $batches = [regex]::Split($raw, '(?m)^\s*GO\s*$')
    $ran = 0
    foreach ($batch in $batches) {
        $bare = Get-SqlBareText $batch
        if ([string]::IsNullOrWhiteSpace($bare)) { continue }
        if (-not (Test-SqlReadOnly $bare)) {
            throw ('فقط SELECT خواندنی مجاز است: ' + [System.IO.Path]::GetFileName($Path))
        }
        Invoke-AftaSelect -ConnectionString $ConnectionString -Sql $batch -Name ([System.IO.Path]::GetFileName($Path))
        $ran++
    }
    if ($ran -eq 0) {
        throw ('فایل SQL دستور SELECT ندارد: ' + [System.IO.Path]::GetFileName($Path))
    }
}

# یک بخش را اجرا می‌کند. خطا بخش‌های بعدی را متوقف نمی‌کند.
function Invoke-CheckSection {
    param([string]$Title, [scriptblock]$Action)
    Write-Out ''
    Write-Out ('=== {0} ===' -f $Title)
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $exit = 1
    $status = 'FAIL'
    $note = ''
    try {
        $result = & $Action
        if ($result -is [System.Collections.IDictionary]) {
            $exit = [int]$result['Exit']
            $status = [string]$result['Status']
            $note = [string]$result['Note']
        } else {
            $note = 'section did not return a status'
        }
    } catch {
        $exit = 1
        $status = 'FAIL'
        $note = $_.Exception.Message
        Write-Out ('ERROR: ' + $note)
    }
    $sw.Stop()
    $elapsed = '{0:hh\:mm\:ss\.fff}' -f $sw.Elapsed
    Write-Out ('elapsed: {0}' -f $elapsed)
    Write-Out ('exit_code: {0}' -f $exit)
    Write-Out ('result: {0}' -f $status)
    if (-not [string]::IsNullOrWhiteSpace($note)) {
        Write-Out ('note: {0}' -f $note)
    }
    $script:Results.Add([pscustomobject]@{
        Title  = $Title
        Status = $status
        Exit   = $exit
        Elapsed = $elapsed
        Note   = (Redact $note)
    }) | Out-Null
}

$failedCount = 0
try {
    Write-Out ('afta-check {0}' -f $stamp)
    Write-Out ('root: {0}' -f $Root)
    Resolve-TestDatabase

    Invoke-CheckSection 'GO_VET' {
        $code = Invoke-Native -File 'go' -ArgumentList @('vet', './...') -WorkingDirectory $Root
        return Get-CommandStatus -Code $code -Note ''
    }

    Invoke-CheckSection 'GO_TEST_UNIT' {
        $code = Invoke-Native -File 'go' -ArgumentList @('test', './...', '-count=1') -WorkingDirectory $Root
        return Get-CommandStatus -Code $code -Note ''
    }

    Invoke-CheckSection 'GO_TEST_INTEGRATION' {
        if (-not $script:DbAllowed) {
            Write-Out 'AFTA_TEST_DATABASE_URL خالی است. تست‌هایی که پایگاه می‌خواهند با t.Skip رد می‌شوند.'
        }
        $code = Invoke-Native -File 'go' -ArgumentList @('test', '-tags', 'integration', './...', '-count=1') -WorkingDirectory $Root
        $note = ''
        if (-not $script:DbAllowed) { $note = 'database not confirmed; db tests skip' }
        return Get-CommandStatus -Code $code -Note $note
    }

    Invoke-CheckSection 'GOVULNCHECK' {
        $cmd = Get-Command govulncheck -ErrorAction SilentlyContinue
        if (-not $cmd) {
            Write-Out 'govulncheck نصب نیست.'
            Write-Out 'برای نصب: go install golang.org/x/vuln/cmd/govulncheck@latest'
            return @{ Exit = 0; Status = 'SKIP'; Note = 'not installed' }
        }
        $code = Invoke-Native -File 'govulncheck' -ArgumentList @('./...') -WorkingDirectory $Root
        return Get-CommandStatus -Code $code -Note ''
    }

    Invoke-CheckSection 'FRONT_TEST' {
        $front = Join-Path $Root 'frontend'
        $code = Invoke-Native -File (Get-NpmCommand) -ArgumentList @('test') -WorkingDirectory $front
        return Get-CommandStatus -Code $code -Note ''
    }

    Invoke-CheckSection 'FRONT_BUILD' {
        $front = Join-Path $Root 'frontend'
        $code = Invoke-Native -File (Get-NpmCommand) -ArgumentList @('run', 'build') -WorkingDirectory $front
        return Get-CommandStatus -Code $code -Note ''
    }

    Invoke-CheckSection 'NPM_AUDIT' {
        $front = Join-Path $Root 'frontend'
        $code = Invoke-Native -File (Get-NpmCommand) -ArgumentList @('audit', '--omit=dev') -WorkingDirectory $front
        return Get-CommandStatus -Code $code -Note ''
    }

    Invoke-CheckSection 'DB_CHECKS' {
        if (-not $script:DbAllowed) {
            Write-Out 'بررسی SQL اجرا نشد چون پایگاه توسعه تأیید نشده است.'
            return @{ Exit = 0; Status = 'SKIP'; Note = 'database not confirmed' }
        }
        $cs = ''
        try {
            $cs = Convert-AftaConnectionString $env:AFTA_TEST_DATABASE_URL
        } catch {
            Write-Out ('ERROR: ' + $_.Exception.Message)
            return @{ Exit = 1; Status = 'FAIL'; Note = 'connection string rejected' }
        }
        $dir = Join-Path $Root 'scripts\db-checks'
        $files = @(Get-ChildItem -LiteralPath $dir -Filter '*.sql' -File -ErrorAction SilentlyContinue | Sort-Object Name)
        if ($files.Count -eq 0) {
            Write-Out 'هیچ فایل sql در scripts/db-checks نیست.'
            return @{ Exit = 0; Status = 'PASS'; Note = 'no sql files' }
        }
        $failed = 0
        foreach ($file in $files) {
            Write-Out ('--- {0} ---' -f $file.Name)
            try {
                Invoke-AftaReadOnlyFile -Path $file.FullName -ConnectionString $cs
                Write-Out 'file_result: PASS'
            } catch {
                $failed++
                Write-Out ('file_result: FAIL ' + $_.Exception.Message)
            }
        }
        if ($failed -gt 0) {
            return @{ Exit = 1; Status = 'FAIL'; Note = ('{0} sql file(s) failed' -f $failed) }
        }
        return @{ Exit = 0; Status = 'PASS'; Note = ('{0} sql file(s)' -f $files.Count) }
    }

    Write-Out ''
    Write-Out '=== SUMMARY ==='
    foreach ($row in $script:Results) {
        $line = '{0} {1} exit={2} time={3}' -f $row.Title, $row.Status, $row.Exit, $row.Elapsed
        if (-not [string]::IsNullOrWhiteSpace($row.Note)) {
            $line = $line + ' ' + $row.Note
        }
        Write-Out $line
        if ($row.Status -eq 'FAIL') { $failedCount++ }
    }
    if ($failedCount -gt 0) {
        Write-Out ('overall: FAIL ({0})' -f $failedCount)
    } else {
        Write-Out 'overall: PASS'
    }
    Write-Out ('output: {0}' -f $outFile)
} catch {
    $failedCount = 1
    if ($script:Writer) {
        Write-Out ('FATAL: ' + $_.Exception.Message)
    }
} finally {
    if ($script:Writer) {
        $script:Writer.Flush()
        $script:Writer.Dispose()
        $script:Writer = $null
    }
}

if ($failedCount -gt 0) { exit 1 }
exit 0
