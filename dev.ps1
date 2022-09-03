# Requires nodemon to be installed (npm i -g nodemon)
nodemon --watch './**/*.go' --signal SIGTERM --exec 'go run .\main.go || exit 1'
