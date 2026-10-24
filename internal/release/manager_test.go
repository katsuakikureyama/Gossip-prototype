package release

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFailedBuildLeavesSlotAAndActiveSelectionUntouched(t *testing.T) {
	manager := newTestManager(t)
	goodBuilder := func(contents string) Builder {
		return func(_ context.Context, output string) error {
			return os.WriteFile(output, []byte(contents), 0o755)
		}
	}
	if _, err := manager.Build(context.Background(), SlotA, "good-a", goodBuilder("slot A")); err != nil {
		t.Fatal(err)
	}
	before, err := manager.Load()
	if err != nil {
		t.Fatal(err)
	}
	failure := errors.New("compile failed")
	if _, err := manager.Build(context.Background(), SlotB, "bad-b", func(context.Context, string) error { return failure }); !errors.Is(err, failure) {
		t.Fatalf("build error = %v", err)
	}
	after, err := manager.Load()
	if err != nil {
		t.Fatal(err)
	}
	if after.ActiveSlot != SlotA || after.Slots[SlotA] != before.Slots[SlotA] {
		t.Fatalf("slot A changed after failed build: before=%#v after=%#v", before, after)
	}
	if _, exists := after.Slots[SlotB]; exists {
		t.Fatalf("failed slot B was published: %#v", after.Slots[SlotB])
	}
}

func TestBuildSlotBThenActivateAndRejectTampering(t *testing.T) {
	manager := newTestManager(t)
	builder := func(_ context.Context, output string) error {
		return os.WriteFile(output, []byte("slot B"), 0o755)
	}
	record, err := manager.Build(context.Background(), SlotB, "good-b", builder)
	if err != nil {
		t.Fatal(err)
	}
	state, _ := manager.Load()
	if state.ActiveSlot != SlotA || state.Slots[SlotB].BuildID != "good-b" {
		t.Fatalf("unexpected pre-activation state: %#v", state)
	}
	if err := manager.Activate(context.Background(), SlotB); err != nil {
		t.Fatal(err)
	}
	state, _ = manager.Load()
	if state.ActiveSlot != SlotB {
		t.Fatalf("active slot = %s", state.ActiveSlot)
	}

	binary := filepath.Join(manager.root, filepath.FromSlash(record.Binary))
	if err := os.Chmod(binary, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binary, []byte("tampered"), 0o555); err != nil {
		t.Fatal(err)
	}
	if err := manager.Activate(context.Background(), SlotB); err == nil {
		t.Fatal("expected digest mismatch")
	}
}

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	manager, err := NewManager(t.TempDir(), func(_ context.Context, binaryPath, _ string) error {
		info, err := os.Stat(binaryPath)
		if err != nil {
			return err
		}
		if info.Size() == 0 {
			return errors.New("empty binary")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return manager
}
