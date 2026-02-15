# Justfile

run args="":
	go run ./src {{args}}

build:
	go build -trimpath -ldflags="-s -w -buildid=" -o cogeni ./src

# Build optimized binary for current platform
release:
	go build -trimpath -ldflags="-s -w -buildid=" -o cogeni ./src
	mise exec -- upx --best --lzma cogeni || true

# Build only Linux binaries using goreleaser-cross
release-linux:
	#!/bin/bash
	set -e
	mkdir -p dist .cache/go/cache .cache/go/mod
	echo "Building Linux binaries with goreleaser-cross..."
	docker run \
		--user $(id -u):$(id -g) \
		-v "{{justfile_directory()}}:/src" \
		-v "{{justfile_directory()}}/.cache/go:/go" \
		-w /src \
		-e CGO_ENABLED=1 \
		-e HOME=/tmp \
		-e GOCACHE=/go/cache \
		-e GOMODCACHE=/go/mod \
		--entrypoint /bin/bash \
		ghcr.io/goreleaser/goreleaser-cross:v1.25.0 \
		-c 'set -e && \
			echo "=== Building Linux AMD64 ===" && \
			GOOS=linux GOARCH=amd64 CC=x86_64-linux-gnu-gcc CXX=x86_64-linux-gnu-g++ \
			go build -trimpath -buildvcs=false -ldflags "-s -w -buildid=" -o /src/dist/cogeni-linux-amd64 ./src && \
			echo "=== Building Linux ARM64 ===" && \
			GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ \
			go build -trimpath -buildvcs=false -ldflags "-s -w -buildid=" -o /src/dist/cogeni-linux-arm64 ./src'
	echo "=== Compressing binaries with UPX ==="
	mise exec -- upx --best --lzma dist/cogeni-linux-amd64 || true
	mise exec -- upx --best --lzma dist/cogeni-linux-arm64 || true
	ls -lh dist/

# Build only Darwin (macOS) binaries using goreleaser-cross
release-darwin:
	#!/bin/bash
	set -e
	mkdir -p dist .cache/go/cache .cache/go/mod
	echo "Building Darwin binaries with goreleaser-cross..."
	docker run \
		--user $(id -u):$(id -g) \
		-v "{{justfile_directory()}}:/src" \
		-v "{{justfile_directory()}}/.cache/go:/go" \
		-w /src \
		-e CGO_ENABLED=1 \
		-e HOME=/tmp \
		-e GOCACHE=/go/cache \
		-e GOMODCACHE=/go/mod \
		--entrypoint /bin/bash \
		ghcr.io/goreleaser/goreleaser-cross:v1.25.0 \
		-c 'set -e && \
			echo "=== Building Darwin AMD64 ===" && \
			GOOS=darwin GOARCH=amd64 CC=o64-clang CXX=o64-clang++ \
			go build -trimpath -buildvcs=false -ldflags "-s -w -buildid=" -o /src/dist/cogeni-darwin-amd64 ./src && \
			echo "=== Building Darwin ARM64 ===" && \
			GOOS=darwin GOARCH=arm64 CC=oa64-clang CXX=oa64-clang++ \
			go build -trimpath -buildvcs=false -ldflags "-s -w -buildid=" -o /src/dist/cogeni-darwin-arm64 ./src'
	echo "Note: macOS binaries are not compressed (UPX has limited macOS support)"
	ls -lh dist/

# Build only Windows binaries using goreleaser-cross
# NOTE: Windows cross-compilation has limitations with purego/CGO
# For best results, build natively on Windows using: go build -trimpath -ldflags "-s -w -buildid=" -o cogeni.exe ./src
release-windows:
	#!/bin/bash
	set -e
	mkdir -p dist .cache/go/cache .cache/go/mod
	echo "⚠️  Windows cross-compilation may have issues with CGO/purego."
	echo "   For best results, build natively on Windows."
	echo ""
	echo "Attempting Windows build with goreleaser-cross..."
	docker run \
		--user $(id -u):$(id -g) \
		-v "{{justfile_directory()}}:/src" \
		-v "{{justfile_directory()}}/.cache/go:/go" \
		-w /src \
		-e CGO_ENABLED=1 \
		-e HOME=/tmp \
		-e GOCACHE=/go/cache \
		-e GOMODCACHE=/go/mod \
		--entrypoint /bin/bash \
		ghcr.io/goreleaser/goreleaser-cross:v1.25.0 \
		-c 'set -e && \
			echo "=== Building Windows AMD64 ===" && \
			GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ \
			go build -trimpath -buildvcs=false -ldflags "-s -w -buildid=" -o /src/dist/cogeni-windows-amd64.exe ./src' || echo "Windows build failed (expected - build on Windows natively)"
	if [ -f dist/cogeni-windows-amd64.exe ]; then
		echo "=== Compressing binary with UPX ==="
		mise exec -- upx --best --lzma dist/cogeni-windows-amd64.exe || true
	fi
	ls -lh dist/ 2>/dev/null || echo "No Windows binary available"

# Cross-compile all platforms using goreleaser-cross Docker image
# Supports: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64
# NOTE: Windows is skipped due to CGO/purego cross-compilation limitations
release-all:
	#!/bin/bash
	set -e
	mkdir -p dist .cache/go/cache .cache/go/mod
	echo "Building with goreleaser-cross..."
	echo "⚠️  Windows build is skipped - build natively on Windows for best results"
	echo ""
	docker run \
		--user $(id -u):$(id -g) \
		-v "{{justfile_directory()}}:/src" \
		-v "{{justfile_directory()}}/.cache/go:/go" \
		-w /src \
		-e CGO_ENABLED=1 \
		-e HOME=/tmp \
		-e GOCACHE=/go/cache \
		-e GOMODCACHE=/go/mod \
		--entrypoint /bin/bash \
		ghcr.io/goreleaser/goreleaser-cross:v1.25.0 \
		-c 'set -e && \
			echo "=== Building Linux AMD64 ===" && \
			GOOS=linux GOARCH=amd64 CC=x86_64-linux-gnu-gcc CXX=x86_64-linux-gnu-g++ \
			go build -trimpath -buildvcs=false -ldflags "-s -w -buildid=" -o /src/dist/cogeni-linux-amd64 ./src && \
			echo "=== Building Linux ARM64 ===" && \
			GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ \
			go build -trimpath -buildvcs=false -ldflags "-s -w -buildid=" -o /src/dist/cogeni-linux-arm64 ./src && \
			echo "=== Building Darwin AMD64 ===" && \
			GOOS=darwin GOARCH=amd64 CC=o64-clang CXX=o64-clang++ \
			go build -trimpath -buildvcs=false -ldflags "-s -w -buildid=" -o /src/dist/cogeni-darwin-amd64 ./src && \
			echo "=== Building Darwin ARM64 ===" && \
			GOOS=darwin GOARCH=arm64 CC=oa64-clang CXX=oa64-clang++ \
			go build -trimpath -buildvcs=false -ldflags "-s -w -buildid=" -o /src/dist/cogeni-darwin-arm64 ./src && \
			echo "=== All builds complete ==="'
	echo "=== Compressing binaries with UPX ==="
	mise exec -- upx --best --lzma dist/cogeni-linux-amd64 || true
	mise exec -- upx --best --lzma dist/cogeni-linux-arm64 || true
	echo "Note: macOS binaries are not compressed (UPX has limited macOS support)"
	echo "=== Done ===" && ls -lh dist/

# Clean build cache
clean:
	rm -rf .cache

fmt:
	mise exec -- pre-commit run --all-files

lint:
	go vet ./...
	mise exec -- golangci-lint run ./...
	just fmt

test-lua: build
	./cogeni run tests/lua/framework/test.lua

test: build
	go test ./...
	./tests/shell/framework/run_all_tests.sh
