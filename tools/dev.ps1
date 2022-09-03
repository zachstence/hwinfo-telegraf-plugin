# Requires nodemon to be installed (npm i -g nodemon)
# From the root of this repo, run .\tools\dev.ps1
nodemon --watch './**/*.go' --signal SIGTERM --exec 'go run .\cmd\main.go || exit 1'
