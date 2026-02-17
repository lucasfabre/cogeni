#!/usr/bin/env bash
set -e

# LuaJIT version/branch
LUAJIT_VERSION="v2.1"
LUAJIT_URL="https://github.com/LuaJIT/LuaJIT/archive/refs/heads/${LUAJIT_VERSION}.tar.gz"

VENDOR_DIR="vendor"
SRC_DIR="${VENDOR_DIR}/src/luajit"
INSTALL_DIR="${VENDOR_DIR}/luajit"

# Ensure vendor directory exists
mkdir -p "${VENDOR_DIR}/src"

# check if already installed
if [ -f "${INSTALL_DIR}/lib/libluajit-5.1.a" ]; then
	echo "LuaJIT already vendored in ${INSTALL_DIR}"
	exit 0
fi

echo "Downloading LuaJIT ${LUAJIT_VERSION}..."
curl -L "${LUAJIT_URL}" -o "${VENDOR_DIR}/luajit.tar.gz"

echo "Extracting LuaJIT..."
mkdir -p "${SRC_DIR}"
tar -xzf "${VENDOR_DIR}/luajit.tar.gz" -C "${SRC_DIR}" --strip-components=1

echo "Building LuaJIT..."
cd "${SRC_DIR}"

# Clean previous build
make clean

# Build static library only (BUILDMODE=static is default for libluajit.a)
# MACOSX_DEPLOYMENT_TARGET is often needed on macOS for compatibility
if [[ $OSTYPE == "darwin"* ]]; then
	export MACOSX_DEPLOYMENT_TARGET=11.0
fi

make amalg

echo "Installing LuaJIT to ${INSTALL_DIR}..."
make install PREFIX="$(pwd)/../../luajit"

# Cleanup
cd ../../..
rm "${VENDOR_DIR}/luajit.tar.gz"

echo "LuaJIT vendored successfully."
