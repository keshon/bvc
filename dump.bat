@echo off
setlocal enabledelayedexpansion

REM --- Parse arguments ---
set "EXT=%~1"
if "%EXT%"=="" goto :usage

set "DIR=%~2"
set "FILTER=%~3"
if "%DIR%"=="" set "DIR=."

REM remove leading dot from extension if present
if "%EXT:~0,1%"=="." set "EXT=%EXT:~1%"

REM --- Determine inclusion/exclusion ---
set "EXCLUDE=0"
if defined FILTER (
    if "%FILTER:~0,1%"=="!" (
        set "EXCLUDE=1"
        set "FILTER=%FILTER:~1%"
    )
)

REM --- Prepare output file ---
set "OUTFILE=dump.txt"
if exist "%OUTFILE%" del "%OUTFILE%"

REM --- Scan files recursively ---
for /r "%DIR%" %%F in (*.%EXT%) do (
    set "FNAME=%%~nxF"
    set "FPATH=%%F"
    
    REM --- Apply filter if set ---
    set "SKIP=0"
    if defined FILTER (
        echo !FNAME! | findstr /i /c:"%FILTER%" >nul
        if !ERRORLEVEL! equ 0 (
            if !EXCLUDE! equ 1 set "SKIP=1"
        ) else (
            if !EXCLUDE! equ 0 set "SKIP=1"
        )
    )

    if !SKIP! equ 0 (
        REM --- Get relative path ---
        set "REL=%%F"
        set "REL=!REL:%CD%\=!"
        set "REL=!REL:\=/!"  REM replace backslashes with slashes

        echo --- file: !REL! --- >> "%OUTFILE%"
        type "%%F" >> "%OUTFILE%"
        echo. >> "%OUTFILE%"
    )
)

echo Done. File contents written to "%OUTFILE%"
goto :eof

:usage
echo(
echo Usage:
echo   %~nx0 ^<extension^> [folder] [filter]
echo(
echo Examples:
echo   %~nx0 go .\internal\command *test*       <-- only test files
echo   %~nx0 go .\internal\command ^!*test*      <-- exclude test files
echo   %~nx0 txt                                 <-- all .txt files in current dir
echo(
echo Notes:
echo   - [filter] supports substring matching (case-insensitive)
echo   - Prefix with "!" to exclude files instead of include
echo   - Output is written to "dump.txt" in the current directory
exit /b 0
