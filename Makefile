.PHONY: build test format clean

build:
	go build -o build/godot-mcp-go ./cmd/godot-mcp-go

test:
	go test -v ./...

format:
	gofmt -w .

clean:
	rm -rf build/
