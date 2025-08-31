@echo off
REM PeerVault Quick Trust Script
REM This script adds basic Windows Defender exclusions for PeerVault

echo PeerVault Quick Trust Setup
echo ===========================
echo.

REM Check if running as Administrator
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo ERROR: This script must be run as Administrator!
    echo Please right-click this file and select "Run as Administrator"
    pause
    exit /b 1
)

echo Adding Windows Defender exclusions...

REM Get current directory
set "PROJECT_PATH=%~dp0.."
set "PROJECT_PATH=%PROJECT_PATH:~0,-1%"

echo Project path: %PROJECT_PATH%

REM Add folder exclusion using PowerShell
powershell -Command "Add-MpPreference -ExclusionPath '%PROJECT_PATH%' -ErrorAction SilentlyContinue"
if %errorLevel% equ 0 (
    echo ✓ Added folder exclusion
) else (
    echo ⚠ Could not add folder exclusion
)

REM Add process exclusions
powershell -Command "Add-MpPreference -ExclusionProcess 'peervault.exe' -ErrorAction SilentlyContinue"
powershell -Command "Add-MpPreference -ExclusionProcess 'peervault-node.exe' -ErrorAction SilentlyContinue"
powershell -Command "Add-MpPreference -ExclusionProcess 'peervault-demo.exe' -ErrorAction SilentlyContinue"
powershell -Command "Add-MpPreference -ExclusionProcess 'go.exe' -ErrorAction SilentlyContinue"
echo ✓ Added process exclusions

echo.
echo Setup complete! Windows Defender should now trust PeerVault applications.
echo.
echo If you still get warnings:
echo 1. Open Windows Security
echo 2. Go to Virus & threat protection
echo 3. Click "Manage settings"
echo 4. Add manual exclusions for this folder
echo.
pause
