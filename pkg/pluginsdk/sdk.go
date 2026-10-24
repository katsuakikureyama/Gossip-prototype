// Package pluginsdk defines the versioned contract for native Go plugins.
//
// Native plugins are linked into a Gossip release binary. They are trusted,
// in-process code; installing a manifest never executes the plugin source.
package pluginsdk

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
)

const APIVersion = "v1alpha1"

// Hook is a stable extension point. New hooks must be additive within an API
// version; incompatible changes require a new API version.
type Hook string

const (
	HookContentBeforeSave Hook = "content.before_save"
	HookContentRender     Hook = "content.render"
)

// Manifest is the machine-readable identity and build contract of a plugin.
type Manifest struct {
	SchemaVersion int    `json:"schema_version"`
	ID            string `json:"id"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	Kind          string `json:"kind"`
	GossipAPI     string `json:"gossip_api"`
	EntryPackage  string `json:"entry_package"`
	License       string `json:"license"`
}

var (
	idPattern      = regexp.MustCompile(`^[a-z0-9]+(?:[.-][a-z0-9]+)*$`)
	versionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:[-+][0-9A-Za-z.-]+)?$`)
)

// Validate rejects ambiguous or unsupported manifests before any build begins.
func (m Manifest) Validate() error {
	var problems []string
	if m.SchemaVersion != 1 {
		problems = append(problems, "schema_version must be 1")
	}
	if !idPattern.MatchString(m.ID) {
		problems = append(problems, "id must be lowercase dot/dash-separated tokens")
	}
	if strings.TrimSpace(m.Name) == "" {
		problems = append(problems, "name is required")
	}
	if !versionPattern.MatchString(m.Version) {
		problems = append(problems, "version must be semantic version text")
	}
	if m.Kind != "native-go" {
		problems = append(problems, "kind must be native-go")
	}
	if m.GossipAPI != APIVersion {
		problems = append(problems, fmt.Sprintf("gossip_api must be %s", APIVersion))
	}
	if strings.TrimSpace(m.EntryPackage) == "" {
		problems = append(problems, "entry_package is required")
	}
	if strings.TrimSpace(m.License) == "" {
		problems = append(problems, "license is required")
	}
	return errors.Join(stringErrors(problems)...)
}

func stringErrors(values []string) []error {
	errs := make([]error, 0, len(values))
	for _, value := range values {
		errs = append(errs, errors.New(value))
	}
	return errs
}

// Event is the versioned data envelope passed to hook handlers.
type Event struct {
	Resource string            `json:"resource"`
	Fields   map[string]string `json:"fields"`
}

// Handler may transform an event or reject an operation.
type Handler func(context.Context, Event) (Event, error)

// Registrar is intentionally narrow so translated and native plugins can use
// the same deterministic hook ordering contract through adapters.
type Registrar interface {
	Handle(hook Hook, priority int, handler Handler) error
}

// Plugin is implemented by statically linked native Go plugins.
type Plugin interface {
	Manifest() Manifest
	Register(Registrar) error
}

type registeredHandler struct {
	pluginID string
	priority int
	order    int
	handler  Handler
}

// Registry stores plugin identities and deterministic hook handlers.
type Registry struct {
	mu       sync.RWMutex
	plugins  map[string]Manifest
	handlers map[Hook][]registeredHandler
	order    int
}

func NewRegistry() *Registry {
	return &Registry{
		plugins:  make(map[string]Manifest),
		handlers: make(map[Hook][]registeredHandler),
	}
}

// Register validates a plugin and atomically publishes its handlers. A failed
// registration never leaves a partial plugin behind.
func (r *Registry) Register(plugin Plugin) error {
	if plugin == nil {
		return errors.New("plugin is nil")
	}
	manifest := plugin.Manifest()
	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("plugin manifest %q: %w", manifest.ID, err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.plugins[manifest.ID]; exists {
		return fmt.Errorf("plugin %q is already registered", manifest.ID)
	}

	pending := &pendingRegistrar{pluginID: manifest.ID}
	if err := plugin.Register(pending); err != nil {
		return fmt.Errorf("register plugin %q: %w", manifest.ID, err)
	}
	for _, item := range pending.handlers {
		r.order++
		item.order = r.order
		r.handlers[item.hook] = append(r.handlers[item.hook], item.registeredHandler)
		sort.SliceStable(r.handlers[item.hook], func(i, j int) bool {
			left, right := r.handlers[item.hook][i], r.handlers[item.hook][j]
			if left.priority != right.priority {
				return left.priority < right.priority
			}
			return left.order < right.order
		})
	}
	r.plugins[manifest.ID] = manifest
	return nil
}

// Apply executes handlers by ascending priority, then registration order.
func (r *Registry) Apply(ctx context.Context, hook Hook, event Event) (Event, error) {
	r.mu.RLock()
	handlers := append([]registeredHandler(nil), r.handlers[hook]...)
	r.mu.RUnlock()

	current := cloneEvent(event)
	for _, item := range handlers {
		next, err := item.handler(ctx, cloneEvent(current))
		if err != nil {
			return event, fmt.Errorf("plugin %q hook %q: %w", item.pluginID, hook, err)
		}
		current = cloneEvent(next)
	}
	return current, nil
}

// Manifests returns a stable snapshot for diagnostics and compatibility data.
func (r *Registry) Manifests() []Manifest {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Manifest, 0, len(r.plugins))
	for _, manifest := range r.plugins {
		result = append(result, manifest)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

type pendingHandler struct {
	hook Hook
	registeredHandler
}

type pendingRegistrar struct {
	pluginID string
	handlers []pendingHandler
}

func (p *pendingRegistrar) Handle(hook Hook, priority int, handler Handler) error {
	if hook != HookContentBeforeSave && hook != HookContentRender {
		return fmt.Errorf("unknown hook %q", hook)
	}
	if priority < 0 || priority > 999 {
		return errors.New("priority must be between 0 and 999")
	}
	if handler == nil {
		return errors.New("handler is nil")
	}
	p.handlers = append(p.handlers, pendingHandler{
		hook: hook,
		registeredHandler: registeredHandler{
			pluginID: p.pluginID,
			priority: priority,
			handler:  handler,
		},
	})
	return nil
}

func cloneEvent(event Event) Event {
	copy := Event{Resource: event.Resource, Fields: make(map[string]string, len(event.Fields))}
	for key, value := range event.Fields {
		copy.Fields[key] = value
	}
	return copy
}
