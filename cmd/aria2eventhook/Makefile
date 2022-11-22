app_name = aria2hook
.DEFAULT_GOAL := build

test:
	go test ./... -v

build: 
	GO111MODULE=on GOARCH=mipsle GOMIPS=softfloat CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o $(app_name)

get:
	go mod download

run:
	go run .

clean:
	rm -f $(app_name)
