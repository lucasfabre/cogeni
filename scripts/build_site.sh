#!/bin/bash
set -e

echo "Building Docusaurus site..."

cd docs/docusaurus

# Install dependencies if node_modules missing
if [ ! -d "node_modules" ]; then
	echo "Installing dependencies..."
	npm ci
else
	# Check if package-lock.json changed? easier to just run npm ci or install
	# npm install is safer if lockfile is out of sync
	npm install
fi

# Build to dist/docs
# Note: Docusaurus cleans the output directory by default.
# So this script should be run BEFORE generating other docs into dist/docs.
echo "Running build..."
npm run build -- --out-dir ../../dist/docs

echo "Docusaurus build complete."
