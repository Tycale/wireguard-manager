.PHONY: build clean fmt run debug rundebug install

# Variables
BINARY = wgm
APP_DIR = app
GCFLAGS =

# Targets
build: fmt
	@echo "Building the project with flags $(GCFLAGS)..."
	@cd $(APP_DIR) && go build $(GCFLAGS) -o ../$(BINARY)
	@echo "Build complete!"

debug: GCFLAGS=-gcflags='-N -l'
debug: build

fmt:
	@echo "Formatting the Go code..."
	@cd $(APP_DIR) && go fmt ./...

clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY)
	@echo "Clean complete!"

run: build
	@echo "Running $(BINARY)..."
	@./$(BINARY)

# go install github.com/go-delve/delve/cmd/dlv
rundebug: debug
	@echo "Running $(BINARY) with Delve..."
	@dlv exec ./$(BINARY)

install: build
	install -m 557 $(BINARY) /usr/local/bin/
