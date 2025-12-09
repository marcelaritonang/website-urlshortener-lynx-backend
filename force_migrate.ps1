# Force Migration with Persistent Verification

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  FORCE MIGRATION SCRIPT" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

$containerName = "lynx-postgres"

# 1. Stop backend if running
Write-Host "[1/5] Stopping backend..." -ForegroundColor Yellow
Get-Process -Name "main" -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Seconds 2
Write-Host "  Backend stopped" -ForegroundColor Green

# 2. Verify container is running
Write-Host "`n[2/5] Checking PostgreSQL container..." -ForegroundColor Yellow
$running = docker ps --filter "name=$containerName" --format "{{.Names}}" 2>$null
if ($running -ne $containerName) {
    Write-Host "  [ERROR] Container not running!" -ForegroundColor Red
    exit 1
}
Write-Host "  Container is running" -ForegroundColor Green

# 3. Create tables directly via SQL (bypass GORM)
Write-Host "`n[3/5] Creating tables via SQL..." -ForegroundColor Yellow

$createUsersSQL = @"
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    reset_token VARCHAR(255),
    reset_token_expiry TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_reset_token ON users(reset_token);
"@

$createUrlsSQL = @"
CREATE TABLE IF NOT EXISTS urls (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    long_url TEXT NOT NULL,
    short_url VARCHAR(255) UNIQUE NOT NULL,
    short_code VARCHAR(50) UNIQUE NOT NULL,
    clicks BIGINT DEFAULT 0,
    is_anonymous BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls(user_id);
CREATE INDEX IF NOT EXISTS idx_urls_short_url ON urls(short_url);
CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code);
CREATE INDEX IF NOT EXISTS idx_urls_is_anonymous ON urls(is_anonymous);
CREATE INDEX IF NOT EXISTS idx_urls_deleted_at ON urls(deleted_at);
"@

# Execute SQL
docker exec $containerName psql -U lynx_user -d lynx_db -c "$createUsersSQL" 2>&1 | Out-Null
docker exec $containerName psql -U lynx_user -d lynx_db -c "$createUrlsSQL" 2>&1 | Out-Null

Write-Host "  Tables created" -ForegroundColor Green

# 4. Verify tables exist
Write-Host "`n[4/5] Verifying tables..." -ForegroundColor Yellow
$tables = docker exec $containerName psql -U lynx_user -d lynx_db -t -A -c "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename;" 2>$null

if ($tables -match "users" -and $tables -match "urls") {
    Write-Host "  [OK] Tables verified!" -ForegroundColor Green
    Write-Host "  Tables: $($tables -split '`n' -join ', ')" -ForegroundColor Gray
} else {
    Write-Host "  [ERROR] Tables not found!" -ForegroundColor Red
    exit 1
}

# 5. Test persistence
Write-Host "`n[5/5] Testing persistence..." -ForegroundColor Yellow
Write-Host "  Restarting container..." -ForegroundColor Gray
docker restart $containerName | Out-Null
Start-Sleep -Seconds 5

$tablesAfter = docker exec $containerName psql -U lynx_user -d lynx_db -t -A -c "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename;" 2>$null

if ($tablesAfter -match "users" -and $tablesAfter -match "urls") {
    Write-Host "  [OK] Tables PERSISTENT after restart!" -ForegroundColor Green
} else {
    Write-Host "  [ERROR] Tables LOST after restart!" -ForegroundColor Red
    Write-Host "  Volume issue detected!" -ForegroundColor Red
    exit 1
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "SUCCESS! Tables are persistent." -ForegroundColor Green
Write-Host "========================================`n" -ForegroundColor Cyan
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Run: .\check_database.ps1" -ForegroundColor White
Write-Host "  2. Start backend: go run main.go" -ForegroundColor White
Write-Host "  3. Backend will find existing tables" -ForegroundColor White
Write-Host "========================================`n" -ForegroundColor Cyan
