# Database outage demo (Windows PowerShell 5.1 or PowerShell 7).
#
# Proves that a PostgreSQL outage loses no events and creates no duplicates:
#   1. count rows in events
#   2. stop the postgres container
#   3. send 20 events to the API (it still answers 202, because it only needs Kafka)
#   4. wait 10 seconds, then start postgres again
#   5. wait for the consumer to catch up
#   6. check that exactly 20 new rows arrived
#
# Needs a running stack: docker compose up --build -d
# Usage (from the repo root):
#   powershell -ExecutionPolicy Bypass -File .\scripts\demo-db-outage.ps1

param(
    [string]$ApiUrl = "http://localhost:8080",
    [int]$Events = 20,
    [int]$CatchUpTimeout = 120
)

$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..")

function Invoke-Sql([string]$Query) {
    $out = docker compose exec -T postgres psql -U events_user -d events_db -tAc $Query
    return ("$out").Trim()
}

function Get-StatusCode([string]$Method, [string]$Url, [string]$Body) {
    try {
        $params = @{ Uri = $Url; Method = $Method; UseBasicParsing = $true; TimeoutSec = 10 }
        if ($Body) { $params.Body = $Body; $params.ContentType = "application/json" }
        return [int](Invoke-WebRequest @params).StatusCode
    } catch {
        if ($_.Exception.Response) { return [int]$_.Exception.Response.StatusCode }
        return 0
    }
}

function Write-Step([string]$Text) { Write-Host ""; Write-Host "==> $Text" -ForegroundColor Cyan }

Write-Step "1. Counting rows in events"
$before = [int](Invoke-Sql "SELECT COUNT(*) FROM events")
Write-Host "rows before: $before"

Write-Step "2. Stopping postgres"
$pgContainer = (docker compose ps -q postgres).Trim()
docker stop $pgContainer | Out-Null
Write-Host "postgres stopped"

Write-Step "3. Sending $Events events while the database is down"
$ids = @()
$accepted = 0
for ($i = 1; $i -le $Events; $i++) {
    $id = [guid]::NewGuid().ToString()
    $ids += "'$id'"
    $body = '{"event_id":"' + $id + '","event_type":"outage_demo","payload":{"n":' + $i + '}}'
    $code = Get-StatusCode "POST" "$ApiUrl/v1/events" $body
    if ($code -eq 202) { $accepted++ }
    Write-Host ("  event {0,2} -> HTTP {1}" -f $i, $code)
}
Write-Host "API accepted $accepted of $Events events with 202 while postgres was down"
$ready = Get-StatusCode "GET" "$ApiUrl/readyz" $null
Write-Host "readiness during outage: HTTP $ready (503 expected)"

Write-Step "4. Waiting 10 seconds (the consumer is retrying, then pausing)"
Start-Sleep -Seconds 10

Write-Step "5. Starting postgres again"
docker start $pgContainer | Out-Null
do {
    Start-Sleep -Seconds 1
    docker compose exec -T postgres pg_isready -U events_user -d events_db *> $null
} until ($LASTEXITCODE -eq 0)
Write-Host "postgres is accepting connections"

Write-Step "6. Waiting for the consumer to catch up (max ${CatchUpTimeout}s)"
$idList = $ids -join ","
$arrived = 0
for ($t = 0; $t -lt $CatchUpTimeout; $t++) {
    try { $arrived = [int](Invoke-Sql "SELECT COUNT(*) FROM events WHERE event_id IN ($idList)") } catch { $arrived = 0 }
    if ($arrived -eq $Events) { break }
    Start-Sleep -Seconds 1
}
$after = [int](Invoke-Sql "SELECT COUNT(*) FROM events")
$dlq = [int](Invoke-Sql "SELECT COUNT(*) FROM dead_letters WHERE original_key IN ($idList)")
$newRows = $after - $before

Write-Step "Result"
Write-Host "rows before:            $before"
Write-Host "rows after:             $after"
Write-Host "new rows:               $newRows"
Write-Host "demo events stored:     $arrived of $Events"
Write-Host "demo events in the DLQ: $dlq"

if ($accepted -eq $Events -and $arrived -eq $Events -and $newRows -eq $Events -and $dlq -eq 0) {
    Write-Host "PASS: exactly $Events new rows arrived (no loss, no duplicates)" -ForegroundColor Green
} else {
    Write-Host "FAIL: expected exactly $Events new rows and none in the DLQ" -ForegroundColor Red
    exit 1
}
