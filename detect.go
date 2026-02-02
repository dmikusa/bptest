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
)

// DetectTest configures and executes buildpack detect tests.
type DetectTest struct {
	appFiles          map[string][]byte
	buildpack         BuildpackConfig
	buildpackTOMLPath string // If set, copy this file instead of generating buildpack.toml
	buildpackFiles    map[string][]byte
	buildpackSymlinks map[string]string
	configurations    []ConfigurationMetadata
	dependencies      []DependencyMetadata
	platformEnv       map[string]string
	env               map[string]string
	bindings          []Binding
	stackID           string
	target            *TargetInfo
	dirManager        *testdir.Manager
	paths             TestPaths
}

// NewDetectTest creates a new DetectTest with default values.
func NewDetectTest() *DetectTest {
	return &DetectTest{
		appFiles:          make(map[string][]byte),
		buildpackFiles:    make(map[string][]byte),
		buildpackSymlinks: make(map[string]string),
		platformEnv:       make(map[string]string),
		env:               make(map[string]string),
		buildpack: BuildpackConfig{
			API:     "0.8",
			ID:      "test/buildpack",
			Version: "0.0.1",
		},
		dirManager: testdir.NewManager(),
	}
}

// NewDetectTestFromBuildpack creates a DetectTest using an actual buildpack.toml file.
// The buildpackTOMLPath should be the path to the buildpack.toml file.
// This is useful for integration-style tests where you want to test with the real
// buildpack configuration rather than synthetic test data.
//
// The buildpack.toml is copied to a temporary directory, so WithBuildpackFile,
// WithBuildpackSymlink, and WithBuildpackDependencyCache still work as expected.
//
// Note: WithBuildpack, WithConfigurationMetadata, and WithDependencyMetadata
// have no effect when using this constructor since the real buildpack.toml is used.
func NewDetectTestFromBuildpack(buildpackTOMLPath string) *DetectTest {
	return &DetectTest{
		appFiles:          make(map[string][]byte),
		buildpackTOMLPath: buildpackTOMLPath,
		buildpackFiles:    make(map[string][]byte),
		buildpackSymlinks: make(map[string]string),
		platformEnv:       make(map[string]string),
		env:               make(map[string]string),
		dirManager:        testdir.NewManager(),
	}
}

// WithAppFile adds a file to the application directory.
func (d *DetectTest) WithAppFile(path string, content []byte) *DetectTest {
	d.appFiles[path] = content
	return d
}

// WithAppFileString adds a file to the application directory with string content.
func (d *DetectTest) WithAppFileString(path, content string) *DetectTest {
	return d.WithAppFile(path, []byte(content))
}

// WithBuildpackFile adds a file to the buildpack directory.
func (d *DetectTest) WithBuildpackFile(path string, content []byte) *DetectTest {
	d.buildpackFiles[path] = content
	return d
}

// WithBuildpackFileString adds a file to the buildpack directory with string content.
func (d *DetectTest) WithBuildpackFileString(path, content string) *DetectTest {
	return d.WithBuildpackFile(path, []byte(content))
}

// WithBuildpackSymlink creates a symlink in the buildpack directory.
// The relativePath is relative to the buildpack directory, and targetPath
// is the absolute path that the symlink will point to.
// This is useful for setting up dependency caches that point to test fixtures.
func (d *DetectTest) WithBuildpackSymlink(relativePath, targetPath string) *DetectTest {
	d.buildpackSymlinks[relativePath] = targetPath
	return d
}

// WithBuildpackDependencyCache creates a symlink from the buildpack's "dependencies"
// directory to the specified target path. This is a convenience method for the common
// case of pointing the dependency cache to a test fixtures directory.
func (d *DetectTest) WithBuildpackDependencyCache(targetPath string) *DetectTest {
	return d.WithBuildpackSymlink("dependencies", targetPath)
}

// WithBuildpack sets the buildpack configuration.
func (d *DetectTest) WithBuildpack(config BuildpackConfig) *DetectTest {
	d.buildpack = config
	return d
}

// WithPlatformEnv sets a platform environment variable.
func (d *DetectTest) WithPlatformEnv(key, value string) *DetectTest {
	d.platformEnv[key] = value
	return d
}

// WithBinding adds a service binding.
func (d *DetectTest) WithBinding(binding Binding) *DetectTest {
	d.bindings = append(d.bindings, binding)
	return d
}

// WithStackID sets the stack ID.
func (d *DetectTest) WithStackID(stackID string) *DetectTest {
	d.stackID = stackID
	return d
}

// WithTarget sets the target platform information.
func (d *DetectTest) WithTarget(target TargetInfo) *DetectTest {
	d.target = &target
	return d
}

// WithConfigurationMetadata adds a configuration entry to the buildpack metadata.
func (d *DetectTest) WithConfigurationMetadata(config ConfigurationMetadata) *DetectTest {
	d.configurations = append(d.configurations, config)
	return d
}

// WithDependencyMetadata adds a dependency entry to the buildpack metadata.
func (d *DetectTest) WithDependencyMetadata(dep DependencyMetadata) *DetectTest {
	d.dependencies = append(d.dependencies, dep)
	return d
}

// WithEnv sets an environment variable for the detect function.
// This simulates environment variables that might be set by previous buildpacks.
func (d *DetectTest) WithEnv(key, value string) *DetectTest {
	d.env[key] = value
	return d
}

// Paths returns the test directory paths after Execute has been called.
func (d *DetectTest) Paths() TestPaths {
	return d.paths
}

// Execute runs the detect function and returns the result.
func (d *DetectTest) Execute(detect libcnb.DetectFunc) (*DetectResult, error) {
	if err := d.setup(); err != nil {
		return nil, err
	}

	// Set environment variables and capture originals for restoration
	origEnv := d.setEnv()
	defer d.restoreEnv(origEnv)

	ctx := d.buildContext()
	result, err := detect(ctx)

	return &DetectResult{
		result: result,
		err:    err,
	}, nil
}

// setEnv sets environment variables and returns original values for restoration.
func (d *DetectTest) setEnv() map[string]*string {
	origEnv := make(map[string]*string)
	for key, value := range d.env {
		if orig, exists := os.LookupEnv(key); exists {
			origEnv[key] = &orig
		} else {
			origEnv[key] = nil
		}
		os.Setenv(key, value)
	}
	return origEnv
}

// restoreEnv restores environment variables to their original values.
func (d *DetectTest) restoreEnv(origEnv map[string]*string) {
	for key, value := range origEnv {
		if value == nil {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, *value)
		}
	}
}

// ExecuteT runs the detect function with automatic cleanup on test completion.
func (d *DetectTest) ExecuteT(t *testing.T, detect libcnb.DetectFunc) *DetectResult {
	t.Helper()
	t.Cleanup(func() { _ = d.Cleanup() })

	result, err := d.Execute(detect)
	if err != nil {
		t.Fatalf("failed to execute detect: %v", err)
	}
	return result
}

// Cleanup removes all temporary directories created during the test.
func (d *DetectTest) Cleanup() error {
	return d.dirManager.Cleanup()
}

func (d *DetectTest) setup() error {
	// Create application directory
	appDir, err := d.dirManager.CreateTempDir("bptest-app-")
	if err != nil {
		return err
	}
	d.paths.AppDir = appDir

	// Write application files
	for path, content := range d.appFiles {
		if err := d.dirManager.WriteFile(appDir, path, content); err != nil {
			return err
		}
	}

	// Create buildpack directory
	bpDir, err := d.dirManager.CreateTempDir("bptest-bp-")
	if err != nil {
		return err
	}
	d.paths.BuildpackDir = bpDir

	// Write buildpack.toml - either copy from file or generate from config
	if d.buildpackTOMLPath != "" {
		// Copy the existing buildpack.toml file
		content, err := os.ReadFile(d.buildpackTOMLPath)
		if err != nil {
			return err
		}
		if err := d.dirManager.WriteFile(bpDir, "buildpack.toml", content); err != nil {
			return err
		}
	} else {
		// Merge configurations and dependencies into buildpack metadata
		if len(d.configurations) > 0 || len(d.dependencies) > 0 {
			if d.buildpack.Metadata == nil {
				d.buildpack.Metadata = make(map[string]any)
			}
			if len(d.configurations) > 0 {
				var configs []map[string]any
				for _, c := range d.configurations {
					configs = append(configs, c.toMap())
				}
				d.buildpack.Metadata["configurations"] = configs
			}
			if len(d.dependencies) > 0 {
				var deps []map[string]any
				for _, dep := range d.dependencies {
					deps = append(deps, dep.toMap())
				}
				d.buildpack.Metadata["dependencies"] = deps
			}
		}

		// Write buildpack.toml
		bpTOML, err := d.buildpack.ToTOML()
		if err != nil {
			return err
		}
		if err := d.dirManager.WriteFile(bpDir, "buildpack.toml", bpTOML); err != nil {
			return err
		}
	}

	// Write additional buildpack files
	for path, content := range d.buildpackFiles {
		if err := d.dirManager.WriteFile(bpDir, path, content); err != nil {
			return err
		}
	}

	// Create buildpack symlinks
	for relativePath, targetPath := range d.buildpackSymlinks {
		symlinkPath := filepath.Join(bpDir, relativePath)
		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(symlinkPath), 0755); err != nil {
			return err
		}
		if err := os.Symlink(targetPath, symlinkPath); err != nil {
			return err
		}
	}

	// Create platform directory
	platformDir, err := d.dirManager.CreateTempDir("bptest-platform-")
	if err != nil {
		return err
	}
	d.paths.PlatformDir = platformDir

	// Write platform environment variables
	if len(d.platformEnv) > 0 {
		envDir := filepath.Join(platformDir, "env")
		if err := os.MkdirAll(envDir, 0755); err != nil {
			return err
		}
		for key, value := range d.platformEnv {
			if err := d.dirManager.WriteFileString(platformDir, filepath.Join("env", key), value); err != nil {
				return err
			}
		}
	}

	// Write bindings
	if len(d.bindings) > 0 {
		bindingsDir := filepath.Join(platformDir, "bindings")
		if err := os.MkdirAll(bindingsDir, 0755); err != nil {
			return err
		}
		for _, binding := range d.bindings {
			bindingDir := filepath.Join("bindings", binding.Name)
			// Write type file
			if err := d.dirManager.WriteFileString(platformDir, filepath.Join(bindingDir, "type"), binding.Type); err != nil {
				return err
			}
			// Write provider file if set
			if binding.Provider != "" {
				if err := d.dirManager.WriteFileString(platformDir, filepath.Join(bindingDir, "provider"), binding.Provider); err != nil {
					return err
				}
			}
			// Write secret files
			for key, value := range binding.Secret {
				if err := d.dirManager.WriteFileString(platformDir, filepath.Join(bindingDir, key), value); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (d *DetectTest) buildContext() libcnb.DetectContext {
	// Build platform bindings from our binding configs
	var bindings libcnb.Bindings
	for _, b := range d.bindings {
		bindingPath := filepath.Join(d.paths.PlatformDir, "bindings", b.Name)
		bindings = append(bindings, libcnb.Binding{
			Name:     b.Name,
			Path:     bindingPath,
			Type:     b.Type,
			Provider: b.Provider,
			Secret:   b.Secret,
		})
	}

	// Read buildpack.toml to get full Buildpack struct
	var buildpack libcnb.Buildpack
	bpPath := filepath.Join(d.paths.BuildpackDir, "buildpack.toml")
	if data, err := os.ReadFile(bpPath); err == nil {
		_ = toml.Unmarshal(data, &buildpack)
	}
	buildpack.Path = d.paths.BuildpackDir

	return libcnb.DetectContext{
		ApplicationPath: d.paths.AppDir,
		Buildpack:       buildpack,
		Logger:          log.NewDiscard(),
		Platform: libcnb.Platform{
			Path:        d.paths.PlatformDir,
			Bindings:    bindings,
			Environment: d.platformEnv,
		},
		StackID: d.stackID,
	}
}
