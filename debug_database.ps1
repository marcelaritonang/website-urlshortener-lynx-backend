# Deep Database Debug Script

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  DATABASE DEBUG SCRIPT" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# 1. Check Docker containers
Write-Host "[1] Docker Containers:" -ForegroundColor Yellow
docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}" | Where-Object { $_ -match "postgres|lynx" }

# 2. Check if container exists
Write-Host "`n[2] PostgreSQL Container Details:" -ForegroundColor Yellow
$containerName = docker ps --filter "name=postgres" --format "{{.Names}}" | Select-Object -First 1
if ($containerName) {
    Write-Host "  Container name: $containerName" -ForegroundColor Green
} else {
    Write-Host "  No PostgreSQL container found!" -ForegroundColor Red
    exit 1
}

# 3. Test connection to database
Write-Host "`n[3] Testing Database Connection:" -ForegroundColor Yellow
$testConnection = docker exec $containerName psql -U lynx_user -d lynx_db -c "SELECT version();" 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "  Connection: OK" -ForegroundColor Green
} else {
    Write-Host "  Connection: FAILED" -ForegroundColor Red
    Write-Host "  Error: $testConnection" -ForegroundColor Red
}

# 4. List all databases
Write-Host "`n[4] All Databases:" -ForegroundColor Yellow
docker exec $containerName psql -U lynx_user -d postgres -c "\l" 2>$null

# 5. List all schemas
Write-Host "`n[5] All Schemas in lynx_db:" -ForegroundColor Yellow
docker exec $containerName psql -U lynx_user -d lynx_db -c "SELECT schema_name FROM information_schema.schemata;" 2>$null

# 6. List tables in public schema
Write-Host "`n[6] Tables in PUBLIC schema:" -ForegroundColor Yellow
docker exec $containerName psql -U lynx_user -d lynx_db -c "\dt public.*" 2>$null

# 7. List ALL tables (any schema)
Write-Host "`n[7] All Tables (any schema):" -ForegroundColor Yellow
docker exec $containerName psql -U lynx_user -d lynx_db -c "SELECT schemaname, tablename FROM pg_tables WHERE schemaname NOT IN ('pg_catalog', 'information_schema') ORDER BY schemaname, tablename;" 2>$null

# 8. Check search_path
Write-Host "`n[8] Current search_path:" -ForegroundColor Yellow
docker exec $containerName psql -U lynx_user -d lynx_db -c "SHOW search_path;" 2>$null

# 9. Count tables
Write-Host "`n[9] Table Counts:" -ForegroundColor Yellow
docker exec $containerName psql -U lynx_user -d lynx_db -t -A -c "SELECT schemaname, COUNT(*) as table_count FROM pg_tables WHERE schemaname NOT IN ('pg_catalog', 'information_schema') GROUP BY schemaname;" 2>$null

# 10. Try to query users and urls directly
Write-Host "`n[10] Direct Query Test:" -ForegroundColor Yellow
Write-Host "  Trying: SELECT * FROM users LIMIT 1" -ForegroundColor Gray
$userTest = docker exec $containerName psql -U lynx_user -d lynx_db -c "SELECT COUNT(*) FROM users;" 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "  Users table: EXISTS (count: $($userTest.Trim()))" -ForegroundColor Green
} else {
    Write-Host "  Users table: NOT FOUND or ERROR" -ForegroundColor Red
    Write-Host "  Error: $userTest" -ForegroundColor Red
}

Write-Host "`n  Trying: SELECT * FROM urls LIMIT 1" -ForegroundColor Gray
$urlTest = docker exec $containerName psql -U lynx_user -d lynx_db -c "SELECT COUNT(*) FROM urls;" 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "  URLs table: EXISTS (count: $($urlTest.Trim()))" -ForegroundColor Green
} else {
    Write-Host "  URLs table: NOT FOUND or ERROR" -ForegroundColor Red
    Write-Host "  Error: $urlTest" -ForegroundColor Red
}

# ✅ NEW: Check Docker volumes
Write-Host "`n[11] Docker Volumes (Data Persistence):" -ForegroundColor Yellow
$volumes = docker volume ls --format "{{.Name}}" | Where-Object { $_ -match "lynx|postgres|redis" }
if ($volumes) {
    Write-Host "  Persistent volumes found:" -ForegroundColor Green
    $volumes | ForEach-Object { 
        $volSize = docker volume inspect $_ --format "{{.Mountpoint}}"
        Write-Host "    - $_" -ForegroundColor Green
    }
} else {
    Write-Host "  [WARNING] No persistent volumes! Data will be lost on restart!" -ForegroundColor Red
    Write-Host "  [FIX] Update docker-compose.yml to use volumes" -ForegroundColor Yellow
}

# ✅ NEW: Check container mounts
Write-Host "`n[12] Container Volume Mounts:" -ForegroundColor Yellow
$mounts = docker inspect $containerName --format "{{json .Mounts}}" | ConvertFrom-Json
if ($mounts) {
    $pgDataMount = $mounts | Where-Object { $_.Destination -eq "/var/lib/postgresql/data" }
    if ($pgDataMount) {
        Write-Host "  PostgreSQL data is persistent: YES" -ForegroundColor Green
        Write-Host "    Volume: $($pgDataMount.Name)" -ForegroundColor Gray
    } else {
        Write-Host "  [ERROR] PostgreSQL data is NOT persistent!" -ForegroundColor Red
        Write-Host "  This is why tables disappear when container stops!" -ForegroundColor Red
    }
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "DEBUG COMPLETE" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan
