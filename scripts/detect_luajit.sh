#!/usr/bin/env bash

# Resolve the absolute path of the vendor directory
VENDOR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../vendor" && pwd)"
LUAJIT_INCLUDE="${VENDOR_DIR}/luajit/include/luajit-2.1"
LUAJIT_LIB="${VENDOR_DIR}/luajit/lib"

# 1. Try vendor directory first
if [ -d "$LUAJIT_INCLUDE" ] && [ -d "$LUAJIT_LIB" ]; then
	echo "export CGO_CFLAGS=\"-I${LUAJIT_INCLUDE}\""
	echo "export CGO_LDFLAGS=\"${LUAJIT_LIB}/libluajit-5.1.a -lm -ldl\""
	exit 0
fi

# 2. Try pkg-config
if pkg-config --exists luajit; then
	echo "export CGO_CFLAGS=\"$(pkg-config --cflags luajit)\""
	echo "export CGO_LDFLAGS=\"$(pkg-config --libs luajit)\""
	exit 0
fi

# 3. Search common paths
PATHS=(
	"/usr/local/include/luajit-2.1 /usr/local/lib"
	"/opt/homebrew/include/luajit-2.1 /opt/homebrew/lib"
	"/home/linuxbrew/.linuxbrew/include/luajit-2.1 /home/linuxbrew/.linuxbrew/lib"
	"/usr/include/luajit-2.1 /usr/lib"
)

for p in "${PATHS[@]}"; do
	inc=$(echo "$p" | cut -d' ' -f1)
	lib=$(echo "$p" | cut -d' ' -f2)
	if [ -d "$inc" ] && [ -f "$inc/lua.h" ]; then
		echo "export CGO_CFLAGS=\"-I$inc\""
		echo "export CGO_LDFLAGS=\"-L$lib -lluajit-5.1 -lm -ldl\""
		exit 0
	fi
done

echo "Error: LuaJIT not found. Please run 'just vendor-pull' or install luajit manually." >&2
exit 1
