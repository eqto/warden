#!/bin/bash

GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o dist/warden cmd/warden/*.go

upx dist/warden

