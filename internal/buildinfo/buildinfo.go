// Package buildinfo exposes values stamped into a Gossip release binary.
package buildinfo

import "runtime/debug"

// These variables are intentionally set with -ldflags by the release builder.
var (
	ID     = "dev"
	Commit = "unknown"
)

// Version is the immutable identity reported by a running binary.
type Version struct {
	BuildID string `json:"build_id"`
	Commit  string `json:"commit"`
	Dirty   bool   `json:"dirty"`
}

// Current returns the stamped build identity and best-effort VCS metadata.
func Current() Version {
	version := Version{BuildID: ID, Commit: Commit}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if version.Commit == "unknown" {
					version.Commit = setting.Value
				}
			case "vcs.modified":
				version.Dirty = setting.Value == "true"
			}
		}
	}
	return version
}
