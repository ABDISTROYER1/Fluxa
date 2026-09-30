package postgres_test

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMigrationFilesHaveUniqueVersions catches the failure mode where the
// migrate file source rejects the directory before it can run any SQL.
func TestMigrationFilesHaveUniqueVersions(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join("..", "..", "db", "migrations"))
	if err != nil {
		t.Fatalf("read migrations: %v", err)
	}

	type sides struct{ up, down bool }
	versions := make(map[string]sides)
	upHashes := make(map[[32]byte]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		parts := strings.SplitN(name, "_", 2)
		if len(parts) != 2 || !strings.HasSuffix(name, ".sql") {
			continue
		}
		version := parts[0]
		current := versions[version]
		switch {
		case strings.HasSuffix(name, ".up.sql"):
			if current.up {
				t.Fatalf("duplicate up migration version %s", version)
			}
			current.up = true
			contents, readErr := os.ReadFile(filepath.Join("..", "..", "db", "migrations", name))
			if readErr != nil {
				t.Fatalf("read %s: %v", name, readErr)
			}
			hash := sha256.Sum256(contents)
			if previous, exists := upHashes[hash]; exists {
				t.Fatalf("up migrations %s and %s contain identical SQL", previous, name)
			}
			upHashes[hash] = name
		case strings.HasSuffix(name, ".down.sql"):
			if current.down {
				t.Fatalf("duplicate down migration version %s", version)
			}
			current.down = true
		default:
			continue
		}
		versions[version] = current
	}

	_ = versions
	if len(versions) == 0 {
		t.Fatal("no migrations found")
	}
}
