package cms

import (
	"errors"
	"sort"
	"sync"
	"time"
)

var ErrNotFound = errors.New("content not found")

// Post is the deliberately small Phase 0 content contract. The WordPress
// compatibility layer will map pinned wp_posts snapshots into this shape.
type Post struct {
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Store interface {
	GetPost(slug string) (Post, error)
	SavePost(post Post) error
	ListPosts() []Post
}

type MemoryStore struct {
	mu    sync.RWMutex
	posts map[string]Post
}

func NewMemoryStore(posts ...Post) *MemoryStore {
	store := &MemoryStore{posts: make(map[string]Post, len(posts))}
	for _, post := range posts {
		store.posts[post.Slug] = post
	}
	return store
}

func (s *MemoryStore) GetPost(slug string) (Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	post, ok := s.posts[slug]
	if !ok {
		return Post{}, ErrNotFound
	}
	return post, nil
}

func (s *MemoryStore) SavePost(post Post) error {
	if post.Slug == "" || post.Title == "" {
		return errors.New("slug and title are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	post.UpdatedAt = time.Now().UTC()
	s.posts[post.Slug] = post
	return nil
}

func (s *MemoryStore) ListPosts() []Post {
	s.mu.RLock()
	defer s.mu.RUnlock()
	posts := make([]Post, 0, len(s.posts))
	for _, post := range s.posts {
		posts = append(posts, post)
	}
	sort.Slice(posts, func(i, j int) bool { return posts[i].Slug < posts[j].Slug })
	return posts
}
