package floxctx

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// ManifestData holds parsed manifest.toml fields we care about.
type ManifestData struct {
	Install  map[string]PackageSpec `toml:"install"`
	Vars     map[string]string      `toml:"vars"`
	Services map[string]ServiceSpec `toml:"services"`
	Raw      string                 // full file contents
}

type PackageSpec struct {
	PkgPath string `toml:"pkg-path"`
	Version string `toml:"version"`
}

type ServiceSpec struct {
	Command string `toml:"command"`
}

func readManifest(projectDir string) (*ManifestData, error) {
	path := filepath.Join(projectDir, ".flox", "env", "manifest.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	m := &ManifestData{Raw: string(data)}
	if _, err := toml.Decode(m.Raw, m); err != nil {
		// Still return raw content even if structured parse has issues.
		return m, nil
	}
	return m, nil
}
