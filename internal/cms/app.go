package cms

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"github.com/katsuakikureyama/Gossip/internal/buildinfo"
	"github.com/katsuakikureyama/Gossip/pkg/pluginsdk"
)

const maxEditorBody = 1 << 20

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Config struct {
	Store        Store
	Plugins      *pluginsdk.Registry
	DevEditor    bool
	DevEditorKey string
}

// NewHandler builds the Phase 0 HTTP application. The editor is deliberately
// disabled unless both DevEditor and a non-empty per-run key are configured.
func NewHandler(config Config) (http.Handler, error) {
	if config.Store == nil {
		return nil, errors.New("store is required")
	}
	if config.Plugins == nil {
		return nil, errors.New("plugin registry is required")
	}
	if config.DevEditor && strings.TrimSpace(config.DevEditorKey) == "" {
		return nil, errors.New("dev editor requires a non-empty key")
	}
	app := &application{config: config}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", app.index)
	mux.HandleFunc("GET /healthz", app.health)
	mux.HandleFunc("GET /version", app.version)
	mux.HandleFunc("GET /api/v1/plugins", app.plugins)
	mux.HandleFunc("GET /posts/{slug}", app.viewPost)
	mux.HandleFunc("GET /dev/posts/{slug}/edit", app.editPost)
	mux.HandleFunc("POST /dev/posts/{slug}/edit", app.editPost)
	return securityHeaders(mux), nil
}

type application struct {
	config Config
}

func (a *application) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	posts := a.config.Store.ListPosts()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := indexTemplate.Execute(w, posts); err != nil {
		http.Error(w, "render failed", http.StatusInternalServerError)
	}
}

func (a *application) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *application) version(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, buildinfo.Current())
}

func (a *application) plugins(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"plugins": a.config.Plugins.Manifests()})
}

func (a *application) viewPost(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if !slugPattern.MatchString(slug) {
		http.NotFound(w, r)
		return
	}
	post, err := a.config.Store.GetPost(slug)
	if errors.Is(err, ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "content unavailable", http.StatusInternalServerError)
		return
	}
	event, err := a.config.Plugins.Apply(r.Context(), pluginsdk.HookContentRender, postEvent(post))
	if err != nil {
		http.Error(w, "plugin render failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := postTemplate.Execute(w, event.Fields); err != nil {
		http.Error(w, "render failed", http.StatusInternalServerError)
	}
}

func (a *application) editPost(w http.ResponseWriter, r *http.Request) {
	if !a.config.DevEditor {
		http.NotFound(w, r)
		return
	}
	slug := r.PathValue("slug")
	if !slugPattern.MatchString(slug) {
		http.NotFound(w, r)
		return
	}
	post, err := a.config.Store.GetPost(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if r.Method == http.MethodGet {
		if editorKey(r) != a.config.DevEditorKey {
			http.Error(w, "invalid development editor key", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := editorTemplate.Execute(w, editorView{Post: post, Key: a.config.DevEditorKey}); err != nil {
			http.Error(w, "render failed", http.StatusInternalServerError)
		}
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxEditorBody)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	if editorKey(r) != a.config.DevEditorKey {
		http.Error(w, "invalid development editor key", http.StatusUnauthorized)
		return
	}
	post.Title = strings.TrimSpace(r.FormValue("title"))
	post.Content = strings.TrimSpace(r.FormValue("content"))
	if post.Title == "" {
		http.Error(w, "title is required", http.StatusUnprocessableEntity)
		return
	}
	event, err := a.config.Plugins.Apply(r.Context(), pluginsdk.HookContentBeforeSave, postEvent(post))
	if err != nil {
		http.Error(w, "plugin rejected save", http.StatusUnprocessableEntity)
		return
	}
	post.Title, post.Content = event.Fields["title"], event.Fields["content"]
	if err := a.config.Store.SavePost(post); err != nil {
		http.Error(w, "save failed", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/posts/"+slug, http.StatusSeeOther)
}

func editorKey(r *http.Request) string {
	if key := r.Header.Get("X-Gossip-Dev-Key"); key != "" {
		return key
	}
	if r.Method == http.MethodGet {
		return r.URL.Query().Get("key")
	}
	return r.FormValue("dev_key")
}

func postEvent(post Post) pluginsdk.Event {
	return pluginsdk.Event{
		Resource: "post:" + post.Slug,
		Fields: map[string]string{
			"slug":    post.Slug,
			"title":   post.Title,
			"content": post.Content,
		},
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; base-uri 'none'; frame-ancestors 'none'")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

var indexTemplate = template.Must(template.New("index").Parse(`<!doctype html>
<html lang="en"><meta charset="utf-8"><title>Gossip</title>
<h1>Gossip development fixture</h1>
<ul>{{range .}}<li><a href="/posts/{{.Slug}}">{{.Title}}</a></li>{{else}}<li>No content</li>{{end}}</ul>
</html>`))

var postTemplate = template.Must(template.New("post").Parse(`<!doctype html>
<html lang="en"><meta charset="utf-8"><title>{{.title}}</title>
<article><h1>{{.title}}</h1><div class="content">{{.content}}</div></article>
{{with .plugin_footer}}<footer>{{.}}</footer>{{end}}
</html>`))

type editorView struct {
	Post Post
	Key  string
}

var editorTemplate = template.Must(template.New("editor").Parse(`<!doctype html>
<html lang="en"><meta charset="utf-8"><title>Edit {{.Post.Title}}</title>
<form method="post"><input type="hidden" name="dev_key" value="{{.Key}}">
<label>Title <input name="title" value="{{.Post.Title}}" required></label>
<label>Content <textarea name="content">{{.Post.Content}}</textarea></label><button type="submit">Save</button></form>
</html>`))
