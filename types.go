package bptest

import (
	"bytes"

	"github.com/BurntSushi/toml"
)

// Binding represents a service binding for testing.
type Binding struct {
	Name     string
	Type     string
	Provider string
	Secret   map[string]string
}

// TargetInfo represents the target platform information.
type TargetInfo struct {
	OS            string
	Arch          string
	DistroName    string
	DistroVersion string
}

// BuildpackConfig holds buildpack metadata for generating buildpack.toml.
type BuildpackConfig struct {
	API     string
	ID      string
	Version string
	Name    string
}

// ToTOML serializes the BuildpackConfig to TOML format.
func (c BuildpackConfig) ToTOML() ([]byte, error) {
	data := map[string]any{
		"api": c.API,
		"buildpack": map[string]any{
			"id":      c.ID,
			"version": c.Version,
		},
	}
	if c.Name != "" {
		data["buildpack"].(map[string]any)["name"] = c.Name
	}

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// LayerConfig configures an existing layer for rebuild scenarios.
type LayerConfig struct {
	Build    bool
	Launch   bool
	Cache    bool
	Metadata map[string]any
	// Files maps relative paths to file contents for pre-populating the layer.
	Files map[string][]byte
	// Env contains environment variables to set in the layer's env directories.
	// Keys are in the format "VAR.default", "VAR.override", etc.
	Env map[string]string
}

// TestPaths holds the directory paths used during test execution.
type TestPaths struct {
	AppDir       string
	BuildpackDir string
	PlatformDir  string
	LayersDir    string
}
