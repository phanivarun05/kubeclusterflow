.PHONY: build build-ui copy-ui build-go clean

# **Main build target** that runs the entire pipeline
build: build-ui copy-ui build-go

# 1. Build your React/Vite frontend
build-ui:
	cd ui && npm run build

# 2. Copy the output into the Go package directory so go:embed can see it
copy-ui:
	rm -rf internal/api/ui/dist
	cp -r ui/dist internal/api/ui/dist

# 3. Build the Go binary
build-go:
	go build -o bin/kubeclusterflow main.go

# Clean up build artifacts
clean:
	rm -rf bin/
	rm -rf internal/api/ui/dist
	rm -rf ui/dist
