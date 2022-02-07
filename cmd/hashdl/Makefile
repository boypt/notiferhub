app_name = hashdl
.DEFAULT_GOAL := build

test:
	go test ./... -v

build: 
	GO111MODULE=on GOARCH=386 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o $(app_name)_386

get:
	go mod download

run:
	go run .

clean:
	rm -f $(app_name)
