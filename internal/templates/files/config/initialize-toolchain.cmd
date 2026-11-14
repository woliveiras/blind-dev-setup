@echo off
call "%~dp0environment.cmd"
if errorlevel 1 exit /b 1
powershell.exe -NoLogo -NoProfile -NonInteractive -ExecutionPolicy Bypass -File "%~dp0initialize-toolchain.ps1"
if errorlevel 1 exit /b 1
exit /b 0
