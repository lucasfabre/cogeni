# Justfile

run args="":
	@eval $(./scripts/detect_luajit.sh) && go run -mod=mod ./src {{args}}

vendor-pull:
	@test -f vendor/luajit/lib/libluajit-5.1.a || (chmod +x scripts/vendor_luajit.sh && ./scripts/vendor_luajit.sh)

build: vendor-pull
	@eval $(./scripts/detect_luajit.sh) && go build -mod=mod -trimpath -ldflags="-s -w -buildid=" -o cogeni ./src

build-static: vendor-pull
	@eval $(./scripts/detect_luajit.sh) && go build -mod=mod -trimpath -ldflags="-s -w -buildid=" -o cogeni ./src

# Build optimized binary for current platform
release: build-static
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
			go build -mod=mod -trimpath -buildvcs=false -ldflags "-s -w -buildid=" -o /src/dist/cogeni-linux-amd64 ./src && \
			echo "=== Building Linux ARM64 ===" && \
			GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ \
			go build -mod=mod -trimpath -buildvcs=false -ldflags "-s -w -buildid=" -o /src/dist/cogeni-linux-arm64 ./src'
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
			go build -mod=mod -trimpath -buildvcs=false -ldflags "-s -w -buildid=" -o /src/dist/cogeni-darwin-amd64 ./src && \
			echo "=== Building Darwin ARM64 ===" && \
			GOOS=darwin GOARCH=arm64 CC=oa64-clang CXX=oa64-clang++ \
			go build -mod=mod -trimpath -buildvcs=false -ldflags "-s -w -buildid=" -o /src/dist/cogeni-darwin-arm64 ./src'
	echo "Note: macOS binaries are not compressed (UPX has limited macOS support)"
	ls -lh dist/

# Cross-compile all platforms using goreleaser-cross Docker image
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
			go build -mod=mod -trimpath -buildvcs=false -ldflags "-s -w -buildid=" -o /src/dist/cogeni-linux-amd64 ./src && \
			echo "=== Building Linux ARM64 ===" && \
			GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ \
			go build -mod=mod -trimpath -buildvcs=false -ldflags "-s -w -buildid=" -o /src/dist/cogeni-linux-arm64 ./src && \
			echo "=== Building Darwin AMD64 ===" && \
			GOOS=darwin GOARCH=amd64 CC=o64-clang CXX=o64-clang++ \
			go build -mod=mod -trimpath -buildvcs=false -ldflags "-s -w -buildid=" -o /src/dist/cogeni-darwin-amd64 ./src && \
			echo "=== Building Darwin ARM64 ===" && \
			GOOS=darwin GOARCH=arm64 CC=oa64-clang CXX=oa64-clang++ \
			go build -mod=mod -trimpath -buildvcs=false -ldflags "-s -w -buildid=" -o /src/dist/cogeni-darwin-arm64 ./src && \
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
	go vet -mod=mod ./...
	mise exec -- golangci-lint run ./...
	just fmt

test-lua: build
	./cogeni run tests/lua/framework/test.lua

test: build
	go test -mod=mod ./...
	./tests/shell/framework/run_all_tests.sh

build-docs: build
	./cogeni run cogeni.lua
	python3 ./scripts/build_man_pages_md.py
	./scripts/build_api_docs_md.sh
	./scripts/build_site.sh

profile-perf: build
	mkdir -p profiles
	@echo "Generating stress test environment..."
	./cogeni run examples/performance/generate.lua
	@echo "Running stress test with CPU profiling..."
	CPUPROFILE=profiles/perf_cpu.prof ./cogeni run examples/performance/cogeni.lua
	@echo "Opening pprof..."
	go tool pprof -http=:8080 profiles/perf_cpu.prof
