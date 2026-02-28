---
sidebar_position: 1
---

# Installation

This guide explains how to install `cogeni` on your system.

## Download Binary

We provide pre-compiled binaries for major operating systems and architectures. You can download the latest version from the releases page (coming soon).

## Install via `go install`

If you have Go 1.25 or higher installed, you can build and install `cogeni` directly from the source repository:

```bash
go install github.com/lucasfabre/codegen/src@latest
```

This will download the source, compile it, and place the `cogeni` executable in your `$GOPATH/bin` directory (which is usually `~/go/bin`). Ensure that this directory is added to your system's `PATH`.

## Build from Source

You can also clone the repository and build the binary yourself. This is useful if you want to contribute to the project or use the absolute latest, unreleased changes.

### Prerequisites
- [Go](https://golang.org/doc/install) 1.25 or higher
- [Git](https://git-scm.com/downloads)
- [Just](https://github.com/casey/just) (optional, task runner)

### Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/lucasfabre/codegen.git
   cd codegen
   ```

2. Build the binary using `just`:
   ```bash
   just build
   ```
   If you don't have `just` installed, you can run the equivalent go command:
   ```bash
   go build -trimpath -ldflags="-s -w -buildid=" -o cogeni ./src
   ```

3. The compiled binary will be located in the root directory. You can move it to a location in your `PATH`:
   ```bash
   mv cogeni /usr/local/bin/
   ```
