// Package toml provides TOML helpers for the bptest package.
package toml

import (
	"os"

	"github.com/BurntSushi/toml"
)

// Read reads and decodes a TOML file into the given type.
func Read[T any](path string) (T, error) {
	var result T
	data, err := os.ReadFile(path)
	if err != nil {
		return result, err
	}
	if err := toml.Unmarshal(data, &result); err != nil {
		return result, err
	}
	return result, nil
}

// Write encodes and writes a value to a TOML file.
func Write(path string, v any) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(v)
}

// LayerContent represents the contents of a layer.toml file.
type LayerContent struct {
	Launch   bool           `toml:"launch"`
	Build    bool           `toml:"build"`
	Cache    bool           `toml:"cache"`
	Metadata map[string]any `toml:"metadata"`
}
