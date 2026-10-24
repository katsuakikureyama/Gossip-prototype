package pluginsdk

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type testPlugin struct {
	id       string
	priority int
	value    string
	fail     bool
}

func (p testPlugin) Manifest() Manifest {
	return Manifest{
		SchemaVersion: 1,
		ID:            p.id,
		Name:          p.id,
		Version:       "0.1.0",
		Kind:          "native-go",
		GossipAPI:     APIVersion,
		EntryPackage:  "example.invalid/" + p.id,
		License:       "MIT",
	}
}

func (p testPlugin) Register(registrar Registrar) error {
	return registrar.Handle(HookContentRender, p.priority, func(_ context.Context, event Event) (Event, error) {
		if p.fail {
			return event, errors.New("expected failure")
		}
		event.Fields["trace"] += p.value
		return event, nil
	})
}

func TestRegistryUsesPriorityThenRegistrationOrder(t *testing.T) {
	registry := NewRegistry()
	for _, plugin := range []testPlugin{
		{id: "third", priority: 20, value: "C"},
		{id: "first", priority: 10, value: "A"},
		{id: "second", priority: 10, value: "B"},
	} {
		if err := registry.Register(plugin); err != nil {
			t.Fatal(err)
		}
	}

	event, err := registry.Apply(context.Background(), HookContentRender, Event{Fields: map[string]string{"trace": ""}})
	if err != nil {
		t.Fatal(err)
	}
	if got := event.Fields["trace"]; got != "ABC" {
		t.Fatalf("trace = %q, want ABC", got)
	}
}

func TestRegistryDoesNotPublishFailedRegistration(t *testing.T) {
	registry := NewRegistry()
	bad := registrationFailurePlugin{}
	if err := registry.Register(bad); err == nil {
		t.Fatal("expected registration failure")
	}
	if got := registry.Manifests(); !reflect.DeepEqual(got, []Manifest{}) {
		t.Fatalf("published manifests after failure: %#v", got)
	}
}

type registrationFailurePlugin struct{}

func (registrationFailurePlugin) Manifest() Manifest {
	return testPlugin{id: "broken"}.Manifest()
}

func (registrationFailurePlugin) Register(Registrar) error {
	return errors.New("register failed")
}
