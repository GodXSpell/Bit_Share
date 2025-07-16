@echo off
echo Sending test.txt to localhost:9000...
echo Make sure receiver is running first!
echo.
timeout /t 2 /nobreak >nul
go run main.go send test.txt localhost 9000
