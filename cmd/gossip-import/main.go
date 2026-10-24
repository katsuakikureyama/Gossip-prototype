// gossip-import is an offline, development-only WP-CLI fixture importer.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/katsuakikureyama/Gossip/internal/cms"
	"github.com/katsuakikureyama/Gossip/internal/wpimport"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("gossip-import", flag.ContinueOnError)
	flags.SetOutput(stderr)
	input := flags.String("input", "", "owned WP-CLI JSON export to inspect")
	output := flags.String("out", "", "new snapshot path; existing paths are never replaced")
	commit := flags.Bool("commit", false, "create the snapshot after reviewing a preview")
	expected := flags.String("expect-sha256", "", "source hash from the reviewed preview (required for commit)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *input == "" {
		return errors.New("-input is required; positional arguments are not accepted")
	}
	if *commit && (*output == "" || *expected == "") {
		return errors.New("commit requires -out and -expect-sha256 from a reviewed preview")
	}
	if !*commit && (*output != "" || *expected != "") {
		return errors.New("preview accepts only -input; add -commit explicitly to save")
	}
	file, err := os.Open(*input)
	if err != nil {
		return err
	}
	defer file.Close()
	plan, err := wpimport.Prepare(file)
	if err != nil {
		return err
	}
	if *commit {
		if *expected != plan.SourceSHA256 {
			return errors.New("source hash differs from reviewed preview; no snapshot created")
		}
		store, err := cms.CreateFileStore(*output, plan.Posts...)
		if err != nil {
			return fmt.Errorf("create snapshot: %w", err)
		}
		if err := store.Close(); err != nil {
			return fmt.Errorf("snapshot created but lock close failed; inspect output before retry: %w", err)
		}
		plan.Status = "saved"
	}
	return json.NewEncoder(stdout).Encode(plan)
}
