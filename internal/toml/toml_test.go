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

package toml_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dmikusa/bptest/internal/toml"
)

func TestRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")

	content := `
name = "test"
version = "1.0.0"

[metadata]
key = "value"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	type Config struct {
		Name     string            `toml:"name"`
		Version  string            `toml:"version"`
		Metadata map[string]string `toml:"metadata"`
	}

	result, err := toml.Read[Config](path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if result.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", result.Name)
	}
	if result.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", result.Version)
	}
	if result.Metadata["key"] != "value" {
		t.Errorf("expected metadata key 'value', got '%s'", result.Metadata["key"])
	}
}

func TestRead_FileNotFound(t *testing.T) {
	_, err := toml.Read[struct{}]("/nonexistent/path/file.toml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestRead_InvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.toml")

	if err := os.WriteFile(path, []byte("this is not valid toml [[["), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := toml.Read[struct{}](path)
	if err == nil {
		t.Error("expected error for invalid TOML")
	}
}

func TestWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output.toml")

	type Config struct {
		Name    string `toml:"name"`
		Version string `toml:"version"`
	}

	config := Config{
		Name:    "myapp",
		Version: "2.0.0",
	}

	err := toml.Write(path, config)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read back and verify
	result, err := toml.Read[Config](path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if result.Name != "myapp" {
		t.Errorf("expected name 'myapp', got '%s'", result.Name)
	}
	if result.Version != "2.0.0" {
		t.Errorf("expected version '2.0.0', got '%s'", result.Version)
	}
}

func TestWrite_InvalidPath(t *testing.T) {
	err := toml.Write("/nonexistent/directory/file.toml", struct{}{})
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestEncode(t *testing.T) {
	type Config struct {
		Name    string `toml:"name"`
		Version string `toml:"version"`
	}

	config := Config{
		Name:    "myapp",
		Version: "1.0.0",
	}

	data, err := toml.Encode(config)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Verify it's valid TOML by reading it back
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	result, err := toml.Read[Config](path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if result.Name != "myapp" {
		t.Errorf("expected name 'myapp', got '%s'", result.Name)
	}
	if result.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", result.Version)
	}
}

func TestLayerContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "layer.toml")

	layer := toml.LayerContent{
		Launch: true,
		Build:  true,
		Cache:  false,
		Metadata: map[string]any{
			"version": "17",
			"vendor":  "adoptium",
		},
	}

	err := toml.Write(path, layer)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	result, err := toml.Read[toml.LayerContent](path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if !result.Launch {
		t.Error("expected Launch to be true")
	}
	if !result.Build {
		t.Error("expected Build to be true")
	}
	if result.Cache {
		t.Error("expected Cache to be false")
	}
	if result.Metadata["version"] != "17" {
		t.Errorf("expected version '17', got '%v'", result.Metadata["version"])
	}
	if result.Metadata["vendor"] != "adoptium" {
		t.Errorf("expected vendor 'adoptium', got '%v'", result.Metadata["vendor"])
	}
}
