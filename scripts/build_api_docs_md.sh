#!/bin/bash
set -e

DOCS_DIR="docs/docusaurus/docs/api"
MODULE=$(go list -m)

mkdir -p "$DOCS_DIR"

if ! command -v gomarkdoc &>/dev/null; then
	echo "gomarkdoc not found. Installing..."
	go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
fi

echo "Generating Go API docs with gomarkdoc..."

# Generate docs for all packages recursively
# We use --output to specify the output format/location
# gomarkdoc doesn't support -r recursive easily with flat output to directories,
# but we can list packages and loop.

# Create _category_.json for the API section
# Ensure slug matches docusaurus.config.ts configuration (/docs/go-api)
cat <<EOF >"$DOCS_DIR/_category_.json"
{
  "label": "Go API",
  "position": 5,
  "link": {
    "type": "generated-index",
    "slug": "go-api",
    "description": "Go API Reference for cogeni packages."
  }
}
EOF

# Function to generate doc for a package
generate_pkg_doc() {
	local pkg=$1
	local relpkg=${pkg#"$MODULE"/}
	if [ "$pkg" == "$MODULE" ]; then relpkg="root"; fi

	# Create directory structure for the package documentation
	local outpath="$DOCS_DIR/$relpkg.md"
	local outdir
	outdir=$(dirname "$outpath")
	mkdir -p "$outdir"
	echo "Generating $pkg -> $outpath"
	gomarkdoc --output "$outpath" "$pkg"
}

# List all packages and generate docs
go list ./src/... | while read -r pkg; do
	generate_pkg_doc "$pkg"
done

echo "Go API docs generated successfully."
