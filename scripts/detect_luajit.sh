#!/bin/bash

# Function to generate flags
print_flags() {
	local inc="$1"
	local lib="$2"
	local static="$3"

	echo "export CGO_CFLAGS=\"-I$inc\""

	if [ "$static" = "true" ]; then
		if [ -f "$lib/libluajit-5.1.a" ]; then
			echo "export CGO_LDFLAGS=\"$lib/libluajit-5.1.a -lm -ldl\""
			return
		elif [ -f "$lib/libluajit.a" ]; then
			echo "export CGO_LDFLAGS=\"$lib/libluajit.a -lm -ldl\""
			return
		else
			echo "Warning: Static library not found in $lib, falling back to dynamic" >&2
		fi
	fi

	# Dynamic linking
	if [ -f "$lib/libluajit-5.1.so" ] || [ -f "$lib/libluajit-5.1.dylib" ]; then
		echo "export CGO_LDFLAGS=\"-L$lib -lluajit-5.1 -lm -ldl\""
	elif [ -f "$lib/libluajit.so" ] || [ -f "$lib/libluajit.dylib" ]; then
		echo "export CGO_LDFLAGS=\"-L$lib -lluajit -lm -ldl\""
	else
		# Default assumption
		echo "export CGO_LDFLAGS=\"-L$lib -lluajit-5.1 -lm -ldl\""
	fi
}

STATIC=${1:-false}

# 1. Try environment variables
if [ -n "$LUAJIT_INC" ] && [ -n "$LUAJIT_LIB" ]; then
	print_flags "$LUAJIT_INC" "$LUAJIT_LIB" "$STATIC"
	exit 0
fi

# 2. Try pkg-config (only for dynamic unless we parse --static libs)
if [ "$STATIC" != "true" ] && pkg-config --exists luajit; then
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
	"/usr/include/luajit-2.0 /usr/lib"
	"$HOME/.luajit/include/luajit-2.1 $HOME/.luajit/lib"
	"$(pwd)/.luajit/src $(pwd)/.luajit/src"
)

for p in "${PATHS[@]}"; do
	inc=$(echo "$p" | cut -d' ' -f1)
	lib=$(echo "$p" | cut -d' ' -f2)
	# Check if include directory exists and contains lua.h
	if [ -d "$inc" ] && [ -f "$inc/lua.h" ]; then
		echo "Found LuaJIT at $inc" >&2
		print_flags "$inc" "$lib" "$STATIC"
		exit 0
	fi
done

echo "Error: LuaJIT not found. Please install luajit or set CGO_CFLAGS/CGO_LDFLAGS manually." >&2
exit 1
