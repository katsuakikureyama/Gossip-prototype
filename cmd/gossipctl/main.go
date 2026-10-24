// gossipctl builds and selects immutable Gossip release binaries. It does not
// install untrusted source or run metadata scripts.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/katsuakikureyama/Gossip/internal/buildinfo"
	"github.com/katsuakikureyama/Gossip/internal/release"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "gossipctl:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return usageError()
	}
	switch args[0] {
	case "build":
		return buildCommand(ctx, args[1:])
	case "activate":
		return activateCommand(ctx, args[1:])
	case "status":
		return statusCommand(args[1:])
	default:
		return usageError()
	}
}

func buildCommand(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("build", flag.ContinueOnError)
	root := flags.String("root", ".gossip", "release state directory")
	slotText := flags.String("slot", "", "candidate slot: A or B")
	buildID := flags.String("build-id", "", "immutable build identifier")
	packagePath := flags.String("package", "./cmd/gossip", "Gossip main package")
	source := flags.String("source", ".", "trusted source checkout")
	commit := flags.String("commit", "unknown", "source commit stamped into the binary")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *slotText == "" || *buildID == "" {
		return errors.New("build requires --slot and --build-id")
	}
	manager, err := release.NewManager(*root, verifyBinary)
	if err != nil {
		return err
	}
	builder := func(ctx context.Context, output string) error {
		absoluteOutput, err := filepath.Abs(output)
		if err != nil {
			return err
		}
		ldflags := strings.Join([]string{
			"-s", "-w",
			"-X", "github.com/katsuakikureyama/Gossip/internal/buildinfo.ID=" + *buildID,
			"-X", "github.com/katsuakikureyama/Gossip/internal/buildinfo.Commit=" + *commit,
		}, " ")
		command := exec.CommandContext(ctx, "go", "build", "-trimpath", "-ldflags", ldflags, "-o", absoluteOutput, *packagePath)
		command.Dir = *source
		command.Env = append(os.Environ(), "CGO_ENABLED=0")
		outputText, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("go build: %w: %s", err, strings.TrimSpace(string(outputText)))
		}
		return nil
	}
	record, err := manager.Build(ctx, release.Slot(*slotText), *buildID, builder)
	if err != nil {
		return err
	}
	return printJSON(record)
}

func activateCommand(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("activate", flag.ContinueOnError)
	root := flags.String("root", ".gossip", "release state directory")
	slotText := flags.String("slot", "", "slot to activate: A or B")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *slotText == "" {
		return errors.New("activate requires --slot")
	}
	manager, err := release.NewManager(*root, verifyBinary)
	if err != nil {
		return err
	}
	return manager.Activate(ctx, release.Slot(*slotText))
}

func statusCommand(args []string) error {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
	root := flags.String("root", ".gossip", "release state directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return usageError()
	}
	manager, err := release.NewManager(*root, verifyBinary)
	if err != nil {
		return err
	}
	state, err := manager.Load()
	if err != nil {
		return err
	}
	return printJSON(state)
}

func verifyBinary(ctx context.Context, binaryPath, expectedBuildID string) error {
	verifyCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	command := exec.CommandContext(verifyCtx, binaryPath, "--version-json")
	output, err := command.Output()
	if err != nil {
		return fmt.Errorf("run version probe: %w", err)
	}
	var version buildinfo.Version
	if err := json.Unmarshal(output, &version); err != nil {
		return fmt.Errorf("decode version probe: %w", err)
	}
	if version.BuildID != expectedBuildID {
		return fmt.Errorf("build ID = %q, want %q", version.BuildID, expectedBuildID)
	}
	return nil
}

func printJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func usageError() error {
	return errors.New("usage: gossipctl <build|activate|status> [flags]")
}
