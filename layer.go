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
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb/v2"

	internaltoml "github.com/dmikusa/bptest/internal/toml"
)

// LayerInspector provides methods for examining contributed layers.
type LayerInspector struct {
	layers    []libcnb.Layer
	layersDir string
}

// Names returns the names of all contributed layers.
func (i *LayerInspector) Names() []string {
	var names []string
	for _, layer := range i.layers {
		names = append(names, layer.Name)
	}
	return names
}

// Get returns a LayerResult for the layer with the given name.
// Returns nil if the layer doesn't exist.
func (i *LayerInspector) Get(name string) *LayerResult {
	for _, layer := range i.layers {
		if layer.Name == name {
			return &LayerResult{layer: layer}
		}
	}
	return nil
}

// Count returns the number of contributed layers.
func (i *LayerInspector) Count() int {
	return len(i.layers)
}

// LayerResult provides methods for inspecting a single layer.
type LayerResult struct {
	layer libcnb.Layer
}

// Name returns the layer name.
func (r *LayerResult) Name() string {
	return r.layer.Name
}

// Path returns the layer's filesystem path.
func (r *LayerResult) Path() string {
	return r.layer.Path
}

// IsBuild returns true if the layer is marked for build.
func (r *LayerResult) IsBuild() bool {
	return r.layer.Build
}

// IsLaunch returns true if the layer is marked for launch.
func (r *LayerResult) IsLaunch() bool {
	return r.layer.Launch
}

// IsCache returns true if the layer is marked for caching.
func (r *LayerResult) IsCache() bool {
	return r.layer.Cache
}

// Metadata returns the layer's metadata.
func (r *LayerResult) Metadata() map[string]any {
	return r.layer.Metadata
}

// MetadataValue returns the value for a specific metadata key.
func (r *LayerResult) MetadataValue(key string) any {
	if r.layer.Metadata == nil {
		return nil
	}
	return r.layer.Metadata[key]
}

// FileExists returns true if the file exists in the layer.
func (r *LayerResult) FileExists(relativePath string) bool {
	fullPath := filepath.Join(r.layer.Path, relativePath)
	_, err := os.Stat(fullPath)
	return err == nil
}

// FileContent returns the content of a file in the layer.
func (r *LayerResult) FileContent(relativePath string) ([]byte, error) {
	fullPath := filepath.Join(r.layer.Path, relativePath)
	return os.ReadFile(fullPath)
}

// FileContentString returns the content of a file as a string.
func (r *LayerResult) FileContentString(relativePath string) (string, error) {
	content, err := r.FileContent(relativePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// FileMode returns the file mode (permissions) of a file in the layer.
func (r *LayerResult) FileMode(relativePath string) (os.FileMode, error) {
	fullPath := filepath.Join(r.layer.Path, relativePath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, err
	}
	return info.Mode(), nil
}

// BinFiles returns the list of files in the layer's bin directory.
func (r *LayerResult) BinFiles() ([]string, error) {
	return r.listDir("bin")
}

// LibFiles returns the list of files in the layer's lib directory.
func (r *LayerResult) LibFiles() ([]string, error) {
	return r.listDir("lib")
}

// EnvBuild returns the build environment variables set by the layer.
func (r *LayerResult) EnvBuild() map[string]string {
	return r.readEnvDir("env.build")
}

// EnvLaunch returns the launch environment variables set by the layer.
func (r *LayerResult) EnvLaunch() map[string]string {
	return r.readEnvDir("env.launch")
}

// EnvShared returns the shared environment variables set by the layer.
func (r *LayerResult) EnvShared() map[string]string {
	return r.readEnvDir("env")
}

// ExecDFiles returns the list of exec.d files in the layer.
func (r *LayerResult) ExecDFiles() ([]string, error) {
	return r.listDir("exec.d")
}

func (r *LayerResult) listDir(subdir string) ([]string, error) {
	dir := filepath.Join(r.layer.Path, subdir)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names, nil
}

func (r *LayerResult) readEnvDir(subdir string) map[string]string {
	dir := filepath.Join(r.layer.Path, subdir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	env := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err == nil {
			env[entry.Name()] = string(content)
		}
	}
	return env
}

// ReadLayerTOML reads and returns the layer.toml content for a layer in the layers directory.
func ReadLayerTOML(layersDir, layerName string) (*internaltoml.LayerContent, error) {
	path := filepath.Join(layersDir, layerName+".toml")
	content, err := internaltoml.Read[internaltoml.LayerContent](path)
	if err != nil {
		return nil, err
	}
	return &content, nil
}
