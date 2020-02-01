go get -u ./go/
go build ./go/server.go ./go/util.go ./go/rankings.go
setcap "cap_net_bind_service=+ep" ./server
./server
