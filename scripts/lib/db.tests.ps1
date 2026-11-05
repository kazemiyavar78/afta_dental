# Converts sample sqlserver:// URLs and checks Data Source. Does not open a connection.
# Run: powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\lib\db.tests.ps1

. (Join-Path $PSScriptRoot 'db.ps1')

function Get-AftaConnectionBuilder {
    param([string]$Url)
    $cs = Convert-AftaConnectionString -Raw $Url
    Add-Type -AssemblyName System.Data
    return New-Object System.Data.SqlClient.SqlConnectionStringBuilder $cs
}

function Register-AftaDbTests {
    Describe 'Convert-AftaConnectionString Data Source' {
        It 'turns host:port into host,port' {
            $builder = Get-AftaConnectionBuilder 'sqlserver://afta:secret@db.internal:1433?database=AftaDb'
            $builder['Data Source'] | Should Be 'db.internal,1433'
            $builder['Initial Catalog'] | Should Be 'AftaDb'
        }

        It 'turns host/INSTANCE into host\INSTANCE' {
            $builder = Get-AftaConnectionBuilder 'sqlserver://afta:secret@db.internal/SQLEXPRESS?database=AftaDb'
            $builder['Data Source'] | Should Be 'db.internal\SQLEXPRESS'
        }

        It 'turns host:port/INSTANCE into host\INSTANCE,port' {
            $builder = Get-AftaConnectionBuilder 'sqlserver://afta:secret@db.internal:1433/SQLEXPRESS?database=AftaDb'
            $builder['Data Source'] | Should Be 'db.internal\SQLEXPRESS,1433'
        }

        It 'turns host/INSTANCE:port into host\INSTANCE,port' {
            $builder = Get-AftaConnectionBuilder 'sqlserver://afta:secret@db.internal/SQLEXPRESS:1433?database=AftaDb'
            $builder['Data Source'] | Should Be 'db.internal\SQLEXPRESS,1433'
        }

        It 'keeps a host without a port' {
            $builder = Get-AftaConnectionBuilder 'sqlserver://afta:secret@db.internal?database=AftaDb'
            $builder['Data Source'] | Should Be 'db.internal'
        }

        It 'does not treat an encoded @ in the password as the host separator' {
            $builder = Get-AftaConnectionBuilder 'sqlserver://afta:p%40ss@db.internal:1433?database=AftaDb'
            $builder['Data Source'] | Should Be 'db.internal,1433'
        }

        It 'honors connection timeout and TrustServerCertificate from the query' {
            $builder = Get-AftaConnectionBuilder 'sqlserver://afta:secret@db.internal:1433?database=AftaDb&TrustServerCertificate=false&connection+timeout=45'
            $builder['Data Source'] | Should Be 'db.internal,1433'
            $builder['TrustServerCertificate'] | Should Be $false
            $builder['Connect Timeout'] | Should Be 45
        }

        It 'honors percent-encoded connection timeout and keeps encrypt=disable' {
            $builder = Get-AftaConnectionBuilder 'sqlserver://afta:secret@10.0.0.8:1433?database=AftaDb&encrypt=disable&connection%20timeout=20&TrustServerCertificate=true'
            $builder['Data Source'] | Should Be '10.0.0.8,1433'
            $builder['Connect Timeout'] | Should Be 20
            $builder['TrustServerCertificate'] | Should Be $true
            $builder['Encrypt'] | Should Be $false
        }

        It 'uses timeout 15 and TrustServerCertificate true when the query omits them' {
            $builder = Get-AftaConnectionBuilder 'sqlserver://afta:secret@db.internal:1433?database=AftaDb'
            $builder['Data Source'] | Should Be 'db.internal,1433'
            $builder['Connect Timeout'] | Should Be 15
            $builder['TrustServerCertificate'] | Should Be $true
        }
    }

    Describe 'DataTable return is not unrolled' {
        It 'keeps an empty DataTable and its row count' {
            $got = Return-AftaDataTable (New-AftaSampleTable 0)
            $got -is [System.Data.DataTable] | Should Be $true
            $got.Rows.Count | Should Be 0
        }

        It 'keeps a two-row DataTable and its row count' {
            $got = Return-AftaDataTable (New-AftaSampleTable 2)
            $got -is [System.Data.DataTable] | Should Be $true
            $got.Rows.Count | Should Be 2
            $got.Columns.Count | Should Be 2
        }
    }
}

# همان الگوی return ,$table در Invoke-AftaQuery. بدون این کاما PowerShell ردیف‌ها را باز می‌کند.
function Return-AftaDataTable {
    param($Table)
    return ,$Table
}

function New-AftaSampleTable {
    param([int]$RowCount)
    $table = New-Object System.Data.DataTable
    [void]$table.Columns.Add('id', [int])
    [void]$table.Columns.Add('name', [string])
    for ($i = 1; $i -le $RowCount; $i++) {
        $row = $table.NewRow()
        $row['id'] = $i
        $row['name'] = ('row' + $i)
        [void]$table.Rows.Add($row)
    }
    return ,$table
}

$calledByPester = $false
foreach ($frame in (Get-PSCallStack)) {
    if ($frame.Command -eq 'Invoke-Pester') {
        $calledByPester = $true
        break
    }
}

if ($calledByPester) {
    Register-AftaDbTests
    return
}

$pester = Get-Module -ListAvailable -Name Pester | Sort-Object Version -Descending | Select-Object -First 1
if ($pester) {
    Import-Module Pester -RequiredVersion $pester.Version -Force
    $result = Invoke-Pester -Script $PSCommandPath -PassThru
    if ($result.FailedCount -gt 0) { exit 1 }
    exit 0
}

Write-Host 'Pester is not installed; running plain assertions'
$failed = 0
function Assert-Equal {
    param($Name, $Actual, $Expected)
    if ($Actual -ne $Expected) {
        Write-Host ("FAIL {0}: actual=[{1}] expected=[{2}]" -f $Name, $Actual, $Expected)
        $script:failed++
    } else {
        Write-Host ("PASS {0}" -f $Name)
    }
}

$plain = @(
    @{ Name = 'host:port'; Url = 'sqlserver://afta:secret@db.internal:1433?database=AftaDb'; DataSource = 'db.internal,1433' },
    @{ Name = 'host/instance'; Url = 'sqlserver://afta:secret@db.internal/SQLEXPRESS?database=AftaDb'; DataSource = 'db.internal\SQLEXPRESS' },
    @{ Name = 'host:port/instance'; Url = 'sqlserver://afta:secret@db.internal:1433/SQLEXPRESS?database=AftaDb'; DataSource = 'db.internal\SQLEXPRESS,1433' },
    @{ Name = 'host/instance:port'; Url = 'sqlserver://afta:secret@db.internal/SQLEXPRESS:1433?database=AftaDb'; DataSource = 'db.internal\SQLEXPRESS,1433' }
)
foreach ($case in $plain) {
    $builder = Get-AftaConnectionBuilder $case.Url
    Assert-Equal $case.Name $builder['Data Source'] $case.DataSource
}
foreach ($rows in @(0, 2)) {
    $got = Return-AftaDataTable (New-AftaSampleTable $rows)
    $typeOk = $got -is [System.Data.DataTable]
    Assert-Equal ("datatable-type-" + $rows) $typeOk $true
    if ($typeOk) { Assert-Equal ("datatable-rows-" + $rows) $got.Rows.Count $rows }
}
if ($failed -gt 0) { exit 1 }
Write-Host 'plain assertions passed'
exit 0
