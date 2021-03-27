#!/bin/bash
GOOS=windows GOARCH=386 go build -o ./dist/atlasmap.exe cmd/atlasmap.go
GOOS=linux go build -o ./dist/atlasmap cmd/atlasmap.go