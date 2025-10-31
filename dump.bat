@echo off
setlocal enabledelayedexpansion

REM === Usage: dump.bat <extension> [folder] ===
REM Example: dump.bat go .\internal\command
REM Example: dump.bat txt

REM --- Parse arguments ---
if "%~1"=="" (
    echo Usage: %~nx0 ^<extension^> [folder]
    exit /b 1
)
set "ext=%~1"

if "%~2"=="" (
    set "scanDir=%cd%"
) else (
    set "scanDir=%~2"
)

REM --- Output file ---
set "output=dump"

REM --- Clear previous output if it exists ---
if exist "%output%" del "%output%"

REM --- Recursively scan all *.%ext% files ---
for /r "%scanDir%" %%f in (*.%ext%) do (
    echo --- Contents of file: %%f --- >> "%output%"
    type "%%f" >> "%output%"
    echo. >> "%output%"
)

echo Done! Contents of all .%ext% files in "%scanDir%" have been written to %output%.
