# Fix Container Issues Script

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  FIX CONTAINERS SCRIPT" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# 1. Stop all containers
Write-Host "[1/7] Stopping containers..." -ForegroundColor Yellow
docker-compose down 2>&1 | Out-Null
Start-Sleep -Seconds 2
Write-Host "  Containers stopped" -ForegroundColor Green

# 2. Remove corrupt volumes
Write-Host "`n[2/7] Removing corrupt volumes..." -ForegroundColor Yellow
$postgresVolume = docker volume ls --format "{{.Name}}" | Where-Object { $_ -match "postgres" }
if ($postgresVolume) {
    Write-Host "  Removing: $postgresVolume" -ForegroundColor Gray
    docker volume rm $postgresVolume 2>&1 | Out-Null
    Write-Host "  Volume removed" -ForegroundColor Green
} else {
    Write-Host "  No PostgreSQL volume found" -ForegroundColor Yellow
}

# 3. Clean up any dangling volumes
Write-Host "`n[3/7] Cleaning dangling volumes..." -ForegroundColor Yellow
docker volume prune -f 2>&1 | Out-Null
Write-Host "  Cleanup complete" -ForegroundColor Green

# 4. Start fresh containers
Write-Host "`n[4/7] Starting fresh containers..." -ForegroundColor Yellow
docker-compose up -d
Write-Host "  Containers starting..." -ForegroundColor Gray

# 5. Wait for containers to be healthy
Write-Host "`n[5/7] Waiting for containers to be healthy (30 seconds)..." -ForegroundColor Yellow
for ($i = 1; $i -le 30; $i++) {
    Write-Progress -Activity "Waiting for containers" -Status "$i/30 seconds" -PercentComplete (($i/30)*100)
    Start-Sleep -Seconds 1
}
Write-Progress -Activity "Waiting for containers" -Completed
Write-Host "  Wait complete" -ForegroundColor Green

# 6. Verify containers are running
Write-Host "`n[6/7] Verifying containers..." -ForegroundColor Yellow
$postgres = docker ps --filter "name=lynx-postgres" --format "{{.Status}}"
$redis = docker ps --filter "name=lynx-redis" --format "{{.Status}}"

if ($postgres -match "Up") {
    Write-Host "  PostgreSQL: RUNNING ✓" -ForegroundColor Green
} else {
    Write-Host "  PostgreSQL: FAILED ✗" -ForegroundColor Red
    Write-Host "  Logs:" -ForegroundColor Yellow
    docker logs lynx-postgres --tail 10
    exit 1
}

if ($redis -match "Up") {
    Write-Host "  Redis: RUNNING ✓" -ForegroundColor Green
} else {
    Write-Host "  Redis: FAILED ✗" -ForegroundColor Red
    exit 1
}

# 7. Test connections
Write-Host "`n[7/7] Testing connections..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

$pgTest = docker exec lynx-postgres pg_isready -U lynx_user -d lynx_db 2>&1
if ($pgTest -match "accepting") {
    Write-Host "  PostgreSQL connection: OK ✓" -ForegroundColor Green
} else {
    Write-Host "  PostgreSQL connection: WAITING..." -ForegroundColor Yellow
    Start-Sleep -Seconds 5
    $pgTest = docker exec lynx-postgres pg_isready -U lynx_user -d lynx_db 2>&1
    if ($pgTest -match "accepting") {
        Write-Host "  PostgreSQL connection: OK ✓" -ForegroundColor Green
    } else {
        Write-Host "  PostgreSQL connection: FAILED ✗" -ForegroundColor Red
    }
}

$redisTest = docker exec lynx-redis redis-cli ping 2>&1
if ($redisTest -eq "PONG") {
    Write-Host "  Redis connection: OK ✓" -ForegroundColor Green
} else {
    Write-Host "  Redis connection: FAILED ✗" -ForegroundColor Red
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Containers are ready!" -ForegroundColor Green
Write-Host "========================================`n" -ForegroundColor Cyan
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Create tables: .\force_migrate.ps1" -ForegroundColor White
Write-Host "  2. Verify setup: .\check_database.ps1" -ForegroundColor White
Write-Host "  3. Start backend: go run main.go" -ForegroundColor White
Write-Host "========================================`n" -ForegroundColor Cyan
