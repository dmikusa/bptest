/*
 * Copyright 2026 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package toml provides TOML helpers for the bptest package.
package toml

import (
	"bytes"
	"os"

	"github.com/BurntSushi/toml"
)

// Encode encodes a value to TOML format and returns the bytes.
func Encode(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

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
