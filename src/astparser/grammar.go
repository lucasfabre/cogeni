// Package astparser provides tools for parsing source code into a detailed
// and serializable Abstract Syntax Tree (AST) using Tree-sitter.
package astparser

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"

	"github.com/ebitengine/purego"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

// GrammarOptions provides configuration for building a Tree-sitter grammar from source.
type GrammarOptions struct {
	// BuildCmd is an optional shell command to compile the grammar (e.g., "make").
	BuildCmd string
	// Artifact is the name of the resulting shared library (e.g., "parser.so").
	Artifact string
	// Branch is the Git branch to use for downloading the source.
	Branch string
	// URL is the Git repository URL for the grammar.
	URL string
}

type httpStatusError struct {
	statusCode int
	url        string
}

func (e *httpStatusError) Error() string {
	return fmt.Sprintf("http error %d for %s", e.statusCode, e.url)
}

// GrammarManager handles the lifecycle of Tree-sitter grammar shared libraries.
// It manages downloading source code, compiling it into shared objects, and
// loading those objects at runtime using dynamic linking.
type GrammarManager struct {
	// baseDir is where compiled shared libraries are stored.
	baseDir string
	// grammars stores custom configuration for specific language names.
	grammars map[string]GrammarOptions
}

// NewGrammarManager initializes a manager with the specified storage directory.
func NewGrammarManager(baseDir string) (*GrammarManager, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create grammar directory: %w", err)
	}
	return &GrammarManager{
		baseDir:  baseDir,
		grammars: make(map[string]GrammarOptions),
	}, nil
}

// RegisterGrammar configures a custom source and build process for a language.
func (m *GrammarManager) RegisterGrammar(name, url string, opts GrammarOptions) {
	opts.URL = url
	m.grammars[name] = opts
}

// GetLanguage returns a Tree-sitter Language instance for the given name.
// It will automatically fetch and compile the grammar if it is not found in the cache.
func (m *GrammarManager) GetLanguage(name string) (*sitter.Language, error) {
	opts, ok := m.grammars[name]
	if !ok {
		// Default to the official Tree-sitter organization on GitHub
		opts = GrammarOptions{
			URL:    fmt.Sprintf("https://github.com/tree-sitter/tree-sitter-%s", name),
			Branch: "master",
		}
	}

	libPath := m.getLibraryPath(name, opts)

	if _, err := os.Stat(libPath); os.IsNotExist(err) {
		if err := m.fetchAndBuild(name, libPath, opts); err != nil {
			var statusErr *httpStatusError
			// Fallback only when the default branch archive is genuinely missing.
			if !ok && opts.Branch == "master" && errors.As(err, &statusErr) && statusErr.statusCode == http.StatusNotFound {
				opts.Branch = "main"
				libPath = m.getLibraryPath(name, opts)
				if _, errStat := os.Stat(libPath); os.IsNotExist(errStat) {
					if errFetch := m.fetchAndBuild(name, libPath, opts); errFetch != nil {
						return nil, fmt.Errorf("failed to fetch/build grammar '%s' (tried master and main): master=%v; main=%w", name, err, errFetch)
					}
				}
			} else {
				return nil, err
			}
		}
	}

	return m.loadSharedLibrary(libPath, name)
}

// getLibraryPath calculates the unique cache path for a grammar based on its URL and branch.
func (m *GrammarManager) getLibraryPath(name string, opts GrammarOptions) string {
	ext := ".so"
	prefix := "lib"
	switch runtime.GOOS {
	case "darwin":
		ext = ".dylib"
	case "windows":
		ext = ".dll"
		prefix = ""
	}

	hash := sha256.Sum256([]byte(opts.URL + opts.Branch))
	return filepath.Join(m.baseDir, fmt.Sprintf("%stree-sitter-%s-%x%s", prefix, name, hash[:8], ext))
}

// fetchAndBuild performs the end-to-end process of downloading and compiling a grammar.
func (m *GrammarManager) fetchAndBuild(name, libPath string, opts GrammarOptions) error {
	fmt.Fprintf(os.Stderr, "Building grammar '%s' from %s#%s...\n", name, opts.URL, opts.Branch)

	tmpDir, err := os.MkdirTemp("", "cogeni-")
	if err != nil {
		return fmt.Errorf("failed to create temp build directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if err := m.downloadAndExtract(opts, tmpDir); err != nil {
		return err
	}

	if opts.BuildCmd != "" {
		return m.runCustomBuild(tmpDir, libPath, opts)
	}

	parserC, err := findParserFile(tmpDir, name)
	if err != nil {
		return err
	}

	return m.compileGrammar(parserC, libPath)
}

// downloadAndExtract fetches the source code (usually as a tarball from GitHub) and extracts it.
func (m *GrammarManager) downloadAndExtract(opts GrammarOptions, dest string) error {
	url := opts.URL
	// Rewrite GitHub URLs to use the tarball download endpoint
	if strings.HasPrefix(url, "https://github.com/") && !strings.HasSuffix(url, ".tar.gz") {
		branch := opts.Branch
		if branch == "" {
			branch = "master"
		}
		url = fmt.Sprintf("%s/archive/refs/heads/%s.tar.gz", strings.TrimSuffix(url, "/"), branch)
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http request failed for %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return &httpStatusError{statusCode: resp.StatusCode, url: url}
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read error: %w", err)
		}

		path := filepath.Join(dest, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				_ = f.Close()
				return err
			}
			if err := f.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}

// runCustomBuild executes a user-provided build command to generate the grammar artifact.
func (m *GrammarManager) runCustomBuild(tmpDir, libPath string, opts GrammarOptions) error {
	// Handle cases where the tarball contains a single root directory
	entries, _ := os.ReadDir(tmpDir)
	buildDir := tmpDir
	if len(entries) == 1 && entries[0].IsDir() {
		buildDir = filepath.Join(tmpDir, entries[0].Name())
	}

	cmd := exec.Command("sh", "-c", opts.BuildCmd)
	cmd.Dir = buildDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("custom build failed: %s: %w", string(out), err)
	}

	artifact := opts.Artifact
	if artifact == "" {
		// Auto-discover the shared library if not specified
		_ = filepath.Walk(buildDir, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() && (strings.HasSuffix(path, ".so") || strings.HasSuffix(path, ".dylib") || strings.HasSuffix(path, ".dll")) {
				artifact = path
				return filepath.SkipAll
			}
			return nil
		})
	} else {
		artifact = filepath.Join(buildDir, artifact)
	}

	if _, err := os.Stat(artifact); err != nil {
		return fmt.Errorf("expected artifact not found: %s", artifact)
	}

	data, err := os.ReadFile(artifact)
	if err != nil {
		return err
	}
	return os.WriteFile(libPath, data, 0755)
}

// compileGrammar uses a system C compiler to build parser.c into a shared library.
func (m *GrammarManager) compileGrammar(parserC, libPath string) error {
	srcDir := filepath.Dir(parserC)
	compiler, err := exec.LookPath("gcc")
	if err != nil {
		compiler, err = exec.LookPath("clang")
	}
	if err != nil {
		return fmt.Errorf("no C compiler (gcc or clang) found in PATH")
	}

	args := []string{"-shared", "-I", srcDir, "-o", libPath, parserC}
	if runtime.GOOS != "windows" {
		args = append(args, "-fPIC")
	}
	// Many grammars also require scanner.c or scanner.cc
	if _, err := os.Stat(filepath.Join(srcDir, "scanner.c")); err == nil {
		args = append(args, filepath.Join(srcDir, "scanner.c"))
	} else if _, err := os.Stat(filepath.Join(srcDir, "scanner.cc")); err == nil {
		args = append(args, filepath.Join(srcDir, "scanner.cc"), "-lstdc++")
	}

	cmd := exec.Command(compiler, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("c compilation failed: %s: %w", string(out), err)
	}
	return nil
}

// loadSharedLibrary uses purego to dynamically load the compiled grammar and retrieve the Language symbol.
func (m *GrammarManager) loadSharedLibrary(path, lang string) (*sitter.Language, error) {
	lib, err := openLibrary(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open shared library %s: %w", path, err)
	}

	var langFunc func() unsafe.Pointer
	// Tree-sitter entry points follow the pattern tree_sitter_<lang>
	symbol := fmt.Sprintf("tree_sitter_%s", strings.ReplaceAll(lang, "-", "_"))
	purego.RegisterLibFunc(&langFunc, lib, symbol)

	return sitter.NewLanguage(langFunc()), nil
}

// findParserFile heuristically locates parser.c within a source directory.
func findParserFile(root, lang string) (string, error) {
	var best, first string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if info.Name() == "parser.c" && strings.Contains(path, "/src/") {
			if first == "" {
				first = path
			}

			// A better match is one where the language name is a directory component just before 'src'
			// e.g., .../typescript/src/parser.c
			components := strings.Split(filepath.ToSlash(path), "/")
			for i := len(components) - 1; i >= 1; i-- {
				if components[i] == "src" && i > 0 && components[i-1] == lang {
					best = path
					return filepath.SkipAll
				}
			}

			// Fallback: if 'best' is not set yet, check if lang is anywhere in the path
			if best == "" && strings.Contains(path, lang) {
				best = path
			}
		}
		return nil
	})

	if best != "" {
		return best, nil
	}
	if first != "" {
		return first, nil
	}
	return "", fmt.Errorf("could not find parser.c in %s", root)
}
