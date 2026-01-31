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

package bptest

import (
	internaltoml "github.com/dmikusa/bptest/internal/toml"
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

	return internaltoml.Encode(data)
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
