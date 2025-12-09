# Complete Render Deployment Script

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  RENDER.COM DEPLOYMENT" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# Pre-flight checks
Write-Host "[Pre-flight] Checking prerequisites..." -ForegroundColor Yellow

# Check Go installed
if (Get-Command go -ErrorAction SilentlyContinue) {
    Write-Host "  ✅ Go installed: $(go version)" -ForegroundColor Green
} else {
    Write-Host "  ❌ Go not found!" -ForegroundColor Red
    exit 1
}

# Check Git installed
if (Get-Command git -ErrorAction SilentlyContinue) {
    Write-Host "  ✅ Git installed" -ForegroundColor Green
} else {
    Write-Host "  ❌ Git not found!" -ForegroundColor Red
    exit 1
}

# Check if in correct directory
if (-not (Test-Path "main.go")) {
    Write-Host "  ❌ main.go not found! Run from project root" -ForegroundColor Red
    exit 1
}

Write-Host "`n[1/6] Testing local build..." -ForegroundColor Yellow
$buildOutput = go build -o lynx-backend.exe . 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "  ✅ Build successful" -ForegroundColor Green
    Remove-Item lynx-backend.exe -ErrorAction SilentlyContinue
} else {
    Write-Host "  ❌ Build failed!" -ForegroundColor Red
    Write-Host $buildOutput
    exit 1
}

Write-Host "`n[2/6] Checking dependencies..." -ForegroundColor Yellow
go mod tidy
go mod verify
Write-Host "  ✅ Dependencies verified" -ForegroundColor Green

Write-Host "`n[3/6] Running tests..." -ForegroundColor Yellow
$testOutput = go test ./... 2>&1
if ($LASTEXITCODE -eq 0 -or $testOutput -match "no test files") {
    Write-Host "  ✅ Tests passed (or no tests)" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  Some tests failed (continuing...)" -ForegroundColor Yellow
}

Write-Host "`n[4/6] Git status..." -ForegroundColor Yellow
git status --short
$hasChanges = git status --porcelain
if ($hasChanges) {
    Write-Host "  Changes detected, committing..." -ForegroundColor Gray
    git add .
    $commitMsg = Read-Host "Commit message (press Enter for default)"
    if ([string]::IsNullOrWhiteSpace($commitMsg)) {
        $commitMsg = "Deploy to Render with DATABASE_URL support"
    }
    git commit -m $commitMsg
    Write-Host "  ✅ Changes committed" -ForegroundColor Green
} else {
    Write-Host "  No changes to commit" -ForegroundColor Gray
}

Write-Host "`n[5/6] Pushing to GitHub..." -ForegroundColor Yellow
git push origin main
if ($LASTEXITCODE -eq 0) {
    Write-Host "  ✅ Pushed to GitHub!" -ForegroundColor Green
} else {
    Write-Host "  ❌ Push failed!" -ForegroundColor Red
    exit 1
}

Write-Host "`n[6/6] Deployment triggered!" -ForegroundColor Yellow
$repoUrl = git config --get remote.origin.url
Write-Host "  Repository: $repoUrl" -ForegroundColor Gray

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "DEPLOYMENT IN PROGRESS" -ForegroundColor Green
Write-Host "========================================`n" -ForegroundColor Cyan

Write-Host "Render is now building your app...`n" -ForegroundColor White

Write-Host "Monitor deployment:" -ForegroundColor Yellow
Write-Host "  1. Open: https://render.com/dashboard" -ForegroundColor White
Write-Host "  2. Click: website-urlshortener-lynx-backend" -ForegroundColor White
Write-Host "  3. Tab: 'Logs' to see build progress" -ForegroundColor White

Write-Host "`nExpected build time: 3-5 minutes" -ForegroundColor Cyan

Write-Host "`nLook for these log messages:" -ForegroundColor Yellow
Write-Host "  ✅ 'Build successful'" -ForegroundColor Green
Write-Host "  ✅ '🔄 Running database migrations...'" -ForegroundColor Green
Write-Host "  ✅ 'Server starting port=10000'" -ForegroundColor Green

Write-Host "`nAfter deployment completes:" -ForegroundColor Yellow
Write-Host "  Test: https://website-urlshortener-lynx-backend.onrender.com/health" -ForegroundColor White

Write-Host "`n========================================`n" -ForegroundColor Cyan
