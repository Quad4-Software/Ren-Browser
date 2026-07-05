package assets_test

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"renbrowser/internal/assets"
)

func TestLoaderDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.html")
	if err := os.WriteFile(path, []byte("<html>ok</html>"), 0o644); err != nil {
		t.Fatal(err)
	}

	loader, err := assets.New(assets.Config{Dir: dir})
	if err != nil {
		t.Fatalf("New dir loader: %v", err)
	}
	defer loader.Close()

	if loader.Kind() != assets.SourceDir {
		t.Fatalf("kind = %q; want dir", loader.Kind())
	}

	body, err := loader.ReadFile("index.html")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(body) != "<html>ok</html>" {
		t.Fatalf("body = %q", string(body))
	}
}

func TestLoaderZip(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "bundle.zip")

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("index.html")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("<html>zip</html>")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	loader, err := assets.New(assets.Config{ZipPath: zipPath})
	if err != nil {
		t.Fatalf("New zip loader: %v", err)
	}
	defer loader.Close()

	if loader.Kind() != assets.SourceZip {
		t.Fatalf("kind = %q; want zip", loader.Kind())
	}

	body, err := loader.ReadFile("index.html")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(body) != "<html>zip</html>" {
		t.Fatalf("body = %q", string(body))
	}
}

func TestLoaderRequiresSource(t *testing.T) {
	_, err := assets.New(assets.Config{})
	if err == nil {
		t.Fatal("expected error for empty config")
	}
}
