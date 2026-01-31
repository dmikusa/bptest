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
	appFiles    map[string][]byte
	buildpack   BuildpackConfig
	platformEnv map[string]string
	bindings    []Binding
	stackID     string
	target      *TargetInfo
	dirManager  *testdir.Manager
	paths       TestPaths
}

// NewDetectTest creates a new DetectTest with default values.
func NewDetectTest() *DetectTest {
	return &DetectTest{
		appFiles:    make(map[string][]byte),
		platformEnv: make(map[string]string),
		buildpack: BuildpackConfig{
			API:     "0.8",
			ID:      "test/buildpack",
			Version: "0.0.1",
		},
		dirManager: testdir.NewManager(),
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

// Paths returns the test directory paths after Execute has been called.
func (d *DetectTest) Paths() TestPaths {
	return d.paths
}

// Execute runs the detect function and returns the result.
func (d *DetectTest) Execute(detect libcnb.DetectFunc) (*DetectResult, error) {
	if err := d.setup(); err != nil {
		return nil, err
	}

	ctx := d.buildContext()
	result, err := detect(ctx)

	return &DetectResult{
		result: result,
		err:    err,
	}, nil
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

	// Write buildpack.toml
	bpTOML, err := d.buildpack.ToTOML()
	if err != nil {
		return err
	}
	if err := d.dirManager.WriteFile(bpDir, "buildpack.toml", bpTOML); err != nil {
		return err
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
