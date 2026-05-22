@echo off
echo Building vfox-gui...

REM Set MinGW gcc path (adjust if your installation path differs)
set MINGW_PATH=C:\Users\Micro\AppData\Local\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin
set PATH=%MINGW_PATH%;%PATH%

REM Enable CGO and set compiler
set CGO_ENABLED=1
set CC=gcc

go build -ldflags="-H windowsgui" -o vfox-gui.exe .

if %ERRORLEVEL% EQU 0 (
    echo Build successful! Run vfox-gui.exe to start.
) else (
    echo Build failed!
    exit /b 1
)
