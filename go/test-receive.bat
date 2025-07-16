@echo off
echo Starting receiver on port 9000...
echo File will be saved as: received_test.txt
echo.
go run main.go receive received_test.txt 9000
