


.PHONY : all
all : windows darwin linux

.PHONY : windows
windows : client-windows-amd64	

.PHONY : darwin
darwin : client-darwin-amd64

.PHONY : linux
linux : client-linux-amd64

.PHONY : client-windows-amd64
client-windows-amd64 :
	GOOS=windows GOARCH=amd64 go build -o ./bin/windows/amd64/client ./cmd/client
	mkdir -p ./bin/windows/amd64/tools/
	cp -p -r ./tools/windows/amd64/ ./bin/windows/amd64/tools/

.PHONY : client-darwin-amd64
client-darwin-amd64 :
	GOOS=darwin GOARCH=amd64 go build -o ./bin/darwin/amd64/client ./cmd/client

.PHONY : client-linux-amd64
client-linux-amd64 :
	GOOS=linux GOARCH=amd64 go build -o ./bin/linux/amd64/client ./cmd/client

