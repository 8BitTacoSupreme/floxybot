package floxctx

import (
	"fmt"
	"strings"
)

// EnvVar is a key-value pair.
type EnvVar struct {
	Key   string
	Value string
}

// SystemInfo holds basic OS details.
type SystemInfo struct {
	OS   string
	Arch string
}

// Context holds the detected Flox environment state.
type Context struct {
	InFloxEnv      bool
	FloxEnv        string
	FloxEnvProject string
	FloxVars       []EnvVar
	Packages       []string
	Manifest       *ManifestData
	System         SystemInfo
}

// Format returns a human-readable summary of the context.
func (c *Context) Format() string {
	var b strings.Builder

	if !c.InFloxEnv {
		b.WriteString("Flox Environment: not detected\n")
		b.WriteString(fmt.Sprintf("System: %s (%s)\n", c.System.OS, c.System.Arch))
		return b.String()
	}

	b.WriteString("Flox Environment: active\n")
	if c.FloxEnv != "" {
		b.WriteString(fmt.Sprintf("  FLOX_ENV: %s\n", c.FloxEnv))
	}
	if c.FloxEnvProject != "" {
		b.WriteString(fmt.Sprintf("  FLOX_ENV_PROJECT: %s\n", c.FloxEnvProject))
	}

	if len(c.Packages) > 0 {
		b.WriteString(fmt.Sprintf("\nInstalled Packages (%d):\n", len(c.Packages)))
		for _, p := range c.Packages {
			b.WriteString(fmt.Sprintf("  - %s\n", p))
		}
	}

	if len(c.FloxVars) > 0 {
		b.WriteString("\nFlox Variables:\n")
		for _, v := range c.FloxVars {
			b.WriteString(fmt.Sprintf("  %s=%s\n", v.Key, v.Value))
		}
	}

	if c.Manifest != nil && len(c.Manifest.Install) > 0 {
		b.WriteString("\nManifest Packages:\n")
		for name, spec := range c.Manifest.Install {
			if spec.PkgPath != "" {
				b.WriteString(fmt.Sprintf("  %s: %s\n", name, spec.PkgPath))
			}
		}
	}

	b.WriteString(fmt.Sprintf("\nSystem: %s (%s)\n", c.System.OS, c.System.Arch))
	return b.String()
}

// ForPrompt returns a compact context string suitable for injection into an LLM system prompt.
func (c *Context) ForPrompt() string {
	if !c.InFloxEnv {
		return "User is NOT in a Flox environment."
	}

	var b strings.Builder
	b.WriteString("User is inside a Flox environment.\n")
	if c.FloxEnvProject != "" {
		b.WriteString(fmt.Sprintf("Project: %s\n", c.FloxEnvProject))
	}
	if len(c.Packages) > 0 {
		b.WriteString(fmt.Sprintf("Installed packages: %s\n", strings.Join(c.Packages, ", ")))
	}
	if c.Manifest != nil && c.Manifest.Raw != "" {
		b.WriteString("\nmanifest.toml:\n```toml\n")
		b.WriteString(c.Manifest.Raw)
		b.WriteString("\n```\n")
	}
	return b.String()
}
