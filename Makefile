


.PHONY : all
all : windows darwin linux

.PHONY : windows
windows : client-windows-amd64 server-windows-amd64

.PHONY : darwin
darwin : client-darwin-amd64 server-darwin-amd64

.PHONY : linux
linux : client-linux-amd64 server-linux-amd64 client-linux-mips server-linux-mips

.PHONY : client-windows-amd64
client-windows-amd64 :
	GOOS=windows GOARCH=amd64 go build -o ./bin/windows/amd64/client ./cmd/client

.PHONY : server-windows-amd64
server-windows-amd64 :
	GOOS=windows GOARCH=amd64 go build -o ./bin/windows/amd64/server ./cmd/server

.PHONY : client-darwin-amd64
client-darwin-amd64 :
	GOOS=darwin GOARCH=amd64 go build -o ./bin/darwin/amd64/client ./cmd/client

.PHONY : server-darwin-amd64
server-darwin-amd64 :
	GOOS=darwin GOARCH=amd64 go build -o ./bin/darwin/amd64/server ./cmd/server

.PHONY : client-linux-amd64
client-linux-amd64 :
	GOOS=linux GOARCH=amd64 go build -o ./bin/linux/amd64/client ./cmd/client

.PHONY : server-linux-amd64
server-linux-amd64 :
	GOOS=linux GOARCH=amd64 go build -o ./bin/linux/amd64/server ./cmd/server

.PHONY : client-linux-mips
client-linux-mips :
	GOOS=linux GOARCH=mips go build -o ./bin/linux/mips/client ./cmd/client

.PHONY : server-linux-mips
server-linux-mips :
	GOOS=linux GOARCH=mips go build -o ./bin/linux/mips/server ./cmd/server


