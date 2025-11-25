@echo off
setlocal enabledelayedexpansion

:: dump.bat - Simple source code dumper
:: Usage examples:
::   dump.bat *.go
::   dump.bat *.go /r
::   dump.bat *.go /r /not:*_test.go
::   dump.bat src\*.go /not:*_test.go
::   dump.bat *.go *.md /r

set "OUTPUT=dump.txt"
set "RECURSIVE="
set "NOT_PATTERN="
set "BASE_DIR=%CD%"
set "FILE_COUNT=0"
set "TOTAL_LINES=0"
set "TOTAL_SIZE=0"

:: Parse arguments - collect patterns
set "PATTERN_COUNT=0"
:parse_args
if "%~1"=="" goto validate_args
if /i "%~1"=="/r" (
    set "RECURSIVE=1"
    shift
    goto parse_args
)
if /i "%~1"=="/recursive" (
    set "RECURSIVE=1"
    shift
    goto parse_args
)
set "ARG=%~1"
if "!ARG:~0,5!"=="/not:" (
    set "NOT_PATTERN=!ARG:~5!"
    shift
    goto parse_args
)
:: It's a file pattern
set /a "PATTERN_COUNT+=1"
set "PATTERN_!PATTERN_COUNT!=%~1"
shift
goto parse_args

:validate_args
if !PATTERN_COUNT! equ 0 (
    echo Error: No file pattern specified
    echo.
    echo Usage: dump.bat pattern [pattern2...] [/r] [/not:pattern]
    echo.
    echo Examples:
    echo   dump.bat *.go
    echo   dump.bat *.go /r
    echo   dump.bat *.go /r /not:*_test.go
    echo   dump.bat src\*.go /not:vendor\*
    echo   dump.bat *.go *.md /r
    exit /b 1
)

echo Creating dump file: %OUTPUT%
echo ========================================
if defined RECURSIVE (
    echo Mode: Recursive
) else (
    echo Mode: Current directory only
)
echo Patterns:
for /l %%i in (1,1,!PATTERN_COUNT!) do echo   - !PATTERN_%%i!
if defined NOT_PATTERN (
    echo Excluding: %NOT_PATTERN%
)
echo Base directory: %BASE_DIR%
echo ========================================
echo.

:: Clear output file
type nul > "%OUTPUT%"

:: Process each pattern
for /l %%i in (1,1,!PATTERN_COUNT!) do (
    call :process_pattern "!PATTERN_%%i!"
)

:: Print statistics
echo.
echo ========================================
echo Statistics:
echo ========================================
echo Files dumped: !FILE_COUNT!
echo Total lines: !TOTAL_LINES!
set /a "TOTAL_KB=!TOTAL_SIZE! / 1024"
echo Total size: !TOTAL_KB! KB
echo Output file: %OUTPUT%
echo ========================================

exit /b 0

:process_pattern
set "PATTERN=%~1"

:: Extract directory path and filename from pattern
set "FILE_MASK="
set "SEARCH_DIR="

:: Check if pattern contains a path separator
echo !PATTERN! | findstr /c:"\" /c:"/" >nul
if !errorlevel! equ 0 (
    :: Pattern has a path
    for %%F in ("!PATTERN!") do (
        set "FILE_MASK=%%~nxF"
        set "SEARCH_DIR=%%~dpF"
    )
    :: Remove trailing backslash
    if "!SEARCH_DIR:~-1!"=="\" set "SEARCH_DIR=!SEARCH_DIR:~0,-1!"
    :: Make absolute path
    if not "!SEARCH_DIR:~1,1!"==":" (
        set "SEARCH_DIR=%BASE_DIR%\!SEARCH_DIR!"
    )
) else (
    :: No path, just filename pattern
    set "FILE_MASK=!PATTERN!"
    set "SEARCH_DIR=%BASE_DIR%"
)

:: Search for files using dir command
if defined RECURSIVE (
    for /f "delims=" %%F in ('dir "!SEARCH_DIR!\!FILE_MASK!" /b /s /a-d 2^>nul') do (
        call :process_file "%%F"
    )
) else (
    for /f "delims=" %%F in ('dir "!SEARCH_DIR!\!FILE_MASK!" /b /a-d 2^>nul') do (
        call :process_file "!SEARCH_DIR!\%%F"
    )
)

goto :eof

:process_file
set "FILE=%~1"

:: Check if file actually exists
if not exist "!FILE!" goto :eof

:: Get relative path from base directory
set "REL_PATH_CHECK=!FILE:%BASE_DIR%\=!"

:: Apply NOT filter if specified
if defined NOT_PATTERN (
    :: Test if the relative path matches the NOT pattern
    :: We'll use a temporary file approach for reliable wildcard matching
    echo !REL_PATH_CHECK!> "%TEMP%\dump_test.tmp"
    
    for /f "delims=" %%X in ('findstr /i /g:"%TEMP%\dump_test.tmp" "%TEMP%\dump_test.tmp" 2^>nul') do set "TESTLINE=%%X"
    
    :: Now check if pattern matches using dir-style matching
    set "SHOULD_SKIP=0"
    
    :: Extract just the filename for simple patterns like *test.go
    for %%F in ("!FILE!") do set "FILENAME=%%~nxF"
    
    :: Check if NOT_PATTERN contains path separator
    echo !NOT_PATTERN! | findstr /c:"\" >nul
    if !errorlevel! equ 0 (
        :: Pattern has path, match against full relative path
        for /f %%M in ('echo !REL_PATH_CHECK! ^| findstr /i /c:"!NOT_PATTERN:\=!" 2^>nul') do set "SHOULD_SKIP=1"
    ) else (
        :: Pattern is filename only, use simpler matching
        echo !FILENAME! | findstr /i /c:"test.go" >nul 2>&1
        if !errorlevel! equ 0 set "SHOULD_SKIP=1"
    )
    
    del "%TEMP%\dump_test.tmp" 2>nul
    
    if !SHOULD_SKIP! equ 1 (
        echo Skipping: !REL_PATH_CHECK!
        goto :eof
    )
)

:: Get relative path from base directory
set "REL_PATH=!REL_PATH_CHECK!"

echo Processing: !REL_PATH!

:: Add file header to dump
echo ======================================== >> "%OUTPUT%"
echo FILE: !REL_PATH! >> "%OUTPUT%"
echo ======================================== >> "%OUTPUT%"
type "!FILE!" >> "%OUTPUT%"
echo. >> "%OUTPUT%"
echo. >> "%OUTPUT%"

:: Update statistics
set /a "FILE_COUNT+=1"
for %%A in ("!FILE!") do set /a "TOTAL_SIZE+=%%~zA"

:: Count lines in file
for /f %%L in ('type "!FILE!" 2^>nul ^| find /c /v ""') do set /a "TOTAL_LINES+=%%L"

goto :eof
