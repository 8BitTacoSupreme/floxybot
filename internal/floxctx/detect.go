package floxctx

import (
	"os"
	"os/exec"
	"strings"
)

// Detect gathers information about the current Flox environment.
func Detect() (*Context, error) {
	ctx := &Context{}

	ctx.FloxEnv = os.Getenv("FLOX_ENV")
	ctx.FloxEnvProject = os.Getenv("FLOX_ENV_PROJECT")
	ctx.InFloxEnv = ctx.FloxEnv != "" || ctx.FloxEnvProject != ""

	// Collect all FLOX_* env vars.
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "FLOX") {
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) == 2 {
				ctx.FloxVars = append(ctx.FloxVars, EnvVar{Key: parts[0], Value: parts[1]})
			}
		}
	}

	// Get installed packages via `flox list`.
	if ctx.InFloxEnv {
		ctx.Packages = listPackages()
	}

	// Read manifest if available.
	if ctx.FloxEnvProject != "" {
		manifest, err := readManifest(ctx.FloxEnvProject)
		if err == nil {
			ctx.Manifest = manifest
		}
	}

	// System info.
	ctx.System = detectSystem()

	return ctx, nil
}

func listPackages() []string {
	out, err := exec.Command("flox", "list").Output()
	if err != nil {
		return nil
	}
	var pkgs []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "=") || strings.HasPrefix(line, "-") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			pkgs = append(pkgs, fields[0])
		}
	}
	return pkgs
}

func detectSystem() SystemInfo {
	info := SystemInfo{}
	if out, err := exec.Command("uname", "-s").Output(); err == nil {
		info.OS = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("uname", "-m").Output(); err == nil {
		info.Arch = strings.TrimSpace(string(out))
	}
	return info
}
