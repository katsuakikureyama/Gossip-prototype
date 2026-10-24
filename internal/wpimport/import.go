// Package wpimport maps a narrow, owned WP-CLI post export to local CMS posts.
// It does not migrate a WordPress database or render WordPress markup.
package wpimport

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/katsuakikureyama/Gossip/internal/cms"
)

const MaxInputBytes = 16 << 20

type sourcePost struct {
	ID       json.Number `json:"ID"`
	Type     *string     `json:"post_type"`
	Status   *string     `json:"post_status"`
	Password *string     `json:"post_password"`
	Slug     *string     `json:"post_name"`
	Title    *string     `json:"post_title"`
	Content  *string     `json:"post_content"`
	Modified *string     `json:"post_modified_gmt"`
}

type Entry struct {
	SourceID int64  `json:"source_id"`
	Slug     string `json:"slug"`
	Title    string `json:"title"`
}

type Plan struct {
	Status       string     `json:"status"`
	SourceSHA256 string     `json:"source_sha256"`
	Entries      []Entry    `json:"posts"`
	Warnings     []string   `json:"warnings"`
	Posts        []cms.Post `json:"-"`
}

var validSlug = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Prepare reads bounded input and rejects the entire batch if any row is
// unsupported. In particular, private/draft/passworded content must never
// become an anonymous public post by silently dropping visibility metadata.
func Prepare(r io.Reader) (Plan, error) {
	var plan Plan
	data, err := io.ReadAll(io.LimitReader(r, MaxInputBytes+1))
	if err != nil {
		return plan, err
	}
	if len(data) > MaxInputBytes || !utf8.Valid(data) {
		return plan, errors.New("input exceeds 16 MiB or is not UTF-8")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var rows []sourcePost
	if err := decoder.Decode(&rows); err != nil {
		return plan, fmt.Errorf("invalid WP-CLI export: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF || len(rows) == 0 {
		return plan, errors.New("expected one non-empty JSON post array")
	}
	ids, slugs := map[int64]bool{}, map[string]bool{}
	for i, row := range rows {
		id, err := row.ID.Int64()
		if err != nil || id <= 0 || row.Type == nil || row.Status == nil || row.Password == nil || row.Slug == nil || row.Title == nil || row.Content == nil || row.Modified == nil {
			return Plan{}, fmt.Errorf("row %d: all selected fields and a positive integer ID are required", i+1)
		}
		if *row.Type != "post" || *row.Status != "publish" || *row.Password != "" {
			return Plan{}, fmt.Errorf("row %d: only published, non-password-protected posts are accepted", i+1)
		}
		if !validSlug.MatchString(*row.Slug) || strings.TrimSpace(*row.Title) == "" || ids[id] || slugs[*row.Slug] {
			return Plan{}, fmt.Errorf("row %d: unsupported slug, blank title or duplicate identity", i+1)
		}
		modified, err := time.Parse("2006-01-02 15:04:05", *row.Modified)
		if err != nil || modified.Year() < 1 {
			return Plan{}, fmt.Errorf("row %d: invalid UTC modification time", i+1)
		}
		ids[id], slugs[*row.Slug] = true, true
		plan.Entries = append(plan.Entries, Entry{SourceID: id, Slug: *row.Slug, Title: *row.Title})
		plan.Posts = append(plan.Posts, cms.Post{Slug: *row.Slug, Title: *row.Title, Content: *row.Content, UpdatedAt: modified})
	}
	hash := sha256.Sum256(data)
	plan.Status, plan.SourceSHA256 = "prepared", hex.EncodeToString(hash[:])
	plan.Warnings = []string{
		"Local fixture only: not a WordPress database migration.",
		"Only slug, raw title/content and UTC modification time persist; IDs are preview provenance only. Authors, terms, media, metadata, URLs and options are not imported.",
		"Raw HTML/block markup is stored as text and escaped on display; WordPress rendering parity is not claimed.",
	}
	return plan, nil
}
