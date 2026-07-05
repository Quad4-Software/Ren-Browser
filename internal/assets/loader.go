package assets

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type SourceKind string

const (
	SourceEmbedded SourceKind = "embedded"
	SourceDir      SourceKind = "dir"
	SourceZip      SourceKind = "zip"
)

type Config struct {
	Embedded fs.FS
	Dir      string
	ZipPath  string
}

type Loader struct {
	kind SourceKind
	root fs.FS
	zip  *zip.ReadCloser
}

func New(cfg Config) (*Loader, error) {
	switch {
	case strings.TrimSpace(cfg.Dir) != "":
		dir := filepath.Clean(cfg.Dir)
		info, err := os.Stat(dir)
		if err != nil {
			return nil, fmt.Errorf("assets dir: %w", err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("assets dir %q is not a directory", dir)
		}
		return &Loader{kind: SourceDir, root: os.DirFS(dir)}, nil
	case strings.TrimSpace(cfg.ZipPath) != "":
		zr, err := zip.OpenReader(filepath.Clean(cfg.ZipPath))
		if err != nil {
			return nil, fmt.Errorf("assets zip: %w", err)
		}
		return &Loader{kind: SourceZip, root: zr, zip: zr}, nil
	case cfg.Embedded != nil:
		return &Loader{kind: SourceEmbedded, root: cfg.Embedded}, nil
	default:
		return nil, fmt.Errorf("no asset source configured")
	}
}

func SubEmbedded(embed fs.FS, dir string) (fs.FS, error) {
	sub, err := fs.Sub(embed, filepath.ToSlash(dir))
	if err != nil {
		return nil, fmt.Errorf("embed sub %q: %w", dir, err)
	}
	return sub, nil
}

func (l *Loader) Kind() SourceKind {
	return l.kind
}

func (l *Loader) FS() fs.FS {
	return l.root
}

func (l *Loader) Close() error {
	if l.zip != nil {
		return l.zip.Close()
	}
	return nil
}

func (l *Loader) Handler() http.Handler {
	switch l.kind {
	case SourceEmbedded:
		return application.BundledAssetFileServer(l.root)
	default:
		return application.AssetFileServerFS(l.root)
	}
}

func (l *Loader) ReadFile(name string) ([]byte, error) {
	name = strings.TrimPrefix(filepath.ToSlash(name), "/")
	f, err := l.root.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}
