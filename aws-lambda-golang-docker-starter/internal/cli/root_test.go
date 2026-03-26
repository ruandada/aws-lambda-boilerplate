package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEventFile(t *testing.T) {
	t.Run("reads valid file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "event.json")
		content := []byte(`{"foo":"bar"}`)
		if err := os.WriteFile(path, content, 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		got, err := loadEventFile(path)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if string(got) != string(content) {
			t.Fatalf("expected %q, got %q", string(content), string(got))
		}
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		_, err := loadEventFile(filepath.Join(t.TempDir(), "missing.json"))
		if err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
	})

	t.Run("returns error for empty file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "empty.json")
		if err := os.WriteFile(path, []byte{}, 0o600); err != nil {
			t.Fatalf("failed to write empty file: %v", err)
		}

		_, err := loadEventFile(path)
		if err == nil {
			t.Fatal("expected error for empty file, got nil")
		}
	})
}
