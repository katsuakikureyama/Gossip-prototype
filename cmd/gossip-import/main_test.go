package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/katsuakikureyama/Gossip/internal/cms"
	"github.com/katsuakikureyama/Gossip/pkg/pluginsdk"
)

const testExport = `[
  {"ID":71,"post_type":"post","post_status":"publish","post_password":"","post_name":"first-post","post_title":"First post","post_content":"<p>first</p>","post_modified_gmt":"2026-01-02 03:04:05"},
  {"ID":72,"post_type":"post","post_status":"publish","post_password":"","post_name":"second-post","post_title":"Second post","post_content":"<p>second</p>","post_modified_gmt":"2026-01-03 04:05:06"}
]`

type importResult struct {
	Status       string `json:"status"`
	SourceSHA256 string `json:"source_sha256"`
}

func writeImportInput(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "wp-posts.json")
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func invokeImport(t *testing.T, arguments ...string) (importResult, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	err := run(arguments, &stdout, &stderr)
	if err != nil {
		return importResult{}, err
	}
	var result importResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("decode importer output: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}
	return result, nil
}

func TestPreviewDoesNotCreateSnapshot(t *testing.T) {
	input := writeImportInput(t, testExport)
	before, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	result, err := invokeImport(t, "-input", input)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "prepared" || result.SourceSHA256 == "" {
		t.Fatalf("preview result = %+v", result)
	}
	after, err := os.ReadFile(input)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("preview changed source: %v", err)
	}
	if entries, err := os.ReadDir(filepath.Dir(input)); err != nil || len(entries) != 1 {
		t.Fatalf("preview created filesystem output: entries=%v err=%v", entries, err)
	}
}

func TestCommitCreatesExactReviewedSnapshot(t *testing.T) {
	input := writeImportInput(t, testExport)
	preview, err := invokeImport(t, "-input", input)
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "content.json")
	committed, err := invokeImport(t, "-input", input, "-commit", "-out", output, "-expect-sha256", preview.SourceSHA256)
	if err != nil {
		t.Fatal(err)
	}
	if committed.Status != "saved" || committed.SourceSHA256 != preview.SourceSHA256 {
		t.Fatalf("commit result = %+v, preview = %+v", committed, preview)
	}
	store, err := cms.OpenFileStore(output)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	expected := []cms.Post{
		{Slug: "first-post", Title: "First post", Content: "<p>first</p>", UpdatedAt: mustUTCTime(t, "2026-01-02 03:04:05")},
		{Slug: "second-post", Title: "Second post", Content: "<p>second</p>", UpdatedAt: mustUTCTime(t, "2026-01-03 04:05:06")},
	}
	if actual := store.ListPosts(); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("snapshot posts = %#v, want %#v", actual, expected)
	}
}

func TestCommitNeverOverwritesExistingTargetOrRepeats(t *testing.T) {
	input := writeImportInput(t, testExport)
	preview, err := invokeImport(t, "-input", input)
	if err != nil {
		t.Fatal(err)
	}
	arguments := []string{"-input", input, "-commit", "-expect-sha256", preview.SourceSHA256}

	existing := filepath.Join(t.TempDir(), "already-there.json")
	if err := os.WriteFile(existing, []byte("do not replace"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := invokeImport(t, append(arguments, "-out", existing)...); err == nil {
		t.Fatal("existing target was accepted")
	}
	if actual, err := os.ReadFile(existing); err != nil || string(actual) != "do not replace" {
		t.Fatalf("existing target changed: %q, %v", actual, err)
	}

	output := filepath.Join(t.TempDir(), "created-once.json")
	if _, err := invokeImport(t, append(arguments, "-out", output)...); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := invokeImport(t, append(arguments, "-out", output)...); err == nil {
		t.Fatal("repeat commit overwrote target")
	}
	after, err := os.ReadFile(output)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("repeat commit changed snapshot: %v", err)
	}
}

func TestCommitRejectsChangedSourceWithoutCreatingOutput(t *testing.T) {
	input := writeImportInput(t, testExport)
	preview, err := invokeImport(t, "-input", input)
	if err != nil {
		t.Fatal(err)
	}
	changed := strings.Replace(testExport, "First post", "Changed after review", 1)
	if err := os.WriteFile(input, []byte(changed), 0600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "must-not-exist.json")
	_, err = invokeImport(t, "-input", input, "-commit", "-out", output, "-expect-sha256", preview.SourceSHA256)
	if err == nil || !strings.Contains(err.Error(), "source hash differs") {
		t.Fatalf("changed source error = %v", err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("changed source created output: %v", err)
	}
	if _, err := os.Stat(output + ".lock"); !os.IsNotExist(err) {
		t.Fatalf("changed source created lock: %v", err)
	}
}

func TestCommitRequiresOutputAndReviewedHash(t *testing.T) {
	input := writeImportInput(t, testExport)
	for _, arguments := range [][]string{
		{"-input", input, "-commit"},
		{"-input", input, "-commit", "-out", filepath.Join(t.TempDir(), "content.json")},
		{"-input", input, "-commit", "-expect-sha256", "reviewed"},
	} {
		if _, err := invokeImport(t, arguments...); err == nil || !strings.Contains(err.Error(), "commit requires") {
			t.Fatalf("arguments %q: error = %v", arguments, err)
		}
	}
}

func TestOwnedWordPressCLIExportRoundTrip(t *testing.T) {
	input := os.Getenv("GOSSIP_WP_POST_EXPORT")
	if input == "" {
		t.Skip("set GOSSIP_WP_POST_EXPORT to an owned WP-CLI JSON export")
	}
	source, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Slug     string `json:"post_name"`
		Title    string `json:"post_title"`
		Content  string `json:"post_content"`
		Modified string `json:"post_modified_gmt"`
	}
	if err := json.Unmarshal(source, &rows); err != nil || len(rows) != 1 {
		t.Fatalf("expected one owned source post: %v", err)
	}
	preview, err := invokeImport(t, "-input", input)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Status != "prepared" || preview.SourceSHA256 != fmt.Sprintf("%x", sha256.Sum256(source)) {
		t.Fatalf("preview does not identify the source bytes: %+v", preview)
	}
	output := filepath.Join(t.TempDir(), "wordpress.json")
	if _, err := invokeImport(t, "-input", input, "-commit", "-out", output, "-expect-sha256", preview.SourceSHA256); err != nil {
		t.Fatal(err)
	}
	store, err := cms.OpenFileStore(output)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if len(store.ListPosts()) != len(rows) {
		t.Fatal("import changed post count")
	}
	for _, row := range rows {
		actual, err := store.GetPost(row.Slug)
		if err != nil || actual.Title != row.Title || actual.Content != row.Content || actual.UpdatedAt.UTC().Format("2006-01-02 15:04:05") != row.Modified {
			t.Fatalf("snapshot differs from raw WP-CLI fields: %+v %v", actual, err)
		}
	}

	post, err := store.GetPost("gossip-oracle-gutenberg")
	if err != nil {
		t.Fatalf("owned export is missing gossip-oracle-gutenberg: %v", err)
	}
	fixture, err := os.ReadFile(filepath.Join("..", "..", "fixtures", "wordpress", "seed", "gutenberg-post.html"))
	if err != nil {
		t.Fatal(err)
	}
	if post.Content != string(fixture) {
		t.Fatalf("gutenberg content differs from checked-in fixture:\nactual: %q\nfixture: %q", post.Content, fixture)
	}
	if post.UpdatedAt.Location() != time.UTC {
		t.Fatalf("timestamp is not UTC: %#v", post.UpdatedAt)
	}

	handler, err := cms.NewHandler(cms.Config{Store: store, Plugins: pluginsdk.NewRegistry()})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/posts/gossip-oracle-gutenberg", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("anonymous GET status = %d: %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	if !strings.Contains(body, html.EscapeString(post.Content)) || strings.Contains(body, post.Content) {
		t.Fatalf("anonymous GET did not escape imported markup: %s", body)
	}
}

func mustUTCTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02 15:04:05", value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
