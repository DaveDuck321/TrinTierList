@echo off
go get -u .\go\
go build .\go\server.go .\go\util.go .\go\rankings.go
.\server.exe localhost:80
