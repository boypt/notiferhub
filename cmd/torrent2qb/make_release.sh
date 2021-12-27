#!/bin/bash


BIN=torrent2qb
SUFFIX=
GITVER=$(git rev-parse --short HEAD)

GOOS=linux
GOARCH=amd64
if [[ $1 == "windows" ]]; then
  SUFFIX=.exe
  GOOS=windows
fi

rm -fv ${BIN}_*
env GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${BIN}_${GOOS}${SUFFIX} -ldflags "-s -w -X main.VERSION=$GITVER"
