package builder

import (
	"bytes"
	"os/exec"
	"runtime/debug"
	"strings"
)

var (
	// Set these at build time with -ldflags "-X 'github.com/idlab-discover/AIBoMGen-cli/internal/builder.Version=...' -X '...Commit=...'"
	Version = ""
	Commit  = ""
)

var readBuildInfo = debug.ReadBuildInfo

func GetAIBoMGenVersion() string {
	// 1) prefer explicit ldflags
	if Version != "" && Version != "dev" {
		return Version
	}
	// 2) module build info
	if info, ok := readBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	// 3) git describe fallback
	if d := gitDescribe(); d != "" {
		return d
	}
	// 4) commit fallback
	if Commit != "" {
		return "commit-" + Commit
	}
	return "devel"
}

func gitDescribe() string {
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	out, err := cmd.Output()
	if err != nil {
		cmd2 := exec.Command("git", "rev-parse", "--short", "HEAD")
		if out2, err2 := cmd2.Output(); err2 == nil {
			return strings.TrimSpace(string(out2))
		}
		return ""
	}
	return strings.TrimSpace(string(bytes.TrimSpace(out)))
}
