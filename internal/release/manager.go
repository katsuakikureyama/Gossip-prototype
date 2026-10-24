// Package release manages immutable binaries and an atomic A/B slot pointer.
package release

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

const schemaVersion = 1

type Slot string

const (
	SlotA Slot = "A"
	SlotB Slot = "B"
)

func (s Slot) Valid() bool { return s == SlotA || s == SlotB }

type Record struct {
	BuildID string    `json:"build_id"`
	Binary  string    `json:"binary"`
	SHA256  string    `json:"sha256"`
	Created time.Time `json:"created_at"`
}

type State struct {
	SchemaVersion int             `json:"schema_version"`
	ActiveSlot    Slot            `json:"active_slot"`
	Slots         map[Slot]Record `json:"slots"`
}

type Builder func(ctx context.Context, outputPath string) error
type Verifier func(ctx context.Context, binaryPath, buildID string) error

type Manager struct {
	root   string
	verify Verifier
	now    func() time.Time
}

func NewManager(root string, verify Verifier) (*Manager, error) {
	if root == "" {
		return nil, errors.New("release root is required")
	}
	if verify == nil {
		return nil, errors.New("release verifier is required")
	}
	return &Manager{root: root, verify: verify, now: time.Now}, nil
}

var buildIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

// Build compiles into an isolated temporary path, verifies the candidate, then
// publishes the immutable binary and updates only the requested inactive slot.
// No state file is changed on build or verification failure.
func (m *Manager) Build(ctx context.Context, slot Slot, buildID string, builder Builder) (Record, error) {
	if !slot.Valid() {
		return Record{}, fmt.Errorf("invalid slot %q", slot)
	}
	if !buildIDPattern.MatchString(buildID) {
		return Record{}, errors.New("invalid build ID")
	}
	if builder == nil {
		return Record{}, errors.New("builder is required")
	}
	state, err := m.Load()
	if err != nil {
		return Record{}, err
	}
	if _, populated := state.Slots[slot]; populated && slot == state.ActiveSlot {
		return Record{}, fmt.Errorf("slot %s is active; build into the inactive slot", slot)
	}
	finalDir := filepath.Join(m.root, "releases", buildID)
	if _, err := os.Stat(finalDir); err == nil {
		return Record{}, fmt.Errorf("release %q already exists", buildID)
	} else if !errors.Is(err, os.ErrNotExist) {
		return Record{}, err
	}
	if err := os.MkdirAll(filepath.Join(m.root, "releases"), 0o755); err != nil {
		return Record{}, err
	}
	tempDir, err := os.MkdirTemp(filepath.Join(m.root, "releases"), ".candidate-")
	if err != nil {
		return Record{}, err
	}
	defer os.RemoveAll(tempDir)
	tempBinary := filepath.Join(tempDir, "gossip")
	if err := builder(ctx, tempBinary); err != nil {
		return Record{}, fmt.Errorf("build candidate: %w", err)
	}
	if err := m.verify(ctx, tempBinary, buildID); err != nil {
		return Record{}, fmt.Errorf("verify candidate: %w", err)
	}
	digest, err := fileSHA256(tempBinary)
	if err != nil {
		return Record{}, err
	}
	if err := os.Chmod(tempBinary, 0o555); err != nil {
		return Record{}, err
	}
	if err := os.Rename(tempDir, finalDir); err != nil {
		return Record{}, fmt.Errorf("publish release: %w", err)
	}

	record := Record{
		BuildID: buildID,
		Binary:  filepath.ToSlash(filepath.Join("releases", buildID, "gossip")),
		SHA256:  digest,
		Created: m.now().UTC(),
	}
	state.Slots[slot] = record
	if err := m.save(state); err != nil {
		return Record{}, fmt.Errorf("save slot state (release is orphaned but inactive): %w", err)
	}
	return record, nil
}

// Activate atomically selects an already verified slot. It never rebuilds.
func (m *Manager) Activate(ctx context.Context, slot Slot) error {
	if !slot.Valid() {
		return fmt.Errorf("invalid slot %q", slot)
	}
	state, err := m.Load()
	if err != nil {
		return err
	}
	record, exists := state.Slots[slot]
	if !exists {
		return fmt.Errorf("slot %s has no candidate", slot)
	}
	binary := filepath.Join(m.root, filepath.FromSlash(record.Binary))
	if err := m.verify(ctx, binary, record.BuildID); err != nil {
		return fmt.Errorf("verify slot %s: %w", slot, err)
	}
	digest, err := fileSHA256(binary)
	if err != nil {
		return err
	}
	if digest != record.SHA256 {
		return fmt.Errorf("slot %s digest mismatch", slot)
	}
	state.ActiveSlot = slot
	return m.save(state)
}

func (m *Manager) Load() (State, error) {
	state := State{SchemaVersion: schemaVersion, ActiveSlot: SlotA, Slots: make(map[Slot]Record)}
	data, err := os.ReadFile(filepath.Join(m.root, "state.json"))
	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}
	if err != nil {
		return State{}, err
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, fmt.Errorf("decode release state: %w", err)
	}
	if state.SchemaVersion != schemaVersion || !state.ActiveSlot.Valid() || state.Slots == nil {
		return State{}, errors.New("invalid release state")
	}
	return state, nil
}

func (m *Manager) save(state State) error {
	if err := os.MkdirAll(m.root, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	temp, err := os.CreateTemp(m.root, ".state-")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if err := temp.Chmod(0o600); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempName, filepath.Join(m.root, "state.json"))
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
