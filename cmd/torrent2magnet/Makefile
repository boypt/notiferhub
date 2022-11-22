app_name = torrent2magnet_windows.exe
.DEFAULT_GOAL := build

test:
	go test ./... -v

build: 
	GOOS=windows GOARCH=amd64 go build -o $(app_name) -ldflags "-s -w -X main.VERSION=$(shell git rev-parse --short HEAD)"

get:
	go mod download

run:
	go run .

clean:
	rm -f $(app_name)
