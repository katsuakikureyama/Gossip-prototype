package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/katsuakikureyama/Gossip/internal/buildinfo"
	"github.com/katsuakikureyama/Gossip/internal/cms"
	"github.com/katsuakikureyama/Gossip/pkg/pluginsdk"
	"github.com/katsuakikureyama/Gossip/plugins/sample"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:8080", "HTTP listen address")
	devEditor := flag.Bool("dev-editor", false, "enable the local fixture editor")
	dataFile := flag.String("data-file", ".gossip/content.json", "local content snapshot (empty for in-memory fixture)")
	versionJSON := flag.Bool("version-json", false, "print build identity and exit")
	flag.Parse()

	if *versionJSON {
		if err := json.NewEncoder(os.Stdout).Encode(buildinfo.Current()); err != nil {
			log.Fatal(err)
		}
		return
	}

	registry := pluginsdk.NewRegistry()
	if err := registry.Register(sample.Plugin{}); err != nil {
		log.Fatalf("register sample plugin: %v", err)
	}
	seed := cms.Post{
		Slug:      "welcome",
		Title:     "Welcome to Gossip",
		Content:   "Welcome to the Gossip development fixture.",
		UpdatedAt: time.Now().UTC(),
	}
	var store cms.Store = cms.NewMemoryStore(seed)
	if *dataFile != "" {
		persistent, err := cms.OpenFileStore(*dataFile, seed)
		if err != nil {
			log.Fatal(err)
		}
		defer persistent.Close()
		store = persistent
	}
	handler, err := cms.NewHandler(cms.Config{
		Store:        store,
		Plugins:      registry,
		DevEditor:    *devEditor,
		DevEditorKey: os.Getenv("GOSSIP_DEV_EDITOR_KEY"),
	})
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:              *listen,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	fmt.Printf("Gossip %s listening on http://%s\n", buildinfo.Current().BuildID, *listen)
	log.Fatal(server.ListenAndServe())
}
