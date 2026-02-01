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
	API      string
	ID       string
	Version  string
	Name     string
	Metadata map[string]any
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
	if len(c.Metadata) > 0 {
		data["metadata"] = c.Metadata
	}

	return internaltoml.Encode(data)
}

// ConfigurationMetadata represents a buildpack configuration entry.
type ConfigurationMetadata struct {
	Name        string
	Description string
	Default     string
	Launch      bool
	Build       bool
}

// toMap converts ConfigurationMetadata to a map for TOML encoding.
func (c ConfigurationMetadata) toMap() map[string]any {
	m := map[string]any{
		"name":        c.Name,
		"description": c.Description,
		"default":     c.Default,
	}
	if c.Launch {
		m["launch"] = true
	}
	if c.Build {
		m["build"] = true
	}
	return m
}

// DependencyMetadata represents a buildpack dependency entry.
type DependencyMetadata struct {
	ID           string
	Name         string
	Version      string
	URI          string
	SHA256       string
	Stacks       []string
	CPEs         []string
	PURL         string
	Source       string
	SourceSHA256 string
	Licenses     []LicenseMetadata
}

// LicenseMetadata represents a license entry for a dependency.
type LicenseMetadata struct {
	Type string
	URI  string
}

// toMap converts DependencyMetadata to a map for TOML encoding.
func (d DependencyMetadata) toMap() map[string]any {
	m := map[string]any{
		"id":      d.ID,
		"name":    d.Name,
		"version": d.Version,
		"uri":     d.URI,
		"sha256":  d.SHA256,
		"stacks":  d.Stacks,
	}
	if len(d.CPEs) > 0 {
		m["cpes"] = d.CPEs
	}
	if d.PURL != "" {
		m["purl"] = d.PURL
	}
	if d.Source != "" {
		m["source"] = d.Source
	}
	if d.SourceSHA256 != "" {
		m["source-sha256"] = d.SourceSHA256
	}
	if len(d.Licenses) > 0 {
		var licenses []map[string]any
		for _, lic := range d.Licenses {
			licenses = append(licenses, map[string]any{
				"type": lic.Type,
				"uri":  lic.URI,
			})
		}
		m["licenses"] = licenses
	}
	return m
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
