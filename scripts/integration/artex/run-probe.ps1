param(
    [Parameter(Mandatory = $true)][string]$BinaryPath,
    [Parameter(Mandatory = $true)][string]$LabRoot,
    [ValidateSet('bootstrap','verify-restart')][string]$Mode = 'bootstrap',
    [switch]$Recreate
)
$ErrorActionPreference = 'Stop'
$runtime = Join-Path ([IO.Path]::GetFullPath($LabRoot)) 'runtime'
$secrets = [IO.File]::ReadAllText((Join-Path $runtime 'credentials.json')) | ConvertFrom-Json
$env:ARTEX_LAB_RUNTIME = $runtime.Replace([char]92, [char]47)
$env:ARTEX_LAB_BINARY = [IO.Path]::GetFullPath($BinaryPath).Replace([char]92, [char]47)
$tokenSurvived = $false
if ($Recreate) {
    if ($Mode -ne 'verify-restart') { throw 'Use -Recreate only with verify-restart.' }
    $login = @{username='ARTEX';password=$secrets.adminPassword} | ConvertTo-Json -Compress
    $token = (Invoke-RestMethod 'http://127.0.0.1:18787/api/auth/login' -Method Post -ContentType 'application/json' -Body $login -TimeoutSec 10).token
    docker compose -f (Join-Path $PSScriptRoot 'compose.yml') up -d --no-deps --force-recreate artex
    if ($LASTEXITCODE -ne 0) { throw 'Lab recreation failed.' }
    $ready = $false
    for ($attempt=0; $attempt -lt 30; $attempt++) {
        try {
            $health = Invoke-RestMethod 'http://127.0.0.1:18787/api/health' -TimeoutSec 2
            if ($health.ok) { $ready = $true; break }
        } catch { }
        Start-Sleep -Milliseconds 500
    }
    if (-not $ready) { throw 'Recreated lab is not ready.' }
    $null = Invoke-RestMethod 'http://127.0.0.1:18787/api/tasks' -Headers @{Authorization=('Bearer '+$token)} -TimeoutSec 10
    $tokenSurvived = $true
    Write-Output 'PASS existing_control_token_survives_container_recreation'
}
$env:ARTEX_LAB_ADMIN_PASSWORD = $secrets.adminPassword
$env:ARTEX_LAB_KEY_A = $secrets.modelKeyA
$env:ARTEX_LAB_KEY_B = $secrets.modelKeyB
$env:ARTEX_LAB_KEY_ROTATED = $secrets.modelKeyRotated
$env:ARTEX_LAB_STATE = Join-Path $runtime 'probe-state.json'
$repoRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$suffix = if ($Mode -eq 'bootstrap') { 'bootstrap' } else { 'recreate' }
$env:ARTEX_LAB_REPORT = Join-Path $repoRoot ('docs/merge-baseline/' + (Get-Date -Format 'yyyy-MM-dd') + '-artex-live-' + $suffix + '.json')
[IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($env:ARTEX_LAB_REPORT)) | Out-Null
node (Join-Path $PSScriptRoot 'probe.mjs') $Mode
$resultCode = $LASTEXITCODE
if (Test-Path -LiteralPath $env:ARTEX_LAB_REPORT) {
    $report = [IO.File]::ReadAllText($env:ARTEX_LAB_REPORT) | ConvertFrom-Json
    $report | Add-Member -NotePropertyName binarySHA256 -NotePropertyValue (Get-FileHash -LiteralPath $BinaryPath -Algorithm SHA256).Hash.ToLowerInvariant() -Force
    if ($tokenSurvived) {
        $report.checks += [pscustomobject]@{name='existing_control_token_survives_container_recreation';passed=$true}
    }
    [IO.File]::WriteAllText($env:ARTEX_LAB_REPORT, ($report | ConvertTo-Json -Depth 6), [Text.UTF8Encoding]::new($false))
}
exit $resultCode
