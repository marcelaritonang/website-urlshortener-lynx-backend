# Database Health Check Script
# Save this file with UTF-8 BOM encoding

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  LYNX URL Shortener - Database Check" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# Check if backend is running
$backendRunning = Get-Process -Name "main" -ErrorAction SilentlyContinue
if ($backendRunning) {
    Write-Host "[OK] Backend is running (PID: $($backendRunning.Id))" -ForegroundColor Green
} else {
    Write-Host "[INFO] Backend is NOT running" -ForegroundColor Yellow
}

Write-Host ("-" * 40)

# Check PostgreSQL container
$postgresRunning = docker ps --filter "name=lynx-postgres" --format "{{.Names}}" 2>$null
if ($postgresRunning -eq "lynx-postgres") {
    Write-Host "[OK] PostgreSQL container is running`n" -ForegroundColor Green
    
    # Check if container is healthy (not restarting)
    $postgresStatus = docker ps --filter "name=lynx-postgres" --format "{{.Status}}"
    if ($postgresStatus -match "Restarting") {
        Write-Host "[ERROR] PostgreSQL is restarting continuously!" -ForegroundColor Red
        Write-Host "   This usually means volume corruption" -ForegroundColor Yellow
        Write-Host "   Run: .\fix_containers.ps1 (will recreate volumes)" -ForegroundColor Yellow
        exit 1
    }
} else {
    Write-Host "[ERROR] PostgreSQL container is NOT running!" -ForegroundColor Red
    Write-Host "   Run: docker-compose up -d" -ForegroundColor Yellow
    Write-Host "   Or: .\fix_containers.ps1 (if container keeps restarting)" -ForegroundColor Yellow
    exit 1
}

# Check tables using pg_tables (more reliable)
Write-Host "Checking tables..." -ForegroundColor Cyan

# Get container name dynamically
$containerName = docker ps --filter "name=postgres" --format "{{.Names}}" | Select-Object -First 1
if (-not $containerName) {
    Write-Host "[ERROR] PostgreSQL container not found!" -ForegroundColor Red
    exit 1
}

# Get all tables in public schema
$tableList = docker exec $containerName psql -U lynx_user -d lynx_db -t -A -c "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename;" 2>$null

if ($tableList -and $tableList.Trim() -ne "") {
    $tables = $tableList -split "`n" | Where-Object { $_ -match '\S' } | ForEach-Object { $_.Trim() }
    
    $hasUsers = $tables -contains "users"
    $hasUrls = $tables -contains "urls"
    
    if ($hasUsers -and $hasUrls) {
        Write-Host "[OK] Tables found: users, urls`n" -ForegroundColor Green
        
        # Count records
        Write-Host "Record counts:" -ForegroundColor Cyan
        $userCount = docker exec $containerName psql -U lynx_user -d lynx_db -t -A -c "SELECT COUNT(*) FROM users;" 2>$null
        $urlCount = docker exec $containerName psql -U lynx_user -d lynx_db -t -A -c "SELECT COUNT(*) FROM urls;" 2>$null
        
        Write-Host "  Users: $($userCount.Trim())" -ForegroundColor White
        Write-Host "  URLs:  $($urlCount.Trim())`n" -ForegroundColor White
        
        # Show recent URLs (if any exist)
        if ([int]$urlCount.Trim() -gt 0) {
            Write-Host "Recent URLs (Last 5):" -ForegroundColor Cyan
            docker exec $containerName psql -U lynx_user -d lynx_db -c "SELECT short_code, left(long_url, 40) as url, clicks, is_anonymous, to_char(created_at, 'YYYY-MM-DD HH24:MI') as created FROM urls ORDER BY created_at DESC LIMIT 5;" 2>$null
            
            # Show click statistics
            Write-Host "`nClick Counts (Top 5):" -ForegroundColor Cyan
            docker exec $containerName psql -U lynx_user -d lynx_db -c "SELECT short_code, left(long_url, 40) as url, clicks FROM urls WHERE clicks > 0 ORDER BY clicks DESC LIMIT 5;" 2>$null
        } else {
            Write-Host "[INFO] No URLs created yet. Create one via API!" -ForegroundColor Yellow
        }
        
        # Show indexes
        Write-Host "`nIndexes:" -ForegroundColor Cyan
        $indexCount = docker exec $containerName psql -U lynx_user -d lynx_db -t -A -c "SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public';" 2>$null
        Write-Host "  Total indexes: $($indexCount.Trim())" -ForegroundColor White
        
    } else {
        Write-Host "[WARNING] Tables incomplete:" -ForegroundColor Yellow
        Write-Host "   Users: $(if($hasUsers){'EXISTS'}else{'MISSING'})" -ForegroundColor Yellow
        Write-Host "   URLs:  $(if($hasUrls){'EXISTS'}else{'MISSING'})" -ForegroundColor Yellow
    }
} else {
    Write-Host "[ERROR] No tables found!" -ForegroundColor Red
    Write-Host "`n[FIX] Run: .\force_migrate.ps1" -ForegroundColor Cyan
}

Write-Host "`n" + ("-" * 40)

# Check data persistence
Write-Host "`nData Persistence Check:" -ForegroundColor Cyan
$volumes = docker volume ls --format "{{.Name}}" | Where-Object { $_ -match "postgres" }
if ($volumes) {
    Write-Host "[OK] PostgreSQL data is persistent" -ForegroundColor Green
    $volumes | ForEach-Object { Write-Host "  Volume: $_" -ForegroundColor Gray }
} else {
    Write-Host "[WARNING] No persistent volume detected!" -ForegroundColor Yellow
}

# Check Redis
$redisRunning = docker ps --filter "name=lynx-redis" --format "{{.Names}}" 2>$null
if ($redisRunning -eq "lynx-redis") {
    Write-Host "[OK] Redis container is running" -ForegroundColor Green
    
    # Test Redis connection
    $redisPing = docker exec lynx-redis redis-cli ping 2>$null
    if ($redisPing -eq "PONG") {
        Write-Host "[OK] Redis responding to PING" -ForegroundColor Green
        
        # Show cached URLs
        Write-Host "`nRedis Cache:" -ForegroundColor Cyan
        $cachedUrls = docker exec lynx-redis redis-cli --scan --pattern "url:*" 2>$null
        if ($cachedUrls) {
            $cacheCount = ($cachedUrls -split "`n" | Where-Object { $_ -match '\S' }).Count
            Write-Host "  Cached URLs: $cacheCount" -ForegroundColor White
            
            # Show sample cached URLs (first 3)
            $sampleKeys = ($cachedUrls -split "`n" | Where-Object { $_ -match '\S' } | Select-Object -First 3)
            Write-Host "  Sample keys:" -ForegroundColor DarkGray
            foreach ($key in $sampleKeys) {
                $value = docker exec lynx-redis redis-cli GET $key 2>$null
                Write-Host "    $key -> $($value.Substring(0, [Math]::Min(40, $value.Length)))..." -ForegroundColor DarkGray
            }
        } else {
            Write-Host "  No cached URLs yet" -ForegroundColor Yellow
        }
        
        # Show click counters
        $clickKeys = docker exec lynx-redis redis-cli --scan --pattern "clicks:*" 2>$null
        if ($clickKeys) {
            $clickCount = ($clickKeys -split "`n" | Where-Object { $_ -match '\S' }).Count
            Write-Host "  Click counters: $clickCount" -ForegroundColor White
        }
    }
} else {
    Write-Host "[ERROR] Redis container is NOT running!" -ForegroundColor Red
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Quick Commands:" -ForegroundColor Yellow
Write-Host "  1. Start backend: go run main.go" -ForegroundColor White
Write-Host "  2. Test health: curl http://localhost:8080/health" -ForegroundColor White
Write-Host "  3. Create URL: curl -X POST http://localhost:8080/api/urls -H 'Content-Type: application/json' -d '{\"long_url\":\"https://google.com\"}'" -ForegroundColor White
Write-Host "  4. Register user: curl -X POST http://localhost:8080/v1/auth/register -H 'Content-Type: application/json' -d '{\"email\":\"test@example.com\",\"password\":\"password123\",\"first_name\":\"Test\",\"last_name\":\"User\"}'" -ForegroundColor White
Write-Host "  5. View in pgAdmin: localhost:5050" -ForegroundColor White
Write-Host "  6. Force migration: .\force_migrate.ps1" -ForegroundColor White
Write-Host "========================================`n" -ForegroundColor Cyan
