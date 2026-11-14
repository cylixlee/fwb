.PHONY: build install check clean

build: clean
	@mkdir build
	@go build -o build ./cmd/fwb

install:
	@go install ./cmd/fwb

check:
	@go build ./...
	@go fix ./...
	@go fmt ./...
	@go vet ./...
	@go test ./...

clean:
	@rm -rf build