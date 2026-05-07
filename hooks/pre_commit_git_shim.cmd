@echo off
set COMMAND=%1
shift
if "%COMMAND%" == "diff" goto allowed
if "%COMMAND%" == "grep" goto allowed
if "%COMMAND%" == "ls-files" goto allowed
if "%COMMAND%" == "rev-list" goto allowed
if "%COMMAND%" == "rev-parse" goto allowed
if "%COMMAND%" == "show" goto allowed
if "%COMMAND%" == "status" goto allowed
echo git is not allowed in parallel hooks (git %COMMAND% %*)
exit /b 1
:allowed
"ACTUAL_GIT" %COMMAND% %*
