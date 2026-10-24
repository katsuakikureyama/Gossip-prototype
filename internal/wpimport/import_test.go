package wpimport

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func fixtureRow() map[string]any {
	return map[string]any{"ID": 42, "post_type": "post", "post_status": "publish", "post_password": "", "post_name": "owned-post", "post_title": "日本語 <b>title</b>", "post_content": "<!-- wp:paragraph -->\n<p>本文</p>\n<!-- /wp:paragraph -->", "post_modified_gmt": "2024-10-01 12:30:00"}
}

func TestPreparePreservesRawFields(t *testing.T) {
	row := fixtureRow()
	data, _ := json.Marshal([]any{row})
	plan, err := Prepare(strings.NewReader(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	if plan.Status != "prepared" || len(plan.SourceSHA256) != 64 || len(plan.Posts) != 1 || plan.Entries[0].SourceID != 42 {
		t.Fatalf("invalid plan: %+v", plan)
	}
	p := plan.Posts[0]
	if p.Title != row["post_title"] || p.Content != row["post_content"] || p.Slug != row["post_name"] || !p.UpdatedAt.Equal(time.Date(2024, 10, 1, 12, 30, 0, 0, time.UTC)) {
		t.Fatalf("mapping changed source fields: %+v", p)
	}
}

func TestPrepareRejectsUnsupportedBatch(t *testing.T) {
	for name, mutate := range map[string]func(map[string]any){
		"draft":              func(r map[string]any) { r["post_status"] = "draft" },
		"private":            func(r map[string]any) { r["post_status"] = "private" },
		"password":           func(r map[string]any) { r["post_password"] = "secret" },
		"page":               func(r map[string]any) { r["post_type"] = "page" },
		"missing visibility": func(r map[string]any) { delete(r, "post_status") },
		"null password":      func(r map[string]any) { r["post_password"] = nil },
		"missing content":    func(r map[string]any) { delete(r, "post_content") },
		"unsupported slug":   func(r map[string]any) { r["post_name"] = "../escape" },
		"encoded slug":       func(r map[string]any) { r["post_name"] = "%e6%97%a5" },
		"blank title":        func(r map[string]any) { r["post_title"] = "  " },
		"bad timestamp":      func(r map[string]any) { r["post_modified_gmt"] = "0000-00-00 00:00:00" },
		"fractional ID":      func(r map[string]any) { r["ID"] = 1.5 },
		"missing ID":         func(r map[string]any) { delete(r, "ID") },
		"unknown metadata":   func(r map[string]any) { r["post_author"] = 1 },
	} {
		t.Run(name, func(t *testing.T) {
			row := fixtureRow()
			row["ID"], row["post_name"] = 43, "second-post"
			mutate(row)
			data, _ := json.Marshal([]any{fixtureRow(), row})
			plan, err := Prepare(strings.NewReader(string(data)))
			if err == nil || len(plan.Posts) != 0 {
				t.Fatalf("unsupported batch accepted partially: %+v %v", plan, err)
			}
		})
	}
	for _, identity := range []string{"ID", "post_name"} {
		t.Run("duplicate "+identity, func(t *testing.T) {
			row := fixtureRow()
			if identity == "ID" {
				row["post_name"] = "other"
			} else {
				row["ID"] = 43
			}
			data, _ := json.Marshal([]any{fixtureRow(), row})
			if _, err := Prepare(strings.NewReader(string(data))); err == nil {
				t.Fatal("duplicate identity accepted")
			}
		})
	}
}

func TestPrepareRejectsInvalidEnvelope(t *testing.T) {
	for _, input := range []string{"null", "[]", "{}", "[", "[] {}", "\xff", strings.Repeat(" ", MaxInputBytes+1)} {
		if _, err := Prepare(strings.NewReader(input)); err == nil {
			t.Fatal("invalid envelope accepted")
		}
	}
}
