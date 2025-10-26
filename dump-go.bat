@echo off
setlocal enabledelayedexpansion

REM Optional: first parameter is the folder to scan
if "%~1"=="" (
    set "scanDir=%cd%"
) else (
    set "scanDir=%~1"
)

REM Output file name
set "output=dump-go.txt"

REM Clear the output file if it exists
if exist "%output%" del "%output%"

REM Recursively scan all *.go files
for /r "%scanDir%" %%f in (*.go) do (
    echo --- Contents of file: %%f --- >> "%output%"
    type "%%f" >> "%output%"
    echo. >> "%output%"
)

echo Done! Contents of all .go files in "%scanDir%" have been written to %output%.
pause
