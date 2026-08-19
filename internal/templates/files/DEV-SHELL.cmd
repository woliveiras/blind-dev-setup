@echo off
setlocal
call "%~dp0config\environment.cmd"
if errorlevel 1 goto error
"%BLIND_DEV_ROOT%\tools\mise\mise.exe" trust --yes "%MISE_CONFIG_FILE%" >nul
if errorlevel 1 goto trust_error
"%BLIND_DEV_ROOT%\tools\mise\mise.exe" reshim
if errorlevel 1 goto mise_error
cd /d "%BLIND_DEV_ROOT%\workspace"
echo Ambiente pronto. Python e Node.js sao geridos pelo mise.
echo Use uv para projetos Python e pnpm para workspaces Node.js.
powershell.exe -NoLogo -NoExit
exit /b 0

:trust_error
echo O mise nao conseguiu confiar na configuracao fixada deste pendrive.
pause
exit /b 1

:mise_error
echo O mise nao conseguiu preparar os comandos do ambiente.
pause
exit /b 1

:error
echo Nao foi possivel carregar o ambiente.
pause
exit /b 1
