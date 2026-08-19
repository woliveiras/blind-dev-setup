@echo off
setlocal
call "%~dp0config\environment.cmd"
if errorlevel 1 goto error
echo Iniciando DBeaver Community. A acessibilidade desta ferramenta ainda e experimental.
start "" "%BLIND_DEV_ROOT%\tools\dbeaver\dbeaver.exe" -data "%BLIND_DEV_ROOT%\data\dbeaver\workspace"
exit /b 0

:error
echo Nao foi possivel carregar o ambiente.
pause
exit /b 1
