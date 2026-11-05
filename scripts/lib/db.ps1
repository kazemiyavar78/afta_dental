# Shared SQL connection-string helpers for scripts/afta-discovery.ps1 and scripts/afta-check.ps1.
# Dot-source this file; do not run it directly. This file never opens a connection.

# مقدار query را به بولین SqlClient تبدیل می‌کند.
# فقط true/yes/1 و false/no/0/disable پذیرفته می‌شود.
function ConvertTo-AftaBool {
    param([string]$Value)
    switch ($Value.Trim().ToLowerInvariant()) {
        'true' { return $true }
        'yes' { return $true }
        '1' { return $true }
        'false' { return $false }
        'no' { return $false }
        '0' { return $false }
        'disable' { return $false }
        default { throw 'DATABASE_URL boolean query value is invalid' }
    }
}

# بخش host پس از @ را به Data Source قابل قبول SqlClient تبدیل می‌کند.
# host:port می‌شود host,port و host/INSTANCE می‌شود host\INSTANCE.
# اگر هر دو باشند نتیجه host\INSTANCE,port است.
function ConvertTo-AftaSqlDataSource {
    param([string]$HostAndPath)
    if ([string]::IsNullOrWhiteSpace($HostAndPath)) {
        throw 'DATABASE_URL host is empty'
    }

    $instance = $null
    $port = $null
    $endpoint = $HostAndPath.Trim()
    $slash = $endpoint.IndexOf('/')
    if ($slash -ge 0) {
        $rawInstance = $endpoint.Substring($slash + 1).Trim('/')
        $endpoint = $endpoint.Substring(0, $slash)
        if (-not [string]::IsNullOrWhiteSpace($rawInstance)) {
            if ($rawInstance -match '^(.+):(\d+)$') {
                $instance = $Matches[1]
                $port = $Matches[2]
            } else {
                $instance = $rawInstance
            }
        }
    }

    $hostName = $endpoint
    if ($endpoint.StartsWith('[')) {
        $end = $endpoint.IndexOf(']')
        if ($end -lt 1) { throw 'DATABASE_URL host is invalid' }
        $hostName = $endpoint.Substring(0, $end + 1)
        $rest = $endpoint.Substring($end + 1)
        if ($rest.StartsWith(':')) {
            $authorityPort = $rest.Substring(1)
            if ($authorityPort -notmatch '^\d+$') { throw 'DATABASE_URL port is invalid' }
            $port = $authorityPort
        } elseif (-not [string]::IsNullOrEmpty($rest)) {
            throw 'DATABASE_URL host is invalid'
        }
    } else {
        $colon = $endpoint.LastIndexOf(':')
        if ($colon -ge 0) {
            $maybePort = $endpoint.Substring($colon + 1)
            if ($maybePort -notmatch '^\d+$') { throw 'DATABASE_URL port is invalid' }
            $hostName = $endpoint.Substring(0, $colon)
            $port = $maybePort
        }
    }

    $hostName = [Uri]::UnescapeDataString($hostName)
    if ($instance) { $instance = [Uri]::UnescapeDataString($instance) }
    if ([string]::IsNullOrWhiteSpace($hostName)) { throw 'DATABASE_URL host is empty' }

    $dataSource = $hostName
    if (-not [string]::IsNullOrWhiteSpace($instance)) {
        $dataSource = $hostName + '\' + $instance
    }
    if (-not [string]::IsNullOrWhiteSpace($port)) {
        $dataSource = $dataSource + ',' + $port
    }
    return $dataSource
}

# URL به سبک Go (sqlserver://...) یا رشته ADO را به رشته اتصال System.Data.SqlClient تبدیل می‌کند.
# اگر connection timeout یا TrustServerCertificate در query باشد همان مقدار استفاده می‌شود؛
# در غیر این صورت مهلت ۱۵ ثانیه و TrustServerCertificate=true می‌ماند.
# این تابع اتصال باز نمی‌کند.
function Convert-AftaConnectionString {
    param([string]$Raw)
    if ([string]::IsNullOrWhiteSpace($Raw)) {
        throw 'DATABASE_URL is empty'
    }
    if ($Raw -notmatch '^[a-zA-Z][a-zA-Z0-9+.-]*://') {
        if ($Raw -notmatch '(?i)Connect Timeout=') {
            return ($Raw.TrimEnd(';') + ';Connect Timeout=15')
        }
        return $Raw
    }

    $rest = [regex]::Replace($Raw, '^[a-zA-Z][a-zA-Z0-9+.-]*://', '')
    $at = $rest.LastIndexOf('@')
    if ($at -lt 1) { throw 'DATABASE_URL URL has no userinfo' }
    $userInfo = $rest.Substring(0, $at)
    $hostPart = $rest.Substring($at + 1)
    $colon = $userInfo.IndexOf(':')
    if ($colon -lt 0) { throw 'DATABASE_URL URL has no password separator' }
    $user = [Uri]::UnescapeDataString($userInfo.Substring(0, $colon))
    $pass = [Uri]::UnescapeDataString($userInfo.Substring($colon + 1))
    $q = $hostPart.IndexOf('?')
    $hostAndPath = if ($q -ge 0) { $hostPart.Substring(0, $q) } else { $hostPart }
    $query = if ($q -ge 0) { $hostPart.Substring($q + 1) } else { '' }

    $database = 'master'
    $encrypt = $true
    $trust = $null
    $timeout = $null
    foreach ($pair in ($query -split '&')) {
        if ([string]::IsNullOrWhiteSpace($pair)) { continue }
        $kv = $pair -split '=', 2
        $key = [Uri]::UnescapeDataString(($kv[0] -replace '\+', ' ')).Trim().ToLowerInvariant()
        $value = ''
        if ($kv.Length -gt 1) {
            $value = [Uri]::UnescapeDataString(($kv[1] -replace '\+', ' ')).Trim()
        }
        switch ($key) {
            'database' { $database = $value }
            'initial catalog' { $database = $value }
            'encrypt' {
                if ($value -eq 'disable' -or $value -eq 'false' -or $value -eq '0') { $encrypt = $false }
            }
            'connection timeout' {
                $parsed = 0
                if (-not [int]::TryParse($value, [ref]$parsed) -or $parsed -lt 0) {
                    throw 'DATABASE_URL connection timeout is invalid'
                }
                $timeout = $parsed
            }
            'connect timeout' {
                $parsed = 0
                if (-not [int]::TryParse($value, [ref]$parsed) -or $parsed -lt 0) {
                    throw 'DATABASE_URL connection timeout is invalid'
                }
                $timeout = $parsed
            }
            'trustservercertificate' { $trust = ConvertTo-AftaBool $value }
            'trust server certificate' { $trust = ConvertTo-AftaBool $value }
        }
    }
    if ($null -eq $trust) { $trust = $true }
    if ($null -eq $timeout) { $timeout = 15 }

    Add-Type -AssemblyName System.Data
    $builder = New-Object System.Data.SqlClient.SqlConnectionStringBuilder
    $builder['Data Source'] = (ConvertTo-AftaSqlDataSource $hostAndPath)
    $builder['Initial Catalog'] = $database
    $builder['User ID'] = $user
    $builder['Password'] = $pass
    $builder['Encrypt'] = $encrypt
    $builder['TrustServerCertificate'] = $trust
    $builder['Connect Timeout'] = $timeout
    return $builder.ConnectionString
}
