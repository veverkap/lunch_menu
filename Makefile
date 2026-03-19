.PHONY: all build

# Build everything: frontend then Go binary
all: build

# Build the Go binary
build:
	go build -o lunch_menu ./cmd/cli