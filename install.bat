@echo off
echo INSTALLING NOMAD, please wait
set source=https://github.com/jonathanMelly/nomad/releases/latest/download/nomad-latest-win64.zip
powershell -command "Start-BitsTransfer -Source %source% -Destination nomad.zip"
powershell -command "Expand-Archive -Force nomad.zip ."
del nomad.zip
start cmd /k nomad.exe --help
