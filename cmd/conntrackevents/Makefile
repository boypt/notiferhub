app_name = connevent
.DEFAULT_GOAL := build

test:
	go test ./... -v

build: 
	GO111MODULE=on GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o $(app_name)

buildstatic: get
	go build -trimpath -ldflags "-s -w -extldflags=-static"  --tags "sqlite_omit_load_extension netgo" -o $(app_name)

get:
	go mod download

run:
	go run .

clean:
	rm -f $(app_name)
