SRC_FILES=$(shell ls -1 *.go | grep -v _test.go)
BIN="bin/$(shell basename "$(shell pwd)")"

all: linux windows

linux:
	@mkdir bin/ 2>/dev/null || true
	@go build -v -o "$(BIN)"
	@strip -s "$(BIN)"

windows:
	@GOOS=windows go build -v -o "$(BIN).exe"

clean:
	@go clean
	@rm -Rf bin/*

run:
	@go run $(SRC_FILES)

.PHONY: all clean run
