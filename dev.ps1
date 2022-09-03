# Requires nodemon to be installed (npm i -g nodemon)
nodemon --watch './**/*.*' --signal SIGTERM --exec 'go run .\hwinfo64-telegraf-plugin.go'
