package cms

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

const maxSnapshotBytes = 16 << 20

// FileStore is a local single-writer snapshot store for Linux, macOS and FreeBSD.
// It is not a WordPress database format or a multi-host storage service.
type FileStore struct {
	mu     sync.RWMutex
	memory *MemoryStore
	path   string
	lock   *os.File
}

type snapshot struct {
	Schema int    `json:"schema_version"`
	Posts  []Post `json:"posts"`
}

// OpenFileStore seeds only a missing file, rejects corrupt/unknown data, and
// holds an OS lock until Close or process exit (including abrupt termination).
func OpenFileStore(path string, seeds ...Post) (*FileStore, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(filepath.Dir(absolute), 0700); err != nil {
		return nil, err
	}
	lock, err := os.OpenFile(absolute+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		lock.Close()
		return nil, fmt.Errorf("content store is already locked or locking unavailable: %w", err)
	}
	store := &FileStore{path: absolute, lock: lock}
	success := false
	defer func() {
		if !success {
			store.Close()
		}
	}()
	file, err := os.Open(absolute)
	if errors.Is(err, os.ErrNotExist) {
		for _, p := range seeds {
			if err := validatePost(p); err != nil {
				return nil, err
			}
		}
		store.memory = NewMemoryStore(seeds...)
		if err := store.persist(store.memory.ListPosts()); err != nil {
			return nil, err
		}
	} else {
		if err != nil {
			return nil, err
		}
		defer file.Close()
		data, err := io.ReadAll(io.LimitReader(file, maxSnapshotBytes+1))
		if err != nil {
			return nil, err
		}
		if len(data) > maxSnapshotBytes {
			return nil, errors.New("content snapshot exceeds 16 MiB")
		}
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		var saved snapshot
		if err := decoder.Decode(&saved); err != nil {
			return nil, fmt.Errorf("invalid content snapshot: %w", err)
		}
		var trailing any
		if err := decoder.Decode(&trailing); err != io.EOF {
			return nil, errors.New("trailing content snapshot data")
		}
		if saved.Schema != 1 || saved.Posts == nil {
			return nil, errors.New("unsupported or invalid content snapshot schema")
		}
		seen := map[string]bool{}
		for _, p := range saved.Posts {
			if err := validatePost(p); err != nil {
				return nil, err
			}
			if seen[p.Slug] {
				return nil, fmt.Errorf("duplicate post slug %q", p.Slug)
			}
			seen[p.Slug] = true
		}
		store.memory = NewMemoryStore(saved.Posts...)
	}
	success = true
	return store, nil
}

func validatePost(post Post) error {
	if !slugPattern.MatchString(post.Slug) || strings.TrimSpace(post.Title) == "" {
		return errors.New("valid slug and non-empty title are required")
	}
	return nil
}

func (s *FileStore) GetPost(slug string) (Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.lock == nil {
		return Post{}, errors.New("store closed")
	}
	return s.memory.GetPost(slug)
}

func (s *FileStore) ListPosts() []Post {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.lock == nil {
		return nil
	}
	return s.memory.ListPosts()
}

func (s *FileStore) SavePost(post Post) error {
	if err := validatePost(post); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lock == nil {
		return errors.New("store closed")
	}
	post.UpdatedAt = time.Now().UTC()
	posts := s.memory.ListPosts()
	replaced := false
	for i := range posts {
		if posts[i].Slug == post.Slug {
			posts[i] = post
			replaced = true
			break
		}
	}
	if !replaced {
		posts = append(posts, post)
	}
	if err := s.persist(posts); err != nil {
		return err
	}
	s.memory = NewMemoryStore(posts...)
	return nil
}

func (s *FileStore) persist(posts []Post) error {
	data, err := json.MarshalIndent(snapshot{Schema: 1, Posts: posts}, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if len(data) > maxSnapshotBytes {
		return errors.New("content snapshot exceeds 16 MiB")
	}
	temporary, err := os.CreateTemp(filepath.Dir(s.path), ".content-")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	if _, err = temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err = temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err = temporary.Close(); err != nil {
		return err
	}
	// Same-directory rename publishes a complete file. A pre-rename failure
	// leaves both the previous snapshot and the in-memory view unchanged.
	return os.Rename(temporary.Name(), s.path)
}

func (s *FileStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lock == nil {
		return nil
	}
	err := s.lock.Close() // The kernel releases flock; never unlink the lock inode.
	s.lock = nil
	return err
}
