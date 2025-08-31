# PeerVault Run Script for Windows
# Usage: .\scripts\run.ps1 [target] [args...]

param(
    [Parameter(Position=0)]
    [ValidateSet("main", "node", "demo")]
    [string]$Target = "main",
    
    [Parameter(Position=1, ValueFromRemainingArguments=$true)]
    [string[]]$Args
)

Write-Host "PeerVault Run Script for Windows" -ForegroundColor Green
Write-Host "Target: $Target" -ForegroundColor Yellow

switch ($Target) {
    "main" {
        Write-Host "Running main application (all-in-one)..." -ForegroundColor Blue
        go run .\cmd\peervault
    }
    "node" {
        $nodeArgs = if ($Args.Count -gt 0) { $Args -join " " } else { "--listen :3000" }
        Write-Host "Running individual node with args: $nodeArgs" -ForegroundColor Blue
        go run .\cmd\peervault-node $nodeArgs
    }
    "demo" {
        $demoArgs = if ($Args.Count -gt 0) { $Args -join " " } else { "--target localhost:5000" }
        Write-Host "Running demo client with args: $demoArgs" -ForegroundColor Blue
        go run .\cmd\peervault-demo $demoArgs
    }
}

if ($LASTEXITCODE -ne 0) {
    Write-Host "Application exited with code: $LASTEXITCODE" -ForegroundColor Red
    exit $LASTEXITCODE
}
