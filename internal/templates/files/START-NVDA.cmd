@echo off
setlocal
call "%~dp0config\environment.cmd"
if errorlevel 1 goto error
echo Iniciando NVDA Portable.
start "" "%BLIND_DEV_ROOT%\tools\nvda\nvda.exe"
exit /b 0

:error
echo Nao foi possivel carregar o ambiente.
pause
exit /b 1
