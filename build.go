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
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/buildpacks/libcnb/v2"
	"github.com/buildpacks/libcnb/v2/log"

	"github.com/dmikusa/bptest/internal/testdir"
	internaltoml "github.com/dmikusa/bptest/internal/toml"
)

// BuildTest configures and executes buildpack build tests.
type BuildTest struct {
	appFiles           map[string][]byte
	buildpack          BuildpackConfig
	configurations     []ConfigurationMetadata
	dependencies       []DependencyMetadata
	platformEnv        map[string]string
	bindings           []Binding
	planEntries        []libcnb.BuildpackPlanEntry
	existingLayers     map[string]*LayerConfig
	persistentMetadata map[string]any
	stackID            string
	target             *TargetInfo
	dirManager         *testdir.Manager
	paths              TestPaths
}

// NewBuildTest creates a new BuildTest with default values.
func NewBuildTest() *BuildTest {
	return &BuildTest{
		appFiles:       make(map[string][]byte),
		platformEnv:    make(map[string]string),
		existingLayers: make(map[string]*LayerConfig),
		buildpack: BuildpackConfig{
			API:     "0.8",
			ID:      "test/buildpack",
			Version: "0.0.1",
		},
		dirManager: testdir.NewManager(),
	}
}

// WithAppFile adds a file to the application directory.
func (b *BuildTest) WithAppFile(path string, content []byte) *BuildTest {
	b.appFiles[path] = content
	return b
}

// WithAppFileString adds a file to the application directory with string content.
func (b *BuildTest) WithAppFileString(path, content string) *BuildTest {
	return b.WithAppFile(path, []byte(content))
}

// WithBuildpack sets the buildpack configuration.
func (b *BuildTest) WithBuildpack(config BuildpackConfig) *BuildTest {
	b.buildpack = config
	return b
}

// WithPlatformEnv sets a platform environment variable.
func (b *BuildTest) WithPlatformEnv(key, value string) *BuildTest {
	b.platformEnv[key] = value
	return b
}

// WithBinding adds a service binding.
func (b *BuildTest) WithBinding(binding Binding) *BuildTest {
	b.bindings = append(b.bindings, binding)
	return b
}

// WithPlanEntry adds a buildpack plan entry.
func (b *BuildTest) WithPlanEntry(name string, metadata map[string]any) *BuildTest {
	b.planEntries = append(b.planEntries, libcnb.BuildpackPlanEntry{
		Name:     name,
		Metadata: metadata,
	})
	return b
}

// WithPlan sets the entire buildpack plan.
func (b *BuildTest) WithPlan(plan libcnb.BuildpackPlan) *BuildTest {
	b.planEntries = plan.Entries
	return b
}

// WithExistingLayer configures an existing layer for rebuild scenarios.
func (b *BuildTest) WithExistingLayer(name string, config *LayerConfig) *BuildTest {
	b.existingLayers[name] = config
	return b
}

// WithPersistentMetadata sets the persistent metadata (store.toml).
func (b *BuildTest) WithPersistentMetadata(metadata map[string]any) *BuildTest {
	b.persistentMetadata = metadata
	return b
}

// WithStackID sets the stack ID.
func (b *BuildTest) WithStackID(stackID string) *BuildTest {
	b.stackID = stackID
	return b
}

// WithTarget sets the target platform information.
func (b *BuildTest) WithTarget(target TargetInfo) *BuildTest {
	b.target = &target
	return b
}

// WithConfigurationMetadata adds a configuration entry to the buildpack metadata.
func (b *BuildTest) WithConfigurationMetadata(config ConfigurationMetadata) *BuildTest {
	b.configurations = append(b.configurations, config)
	return b
}

// WithDependencyMetadata adds a dependency entry to the buildpack metadata.
func (b *BuildTest) WithDependencyMetadata(dep DependencyMetadata) *BuildTest {
	b.dependencies = append(b.dependencies, dep)
	return b
}

// Paths returns the test directory paths after Execute has been called.
func (b *BuildTest) Paths() TestPaths {
	return b.paths
}

// Execute runs the build function and returns the result.
func (b *BuildTest) Execute(build libcnb.BuildFunc) (*BuildResult, error) {
	if err := b.setup(); err != nil {
		return nil, err
	}

	ctx := b.buildContext()
	result, err := build(ctx)

	return &BuildResult{
		result:    result,
		err:       err,
		layersDir: b.paths.LayersDir,
	}, nil
}

// ExecuteT runs the build function with automatic cleanup on test completion.
func (b *BuildTest) ExecuteT(t *testing.T, build libcnb.BuildFunc) *BuildResult {
	t.Helper()
	t.Cleanup(func() { _ = b.Cleanup() })

	result, err := b.Execute(build)
	if err != nil {
		t.Fatalf("failed to execute build: %v", err)
	}
	return result
}

// Cleanup removes all temporary directories created during the test.
func (b *BuildTest) Cleanup() error {
	return b.dirManager.Cleanup()
}

func (b *BuildTest) setup() error {
	// Create application directory
	appDir, err := b.dirManager.CreateTempDir("bptest-app-")
	if err != nil {
		return err
	}
	b.paths.AppDir = appDir

	// Write application files
	for path, content := range b.appFiles {
		if err := b.dirManager.WriteFile(appDir, path, content); err != nil {
			return err
		}
	}

	// Create buildpack directory
	bpDir, err := b.dirManager.CreateTempDir("bptest-bp-")
	if err != nil {
		return err
	}
	b.paths.BuildpackDir = bpDir

	// Merge configurations and dependencies into buildpack metadata
	if len(b.configurations) > 0 || len(b.dependencies) > 0 {
		if b.buildpack.Metadata == nil {
			b.buildpack.Metadata = make(map[string]any)
		}
		if len(b.configurations) > 0 {
			var configs []map[string]any
			for _, c := range b.configurations {
				configs = append(configs, c.toMap())
			}
			b.buildpack.Metadata["configurations"] = configs
		}
		if len(b.dependencies) > 0 {
			var deps []map[string]any
			for _, d := range b.dependencies {
				deps = append(deps, d.toMap())
			}
			b.buildpack.Metadata["dependencies"] = deps
		}
	}

	// Write buildpack.toml
	bpTOML, err := b.buildpack.ToTOML()
	if err != nil {
		return err
	}
	if err := b.dirManager.WriteFile(bpDir, "buildpack.toml", bpTOML); err != nil {
		return err
	}

	// Create platform directory
	platformDir, err := b.dirManager.CreateTempDir("bptest-platform-")
	if err != nil {
		return err
	}
	b.paths.PlatformDir = platformDir

	// Write platform environment variables
	if len(b.platformEnv) > 0 {
		envDir := filepath.Join(platformDir, "env")
		if err := os.MkdirAll(envDir, 0755); err != nil {
			return err
		}
		for key, value := range b.platformEnv {
			if err := b.dirManager.WriteFileString(platformDir, filepath.Join("env", key), value); err != nil {
				return err
			}
		}
	}

	// Write bindings
	if len(b.bindings) > 0 {
		bindingsDir := filepath.Join(platformDir, "bindings")
		if err := os.MkdirAll(bindingsDir, 0755); err != nil {
			return err
		}
		for _, binding := range b.bindings {
			bindingDir := filepath.Join("bindings", binding.Name)
			if err := b.dirManager.WriteFileString(platformDir, filepath.Join(bindingDir, "type"), binding.Type); err != nil {
				return err
			}
			if binding.Provider != "" {
				if err := b.dirManager.WriteFileString(platformDir, filepath.Join(bindingDir, "provider"), binding.Provider); err != nil {
					return err
				}
			}
			for key, value := range binding.Secret {
				if err := b.dirManager.WriteFileString(platformDir, filepath.Join(bindingDir, key), value); err != nil {
					return err
				}
			}
		}
	}

	// Create layers directory
	layersDir, err := b.dirManager.CreateTempDir("bptest-layers-")
	if err != nil {
		return err
	}
	b.paths.LayersDir = layersDir

	// Set up existing layers for rebuild scenarios
	for name, config := range b.existingLayers {
		layerDir := filepath.Join(layersDir, name)
		if err := os.MkdirAll(layerDir, 0755); err != nil {
			return err
		}

		// Write layer.toml
		layerContent := internaltoml.LayerContent{
			Build:    config.Build,
			Launch:   config.Launch,
			Cache:    config.Cache,
			Metadata: config.Metadata,
		}
		if err := internaltoml.Write(filepath.Join(layersDir, name+".toml"), layerContent); err != nil {
			return err
		}

		// Write layer files
		for path, content := range config.Files {
			if err := b.dirManager.WriteFile(layerDir, path, content); err != nil {
				return err
			}
		}

		// Write env vars
		if len(config.Env) > 0 {
			envDir := filepath.Join(layerDir, "env")
			if err := os.MkdirAll(envDir, 0755); err != nil {
				return err
			}
			for key, value := range config.Env {
				if err := b.dirManager.WriteFileString(layerDir, filepath.Join("env", key), value); err != nil {
					return err
				}
			}
		}
	}

	// Write store.toml if persistent metadata provided
	if len(b.persistentMetadata) > 0 {
		store := struct {
			Metadata map[string]any `toml:"metadata"`
		}{Metadata: b.persistentMetadata}
		if err := internaltoml.Write(filepath.Join(layersDir, "store.toml"), store); err != nil {
			return err
		}
	}

	return nil
}

func (b *BuildTest) buildContext() libcnb.BuildContext {
	// Build platform bindings
	var bindings libcnb.Bindings
	for _, binding := range b.bindings {
		bindingPath := filepath.Join(b.paths.PlatformDir, "bindings", binding.Name)
		bindings = append(bindings, libcnb.Binding{
			Name:     binding.Name,
			Path:     bindingPath,
			Type:     binding.Type,
			Provider: binding.Provider,
			Secret:   binding.Secret,
		})
	}

	// Read buildpack.toml
	var buildpack libcnb.Buildpack
	bpPath := filepath.Join(b.paths.BuildpackDir, "buildpack.toml")
	if data, err := os.ReadFile(bpPath); err == nil {
		_ = toml.Unmarshal(data, &buildpack)
	}
	buildpack.Path = b.paths.BuildpackDir

	ctx := libcnb.BuildContext{
		ApplicationPath: b.paths.AppDir,
		Buildpack:       buildpack,
		Layers:          libcnb.Layers{Path: b.paths.LayersDir},
		Logger:          log.NewDiscard(),
		PersistentMetadata: b.persistentMetadata,
		Plan: libcnb.BuildpackPlan{
			Entries: b.planEntries,
		},
		Platform: libcnb.Platform{
			Path:        b.paths.PlatformDir,
			Bindings:    bindings,
			Environment: b.platformEnv,
		},
		StackID: b.stackID,
	}

	// Set target info if provided
	if b.target != nil {
		ctx.TargetInfo = libcnb.TargetInfo{
			OS:   b.target.OS,
			Arch: b.target.Arch,
		}
		ctx.TargetDistro = libcnb.TargetDistro{
			Name:    b.target.DistroName,
			Version: b.target.DistroVersion,
		}
	}

	return ctx
}
