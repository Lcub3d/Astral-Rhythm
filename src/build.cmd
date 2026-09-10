@echo off
setlocal
pushd "%~dp0"
where go >nul 2>nul
if errorlevel 1 (
  echo Go 1.23 or newer is required only to rebuild from source.
  echo Normal users can run the supplied AstralRhythm.exe directly.
  pause
  exit /b 1
)
go test -v ./...
if errorlevel 1 goto failed
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
go build -trimpath -ldflags="-H=windowsgui -s -w" -o "..\AstralRhythm.exe" .
if errorlevel 1 goto failed
echo Build succeeded: ..\AstralRhythm.exe
popd
pause
exit /b 0
:failed
echo Build failed. See the error above.
popd
pause
exit /b 1
