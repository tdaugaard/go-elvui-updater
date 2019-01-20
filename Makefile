SRC_FILES=$(shell ls -1 *.go | grep -v _test.go)
BIN="bin/$(shell basename "$(shell pwd)")"
LINUX_PLATFORM="linux-$(shell uname -p)"

all: linux windows

linux:
	@mkdir bin/ 2>/dev/null || true
	@GOOS=linux GOARCH=amd64 go build -v -o "$(BIN)-$(LINUX_PLATFORM)"
	@strip -s "$(BIN)-$(LINUX_PLATFORM)"

windows:
	@GOOS=windows GOARCH=amd64 go build -v -o "$(BIN)-windows-amd64.exe"

clean:
	@go clean
	@rm -Rf bin/*

run:
	@go run $(SRC_FILES)

.PHONY: all clean run
