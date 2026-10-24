// Package sample is a deliberately small native Go plugin used by the build
// and hook-ordering acceptance tests.
package sample

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/katsuakikureyama/Gossip/pkg/pluginsdk"
)

//go:embed plugin.json
var manifestJSON []byte

type Plugin struct{}

func (Plugin) Manifest() pluginsdk.Manifest {
	var manifest pluginsdk.Manifest
	if err := json.Unmarshal(manifestJSON, &manifest); err != nil {
		panic(fmt.Sprintf("sample plugin manifest: %v", err))
	}
	return manifest
}

func (Plugin) Register(registrar pluginsdk.Registrar) error {
	return registrar.Handle(pluginsdk.HookContentRender, 50, func(_ context.Context, event pluginsdk.Event) (pluginsdk.Event, error) {
		event.Fields["plugin_footer"] = "Rendered by the native Go sample plugin."
		return event, nil
	})
}
