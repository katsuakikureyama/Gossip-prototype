package cms

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/katsuakikureyama/Gossip/pkg/pluginsdk"
	"github.com/katsuakikureyama/Gossip/plugins/sample"
)

func testHandler(t *testing.T, editor bool) (http.Handler, *MemoryStore) {
	t.Helper()
	registry := pluginsdk.NewRegistry()
	if err := registry.Register(sample.Plugin{}); err != nil {
		t.Fatal(err)
	}
	store := NewMemoryStore(Post{Slug: "welcome", Title: "Welcome", Content: "Hello"})
	handler, err := NewHandler(Config{Store: store, Plugins: registry, DevEditor: editor, DevEditorKey: "test-key"})
	if err != nil {
		t.Fatal(err)
	}
	return handler, store
}

func TestViewRunsSamplePluginAndEscapesContent(t *testing.T) {
	handler, store := testHandler(t, false)
	if err := store.SavePost(Post{Slug: "welcome", Title: "Welcome", Content: "<script>alert(1)</script>"}); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/posts/welcome", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	body := response.Body.String()
	if strings.Contains(body, "<script>") {
		t.Fatalf("content was not escaped: %s", body)
	}
	if !strings.Contains(body, "Rendered by the native Go sample plugin") {
		t.Fatalf("sample plugin output missing: %s", body)
	}
}

func TestDevEditorRequiresFeatureAndKeyThenSaves(t *testing.T) {
	disabled, _ := testHandler(t, false)
	request := httptest.NewRequest(http.MethodPost, "/dev/posts/welcome/edit", nil)
	response := httptest.NewRecorder()
	disabled.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("disabled status = %d", response.Code)
	}

	enabled, store := testHandler(t, true)
	form := url.Values{"title": {"Updated"}, "content": {"Edited fixture"}}
	request = httptest.NewRequest(http.MethodPost, "/dev/posts/welcome/edit", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	enabled.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("missing key status = %d", response.Code)
	}

	form.Set("dev_key", "test-key")
	request = httptest.NewRequest(http.MethodPost, "/dev/posts/welcome/edit", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response = httptest.NewRecorder()
	enabled.ServeHTTP(response, request)
	if response.Code != http.StatusSeeOther {
		t.Fatalf("save status = %d, body = %s", response.Code, response.Body.String())
	}
	post, err := store.GetPost("welcome")
	if err != nil {
		t.Fatal(err)
	}
	if post.Title != "Updated" || post.Content != "Edited fixture" {
		t.Fatalf("saved post = %#v", post)
	}
}

func TestDevEditorGETCarriesPerRunKeyIntoForm(t *testing.T) {
	handler, _ := testHandler(t, true)
	request := httptest.NewRequest(http.MethodGet, "/dev/posts/welcome/edit?key=test-key", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), `name="dev_key" value="test-key"`) {
		t.Fatalf("editor key missing from form: %s", response.Body.String())
	}
}

func TestEditorConfigRequiresKey(t *testing.T) {
	_, err := NewHandler(Config{Store: NewMemoryStore(), Plugins: pluginsdk.NewRegistry(), DevEditor: true})
	if err == nil {
		t.Fatal("expected missing editor key error")
	}
}
