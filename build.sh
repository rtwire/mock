#!/bin/bash

PACKAGE=github.com/rtwire/mock

env GOOS=windows GOARCH=386 go build -v -o windows/386/mock.exe $PACKAGE 
env GOOS=linux GOARCH=amd64 go build -v -o linux/amd64/mock $PACKAGE
env GOOS=darwin GOARCH=amd64 go build -v -o darwin/amd64/mock $PACKAGE

zip mock-windows-386.zip windows/386/mock.exe
zip mock-linux-amd64.zip linux/amd64/mock
zip mock-darwin-amd64.zip darwin/amd64/mock
