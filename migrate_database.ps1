# Force Database Migration Script

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  LYNX - Force Database Migration" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# Check if PostgreSQL is running
$containerName = docker ps --filter "name=postgres" --format "{{.Names}}" | Select-Object -First 1
if (-not $containerName) {
    Write-Host "[ERROR] PostgreSQL container is NOT running!" -ForegroundColor Red
    Write-Host "   Run: docker-compose up -d" -ForegroundColor Yellow
    exit 1
}

Write-Host "  Using container: $containerName" -ForegroundColor Green

Write-Host "[1/3] Stopping backend if running..." -ForegroundColor Yellow
$backendProcess = Get-Process -Name "main" -ErrorAction SilentlyContinue
if ($backendProcess) {
    Stop-Process -Name "main" -Force
    Start-Sleep -Seconds 2
    Write-Host "   Backend stopped" -ForegroundColor Green
} else {
    Write-Host "   Backend was not running" -ForegroundColor Gray
}

Write-Host "`n[2/3] Checking database connection..." -ForegroundColor Yellow
$dbTest = docker exec $containerName psql -U lynx_user -d lynx_db -c "SELECT version();" 2>$null
if ($LASTEXITCODE -eq 0) {
    Write-Host "   Database connection OK" -ForegroundColor Green
} else {
    Write-Host "   [ERROR] Cannot connect to database!" -ForegroundColor Red
    exit 1
}

Write-Host "`n[3/3] Running migration (starting backend)..." -ForegroundColor Yellow
Write-Host "   This will auto-migrate tables..." -ForegroundColor Gray

# Start backend in background
Start-Process -FilePath "go" -ArgumentList "run", "main.go" -NoNewWindow

Write-Host "`n   Waiting for migration to complete (15 seconds)..." -ForegroundColor Gray
Start-Sleep -Seconds 15

# Check if tables were created
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Verifying tables..." -ForegroundColor Cyan

# Use a more reliable query to check tables - use count instead of EXISTS
$userTableCheck = docker exec $containerName psql -U lynx_user -d lynx_db -t -A -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users';" 2>$null
$urlTableCheck = docker exec $containerName psql -U lynx_user -d lynx_db -t -A -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'urls';" 2>$null

$userTableExists = ($userTableCheck -and $userTableCheck.Trim() -eq "1")
$urlTableExists = ($urlTableCheck -and $urlTableCheck.Trim() -eq "1")

if ($userTableExists -and $urlTableExists) {
    Write-Host "[OK] Migration successful! Tables created:" -ForegroundColor Green
    Write-Host ""
    docker exec lynx-postgres psql -U lynx_user -d lynx_db -c "\dt" 2>$null
    Write-Host ""
    
    # Show table details
    Write-Host "Table Statistics:" -ForegroundColor Cyan
    $userCount = docker exec lynx-postgres psql -U lynx_user -d lynx_db -t -c "SELECT COUNT(*) FROM users;" 2>$null
    $urlCount = docker exec lynx-postgres psql -U lynx_user -d lynx_db -t -c "SELECT COUNT(*) FROM urls;" 2>$null
    Write-Host "  Users table: $($userCount.Trim()) records" -ForegroundColor White
    Write-Host "  URLs table:  $($urlCount.Trim()) records" -ForegroundColor White
    
} else {
    Write-Host "[ERROR] Tables verification failed!" -ForegroundColor Red
    Write-Host "  Users table exists: $userTableExists" -ForegroundColor Yellow
    Write-Host "  URLs table exists: $urlTableExists" -ForegroundColor Yellow
    Write-Host "`nChecking what tables exist:" -ForegroundColor Yellow
    docker exec lynx-postgres psql -U lynx_user -d lynx_db -c "SELECT tablename FROM pg_tables WHERE schemaname = 'public';" 2>$null
    
    # Debug output
    Write-Host "`n[DEBUG] User table check: '$userTableCheck'" -ForegroundColor DarkGray
    Write-Host "[DEBUG] URL table check: '$urlTableCheck'" -ForegroundColor DarkGray
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Run: .\check_database.ps1 (verify tables)" -ForegroundColor White
Write-Host "  2. Backend is now running in background" -ForegroundColor White
Write-Host "  3. Test API: curl http://localhost:8080/health" -ForegroundColor White
Write-Host "  4. To stop: Get-Process -Name 'main' | Stop-Process" -ForegroundColor White
Write-Host "========================================`n" -ForegroundColor Cyan
