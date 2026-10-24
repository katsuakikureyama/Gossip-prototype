package compat_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/katsuakikureyama/Gossip/pkg/pluginsdk"
)

type hookHandler struct {
	Priority int    `json:"priority"`
	Prefix   string `json:"prefix"`
	Suffix   string `json:"suffix"`
}

type hookCase struct {
	ID       string        `json:"id"`
	Input    string        `json:"input"`
	Handlers []hookHandler `json:"handlers"`
}

type observation struct {
	Value string `json:"value"`
	Trace []int  `json:"trace"`
}

type fixturePlugin struct {
	fixture hookCase
	trace   *[]int
}

func (p fixturePlugin) Manifest() pluginsdk.Manifest {
	return pluginsdk.Manifest{SchemaVersion: 1, ID: "fixture.hooks", Name: "Owned hook fixture",
		Version: "0.1.0", Kind: "native-go", GossipAPI: pluginsdk.APIVersion,
		EntryPackage: "github.com/katsuakikureyama/Gossip/tests/compat", License: "NOASSERTION"}
}

func (p fixturePlugin) Register(r pluginsdk.Registrar) error {
	for index, handler := range p.fixture.Handlers {
		index, handler := index, handler
		if err := r.Handle(pluginsdk.HookContentRender, handler.Priority,
			func(_ context.Context, event pluginsdk.Event) (pluginsdk.Event, error) {
				*p.trace = append(*p.trace, index)
				event.Fields["content"] = handler.Prefix + event.Fields["content"] + handler.Suffix
				return event, nil
			}); err != nil {
			return err
		}
	}
	return nil
}

// Opt-in integration test: a normal unit test run must not pull/run Docker.
func TestWordPressHookDifferential(t *testing.T) {
	image := os.Getenv("GOSSIP_WP_ORACLE_IMAGE")
	if image == "" {
		t.Skip("set GOSSIP_WP_ORACLE_IMAGE to an already pulled WordPress digest to run the real oracle")
	}
	if !regexp.MustCompile(`^wordpress@sha256:[0-9a-f]{64}$`).MatchString(image) {
		t.Fatal("oracle must be an immutable official WordPress digest")
	}
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	fixtureBytes, err := os.ReadFile(filepath.Join(dir, "hooks-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Schema string     `json:"schema"`
		Cases  []hookCase `json:"cases"`
	}
	if err := json.Unmarshal(fixtureBytes, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Schema != "gossip.hook-fixture/v1" || len(fixture.Cases) == 0 {
		t.Fatal("invalid fixture")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "docker", "run", "--rm", "--pull=never",
		"--network=none", "--read-only", "--cap-drop=ALL", "--security-opt=no-new-privileges",
		"--user=65532:65532", "--memory=128m", "--pids-limit=32", "--cpus=1",
		"--mount", "type=bind,source="+dir+",target=/fixtures,readonly",
		"--entrypoint=php", image, "/fixtures/hooks-oracle.php")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("WordPress oracle: %v\n%s", err, output)
	}
	var oracle struct {
		Schema    string                   `json:"schema"`
		WordPress string                   `json:"wordpress"`
		PHP       string                   `json:"php"`
		Results   map[string][]observation `json:"results"`
	}
	if err := json.Unmarshal(output, &oracle); err != nil {
		t.Fatalf("oracle JSON: %v\n%s", err, output)
	}
	if oracle.Schema != "gossip.hook-observation/v1" || oracle.WordPress != "6.6.2" {
		t.Fatalf("unexpected oracle schema/version: %s / %s", oracle.Schema, oracle.WordPress)
	}
	if len(oracle.Results) != len(fixture.Cases) {
		t.Fatal("oracle case count mismatch")
	}
	t.Logf("image=%s wordpress=%s php=%s fixture_sha256=%x", image, oracle.WordPress, oracle.PHP, sha256.Sum256(fixtureBytes))
	for _, testCase := range fixture.Cases {
		t.Run(testCase.ID, func(t *testing.T) {
			expected, exists := oracle.Results[testCase.ID]
			if !exists || len(expected) != 2 {
				t.Fatal("missing oracle runs")
			}
			registry := pluginsdk.NewRegistry()
			trace := []int{}
			if err := registry.Register(fixturePlugin{fixture: testCase, trace: &trace}); err != nil {
				t.Fatal(err)
			}
			for run := 0; run < 2; run++ {
				trace = []int{}
				event, err := registry.Apply(ctx, pluginsdk.HookContentRender,
					pluginsdk.Event{Resource: "fixture:" + testCase.ID, Fields: map[string]string{"content": testCase.Input}})
				if err != nil {
					t.Fatal(err)
				}
				actual := observation{Value: event.Fields["content"], Trace: trace}
				if !reflect.DeepEqual(actual, expected[run]) {
					t.Fatalf("run %d: Gossip=%+v WordPress=%+v", run, actual, expected[run])
				}
				t.Log(fmt.Sprintf("run=%d trace=%v value=%q", run, trace, actual.Value))
			}
		})
	}
}
