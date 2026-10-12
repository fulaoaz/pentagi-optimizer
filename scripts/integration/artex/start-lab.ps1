param(
    [Parameter(Mandatory = $true)][string]$BinaryPath,
    [Parameter(Mandatory = $true)][string]$LabRoot
)
$ErrorActionPreference = 'Stop'
$binary = [IO.Path]::GetFullPath($BinaryPath)
if (-not (Test-Path -LiteralPath $binary -PathType Leaf)) { throw 'Build the pinned Linux ARTEX binary first.' }
$signature = [IO.File]::OpenRead($binary)
try {
    $header = New-Object byte[] 4
    $null = $signature.Read($header, 0, 4)
    if ([BitConverter]::ToString($header) -ne '7F-45-4C-46') { throw 'ARTEX binary must be a Linux ELF executable.' }
} finally { $signature.Dispose() }
$runtime = Join-Path ([IO.Path]::GetFullPath($LabRoot)) 'runtime'
$repoRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..')).TrimEnd([char]92, [char]47)
if ($runtime.StartsWith($repoRoot + [IO.Path]::DirectorySeparatorChar, [StringComparison]::OrdinalIgnoreCase)) {
    throw 'Keep lab credentials outside the Git working tree.'
}
[IO.Directory]::CreateDirectory($runtime) | Out-Null
if ($env:OS -eq 'Windows_NT') {
    $account = [Security.Principal.WindowsIdentity]::GetCurrent().Name
    icacls $runtime /inheritance:r /grant:r ($account + ':(OI)(CI)F') | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'Failed to restrict local lab credential permissions.' }
}
$secretPath = Join-Path $runtime 'credentials.json'
$utf8 = [Text.UTF8Encoding]::new($false)
if (-not (Test-Path -LiteralPath $secretPath)) {
    $secrets = [ordered]@{
        databasePassword = [guid]::NewGuid().ToString('N')
        adminPassword = 'Lab!' + [guid]::NewGuid().ToString('N')
        modelKeyA = 'fixture_' + [guid]::NewGuid().ToString('N')
        modelKeyB = 'fixture_' + [guid]::NewGuid().ToString('N')
        modelKeyRotated = 'fixture_' + [guid]::NewGuid().ToString('N')
    }
    [IO.File]::WriteAllText($secretPath, ($secrets | ConvertTo-Json), $utf8)
}
$secrets = [IO.File]::ReadAllText($secretPath) | ConvertFrom-Json
$jwtFile = Join-Path $runtime 'jwt.key'
if (-not (Test-Path -LiteralPath $jwtFile)) {
    [IO.File]::WriteAllText($jwtFile, ([guid]::NewGuid().ToString('N') + [guid]::NewGuid().ToString('N')), $utf8)
}
[IO.File]::WriteAllLines((Join-Path $runtime 'postgres.env'), @(
    'POSTGRES_USER=artex', 'POSTGRES_DB=artex', ('POSTGRES_PASSWORD=' + $secrets.databasePassword)
), $utf8)
[IO.File]::WriteAllLines((Join-Path $runtime 'artex.env'), @(
    ('ARTEX_PG_DSN=postgres://artex:' + $secrets.databasePassword + '@postgres:5432/artex?sslmode=disable')
), $utf8)
$env:ARTEX_LAB_RUNTIME = $runtime.Replace([char]92, [char]47)
$env:ARTEX_LAB_BINARY = $binary.Replace([char]92, [char]47)
$compose = Join-Path $PSScriptRoot 'compose.yml'
docker compose -f $compose config --quiet
if ($LASTEXITCODE -ne 0) { throw 'Lab compose configuration failed validation.' }
docker compose -f $compose up -d
if ($LASTEXITCODE -ne 0) { throw 'Lab startup failed.' }
$until = (Get-Date).AddSeconds(50)
do {
    try {
        $health = Invoke-RestMethod 'http://127.0.0.1:18787/api/health' -TimeoutSec 2
        if ($health.ok -and $health.service -eq 'artex') {
            Write-Output 'Real ARTEX API is healthy at loopback port 18787; private lab network active.'
            exit 0
        }
    } catch { }
    Start-Sleep -Milliseconds 500
} while ((Get-Date) -lt $until)
throw 'ARTEX did not become healthy before the startup deadline.'
