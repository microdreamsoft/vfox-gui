@echo off
echo Building vfox-gui...

REM Check if gcc is available
where gcc >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: gcc not found in PATH.
    echo Please install MinGW and add to system PATH, or modify this script to set MINGW_PATH.
    exit /b 1
)

REM Enable CGO and set compiler
set CGO_ENABLED=1
set CC=gcc

echo go build...
go build -ldflags="-H windowsgui" -o vfox-gui.exe .

if %ERRORLEVEL% EQU 0 (
    echo Build successful! Run vfox-gui.exe to start.
) else (
    echo Build failed!
    exit /b 1
)
