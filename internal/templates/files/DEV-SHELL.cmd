@echo off
setlocal
call "%~dp0config\initialize-toolchain.cmd"
if errorlevel 1 goto error
cd /d "%BLIND_DEV_ROOT%\workspace"
echo Ambiente pronto. Python e Node.js sao geridos pelo mise.
echo Use uv para projetos Python e pnpm para workspaces Node.js.
powershell.exe -NoLogo -NoProfile -NoExit
exit /b 0

:error
echo Nao foi possivel carregar o ambiente.
pause
exit /b 1
