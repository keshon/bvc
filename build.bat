@echo off
rem
rem BUILD
rem

go run cmd\generate_readme\generate_readme.go

rem Build command
go build -o temp\bvc.exe cmd\bvc\main.go
