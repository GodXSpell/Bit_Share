@echo off
echo File Share System - Test Scripts
echo.
echo Available commands:
echo 1. test-receive.bat - Start receiver on port 9000
echo 2. test-send.bat - Send test.txt to localhost:9000
echo 3. Or use manual commands:
echo    go run main.go receive received_test.txt 9000
echo    go run main.go send test.txt localhost 9000
echo.
pause
