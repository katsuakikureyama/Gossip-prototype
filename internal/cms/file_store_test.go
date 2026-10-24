package cms

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

func TestCreateFileStoreNeverReplacesExistingPath(t *testing.T) {
	for _, kind := range []string{"snapshot", "corrupt", "symlink", "dangling symlink"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "content.json")
			if kind == "snapshot" {
				store, err := CreateFileStore(path, Post{Slug: "original", Title: "Original"})
				if err != nil {
					t.Fatal(err)
				}
				store.Close()
			} else if kind == "corrupt" {
				if err := os.WriteFile(path, []byte("broken"), 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				target := filepath.Join(dir, "target")
				if kind == "symlink" {
					if err := os.WriteFile(target, []byte("retain target"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.Symlink(target, path); err != nil {
					t.Fatal(err)
				}
			}
			before, beforeErr := os.ReadFile(path)
			info, _ := os.Lstat(path)
			if store, err := CreateFileStore(path, Post{Slug: "replacement", Title: "Must not replace"}); !errors.Is(err, os.ErrExist) {
				if store != nil {
					store.Close()
				}
				t.Fatalf("expected exists error, got %v", err)
			}
			after, afterErr := os.ReadFile(path)
			afterInfo, _ := os.Lstat(path)
			if string(before) != string(after) || (beforeErr == nil) != (afterErr == nil) || !os.SameFile(info, afterInfo) {
				t.Fatal("existing destination changed")
			}
		})
	}
}

func TestCreateFileStoreRejectsDuplicateSeeds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "content.json")
	if store, err := CreateFileStore(path, Post{Slug: "same", Title: "A"}, Post{Slug: "same", Title: "B"}); err == nil {
		store.Close()
		t.Fatal("duplicate seed silently lost")
	}
	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("snapshot created: %v", err)
	}
}

func TestFileStoreRestartAndLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "content.json")
	seed := Post{Slug: "welcome", Title: "Initial"}
	first, err := OpenFileStore(path, seed)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	if other, err := OpenFileStore(path); err == nil {
		other.Close()
		t.Fatal("second writer allowed")
	}
	seed.Title, seed.Content = "保存", "日本語\n<script>literal</script>"
	if err := first.SavePost(seed); err != nil {
		t.Fatal(err)
	}
	saved, _ := first.GetPost(seed.Slug)
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := OpenFileStore(path, Post{Slug: "welcome", Title: "Must not overwrite"})
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	loaded, err := second.GetPost(seed.Slug)
	if err != nil || !reflect.DeepEqual(saved, loaded) {
		t.Fatalf("restart mismatch: %+v %+v %v", saved, loaded, err)
	}
}

func TestFileStoreFailedWritePreservesSnapshotAndMemory(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "data")
	path := filepath.Join(dir, "content.json")
	store, err := OpenFileStore(path, Post{Slug: "welcome", Title: "Original"})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	before, _ := os.ReadFile(path)
	// A real filesystem failure: move our test directory so temp creation fails.
	if err := os.Rename(dir, filepath.Join(parent, "retained")); err != nil {
		t.Fatal(err)
	}
	if err := store.SavePost(Post{Slug: "welcome", Title: "Must not publish"}); err == nil {
		t.Fatal("expected write failure")
	}
	after, err := os.ReadFile(filepath.Join(parent, "retained", "content.json"))
	if err != nil || string(before) != string(after) {
		t.Fatalf("snapshot changed: %v", err)
	}
	post, _ := store.GetPost("welcome")
	if post.Title != "Original" {
		t.Fatalf("memory changed: %+v", post)
	}
}

func TestFileStoreRejectsCorruptionWithoutReseeding(t *testing.T) {
	for _, data := range []string{
		`{`, `{"schema_version":2,"posts":[]}`, `{"schema_version":1,"posts":null}`,
		`{"schema_version":1,"posts":[]} {}`, `{"schema_version":1,"posts":[{"slug":"bad/path","title":"x"}]}`,
		`{"schema_version":1,"posts":[{"slug":"same","title":"x"},{"slug":"same","title":"y"}]}`,
	} {
		t.Run(data, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "content.json")
			if err := os.WriteFile(path, []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			store, err := OpenFileStore(path, Post{Slug: "welcome", Title: "seed"})
			if err == nil {
				store.Close()
				t.Fatal("invalid snapshot accepted")
			}
			actual, _ := os.ReadFile(path)
			if string(actual) != data {
				t.Fatal("invalid snapshot overwritten")
			}
		})
	}
}

func TestFileStoreConcurrentSavesSurviveReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "content.json")
	store, err := OpenFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	var wg sync.WaitGroup
	for _, slug := range []string{"one", "two", "three"} {
		wg.Add(1)
		go func(slug string) {
			defer wg.Done()
			if err := store.SavePost(Post{Slug: slug, Title: slug}); err != nil {
				t.Error(err)
			}
		}(slug)
	}
	wg.Wait()
	store.Close()
	reopened, err := OpenFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if len(reopened.ListPosts()) != 3 {
		t.Fatal("concurrent save lost a post")
	}
}
