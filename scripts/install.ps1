# install.ps1 — installs mcp-postman from GitHub Releases on Windows
# Usage: irm https://raw.githubusercontent.com/jinkp/postman-go-mcp/main/scripts/install.ps1 | iex
#
# Options (set before running):
#   $env:INSTALL_DIR = "C:\custom\path"   # default: $env:LOCALAPPDATA\mcp-postman

$ErrorActionPreference = "Stop"

$Repo    = "jinkp/postman-go-mcp"
$Binary  = "mcp-postman"
$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { "$env:LOCALAPPDATA\mcp-postman" }

# ── helpers ────────────────────────────────────────────────────────────────────

function Write-Ok($msg)   { Write-Host "  " -NoNewline; Write-Host "✔" -ForegroundColor Green -NoNewline; Write-Host " $msg" }
function Write-Warn($msg) { Write-Host "  " -NoNewline; Write-Host "⚠" -ForegroundColor Yellow -NoNewline; Write-Host " $msg" }
function Write-Fail($msg) { Write-Host "  " -NoNewline; Write-Host "✗" -ForegroundColor Red -NoNewline; Write-Host " $msg"; exit 1 }

# ── fetch latest version ───────────────────────────────────────────────────────

function Get-LatestVersion {
    $apiUrl  = "https://api.github.com/repos/$Repo/releases/latest"
    $headers = @{ "User-Agent" = "mcp-postman-installer" }
    try {
        $resp = Invoke-RestMethod -Uri $apiUrl -Headers $headers
        return $resp.tag_name
    } catch {
        Write-Fail "Could not fetch latest version: $_"
    }
}

# ── download and install ───────────────────────────────────────────────────────

function Install-Binary($version) {
    $asset       = "${Binary}_windows_amd64"
    $archiveName = "${asset}.zip"
    $downloadUrl = "https://github.com/$Repo/releases/download/$version/$archiveName"
    $checksumUrl = "https://github.com/$Repo/releases/download/$version/checksums.txt"

    $tmpDir = [System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), [System.Guid]::NewGuid().ToString())
    New-Item -ItemType Directory -Path $tmpDir | Out-Null

    try {
        Write-Host "  Downloading $Binary $version (windows_amd64)..."

        $archivePath = Join-Path $tmpDir $archiveName
        Invoke-WebRequest -Uri $downloadUrl -OutFile $archivePath -UseBasicParsing

        # Verify checksum (best effort)
        try {
            $checksumPath = Join-Path $tmpDir "checksums.txt"
            Invoke-WebRequest -Uri $checksumUrl -OutFile $checksumPath -UseBasicParsing
            $expected = (Get-Content $checksumPath | Select-String $archiveName) -replace "\s+.*", "" -replace ".*\s+", ""
            $actual   = (Get-FileHash $archivePath -Algorithm SHA256).Hash.ToLower()
            if ($expected -and $actual -eq $expected) {
                Write-Ok "Checksum verified"
            } else {
                Write-Warn "Could not verify checksum — proceeding anyway"
            }
        } catch {
            Write-Warn "Checksum verification skipped"
        }

        # Extract
        Expand-Archive -Path $archivePath -DestinationPath $tmpDir -Force

        # Ensure install dir exists
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir | Out-Null
        }

        # Copy binary
        $exeName = "${Binary}.exe"
        $src     = Join-Path $tmpDir $exeName
        $dst     = Join-Path $InstallDir $exeName
        Copy-Item -Path $src -Destination $dst -Force

        Write-Ok "Installed: $dst"
        return $dst

    } finally {
        Remove-Item -Recurse -Force $tmpDir -ErrorAction SilentlyContinue
    }
}

# ── add to PATH ────────────────────────────────────────────────────────────────

function Add-ToPath($dir) {
    $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($userPath -notlike "*$dir*") {
        [Environment]::SetEnvironmentVariable("PATH", "$userPath;$dir", "User")
        $env:PATH += ";$dir"
        Write-Ok "Added to PATH (user): $dir"
        Write-Warn "Restart your terminal for PATH changes to take effect"
    } else {
        Write-Ok "Already in PATH: $dir"
    }
}

# ── verify install ─────────────────────────────────────────────────────────────

function Test-Install($exePath) {
    try {
        $ver = & $exePath --version 2>$null
        Write-Ok "Version: $ver"
    } catch {
        Write-Warn "Could not verify installation — try running: $exePath --version"
    }
}

# ── main ───────────────────────────────────────────────────────────────────────

Write-Host ""
Write-Host "  postman-go-mcp installer" -ForegroundColor White
Write-Host ""

$version = Get-LatestVersion
$exePath = Install-Binary $version
Add-ToPath $InstallDir
Test-Install $exePath

Write-Host ""
Write-Host "  Done! " -ForegroundColor Green -NoNewline
Write-Host "Run the setup wizard to configure your AI assistant:"
Write-Host ""
Write-Host "    mcp-postman setup" -ForegroundColor Cyan
Write-Host ""
