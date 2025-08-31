# PeerVault Build Script for Windows
# Usage: .\scripts\build.ps1 [target]

param(
    [Parameter(Position=0)]
    [ValidateSet("all", "main", "node", "demo", "clean")]
    [string]$Target = "all"
)

Write-Host "PeerVault Build Script for Windows" -ForegroundColor Green
Write-Host "Target: $Target" -ForegroundColor Yellow

# Create bin directory if it doesn't exist
if (!(Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
    Write-Host "Created bin directory" -ForegroundColor Blue
}

switch ($Target) {
    "main" {
        Write-Host "Building main application..." -ForegroundColor Blue
        go build -o bin\peervault.exe .\cmd\peervault
        if ($LASTEXITCODE -eq 0) {
            Write-Host "Main application built successfully" -ForegroundColor Green
        } else {
            Write-Host "Failed to build main application" -ForegroundColor Red
            exit 1
        }
    }
    "node" {
        Write-Host "Building node binary..." -ForegroundColor Blue
        go build -o bin\peervault-node.exe .\cmd\peervault-node
        if ($LASTEXITCODE -eq 0) {
            Write-Host "Node binary built successfully" -ForegroundColor Green
        } else {
            Write-Host "Failed to build node binary" -ForegroundColor Red
            exit 1
        }
    }
    "demo" {
        Write-Host "Building demo client..." -ForegroundColor Blue
        go build -o bin\peervault-demo.exe .\cmd\peervault-demo
        if ($LASTEXITCODE -eq 0) {
            Write-Host "Demo client built successfully" -ForegroundColor Green
        } else {
            Write-Host "Failed to build demo client" -ForegroundColor Red
            exit 1
        }
    }
    "all" {
        Write-Host "Building all binaries..." -ForegroundColor Blue
        & $PSCommandPath "main"
        & $PSCommandPath "node"
        & $PSCommandPath "demo"
        Write-Host "All binaries built successfully" -ForegroundColor Green
    }
    "clean" {
        Write-Host "Cleaning build artifacts..." -ForegroundColor Blue
        if (Test-Path "bin") {
            Remove-Item -Recurse -Force "bin"
            Write-Host "Removed bin directory" -ForegroundColor Green
        }
        go clean -cache -testcache
        Write-Host "Cleaned Go cache" -ForegroundColor Green
    }
}

Write-Host "Build completed!" -ForegroundColor Green
