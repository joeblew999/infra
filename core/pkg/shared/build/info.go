package build

import (
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
)

// Info captures build metadata embedded via Go's module build information.
type Info struct {
	// Available reports whether build metadata was present.
	Available   bool
	Version     string
	Revision    string
	Modified    bool
	ModifiedSet bool
	BuildTime   string
	GoVersion   string
}

var (
	cached Info
	once   sync.Once
)

// Get returns the build metadata for the current binary. Values are cached on
// first access. When build information is unavailable (for example when running
// "go run"), the Version falls back to "development" and Available is false.
func Get() Info {
	once.Do(func() {
		cached = Info{Version: "development", GoVersion: runtime.Version()}
		if info, ok := debug.ReadBuildInfo(); ok {
			cached.Available = true
			cached.Version = sanitizeVersion(info.Main.Version)
			if info.GoVersion != "" {
				cached.GoVersion = info.GoVersion
			}
			for _, setting := range info.Settings {
				switch setting.Key {
				case "vcs.revision":
					cached.Revision = setting.Value
				case "vcs.modified":
					cached.Modified = strings.EqualFold(setting.Value, "true")
					cached.ModifiedSet = true
				case "vcs.time":
					cached.BuildTime = setting.Value
				}
			}
		}
	})
	return cached
}

func sanitizeVersion(v string) string {
	trimmed := strings.TrimSpace(v)
	if trimmed == "" || trimmed == "(devel)" {
		return "development"
	}
	return trimmed
}
