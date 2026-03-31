@echo off
title Morris Family Recipe Tracker
echo 1. Checking for updates...
go build -o morris-tracker.exe .
echo.
echo 2. Launching Kitchen Server...
echo.
cls
morris-tracker.exe
pause