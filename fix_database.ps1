Write-Host "🔧 Fixing Database - Creating Tables" -ForegroundColor Cyan
Write-Host "=" * 60

# 1. Check PostgreSQL is running
Write-Host "`n1. Checking PostgreSQL..." -ForegroundColor Yellow
$pgRunning = docker ps --filter "name=lynx-postgres" --format "{{.Names}}"
if (-not $pgRunning) {
    Write-Host "❌ PostgreSQL not running! Starting..." -ForegroundColor Red
    docker-compose up -d
    Start-Sleep -Seconds 5
}

# 2. Create SQL migration file
Write-Host "`n2. Creating migration SQL..." -ForegroundColor Yellow

$migrationSQL = @'
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    reset_token VARCHAR(255),
    reset_token_expiry TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS urls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    long_url TEXT NOT NULL,
    short_code VARCHAR(20) UNIQUE NOT NULL,
    short_url VARCHAR(255) NOT NULL,
    clicks BIGINT DEFAULT 0,
    is_anonymous BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls(user_id);
CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code);
CREATE INDEX IF NOT EXISTS idx_urls_created_at ON urls(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
'@

# Save to temp file
$migrationSQL | Out-File -FilePath "./temp_migration.sql" -Encoding UTF8

# 3. Copy to container
Write-Host "3. Copying SQL to container..." -ForegroundColor Yellow
docker cp ./temp_migration.sql lynx-postgres:/tmp/migration.sql

# 4. Run migration
Write-Host "4. Running migration..." -ForegroundColor Yellow
docker exec lynx-postgres psql -U lynx_user -d lynx_db -f /tmp/migration.sql

# 5. Verify tables
Write-Host "`n5. Verifying tables..." -ForegroundColor Yellow
docker exec lynx-postgres psql -U lynx_user -d lynx_db -c "\dt"

# 6. Show table details
Write-Host "`n6. Table details:" -ForegroundColor Yellow
docker exec lynx-postgres psql -U lynx_user -d lynx_db -c "\d users"
docker exec lynx-postgres psql -U lynx_user -d lynx_db -c "\d urls"

# 7. Cleanup
Remove-Item ./temp_migration.sql

Write-Host "`n" + "=" * 60
Write-Host "✅ Database fix complete!" -ForegroundColor Green
Write-Host "`nNext steps:" -ForegroundColor Cyan
Write-Host "  1. Run: docker exec -it lynx-postgres psql -U lynx_user -d lynx_db"
Write-Host "  2. Test: SELECT * FROM users LIMIT 10;"
Write-Host "  3. Run backend: go run main.go"
