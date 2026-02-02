# bptest

A Go testing library for [Cloud Native Buildpacks](https://buildpacks.io/) that allows you to test buildpack `detect` and `build` functions directly by invoking them with configurable test data.

## Why bptest?

Testing buildpacks traditionally requires running the full lifecycle or spawning subprocesses. `bptest` takes a different approach—it lets you call your `detect` and `build` functions directly in unit tests with complete control over the inputs (application files, platform environment, service bindings, etc.) and easy inspection of the outputs (build plans, layers, processes, labels).

## Installation

```bash
go get github.com/dmikusa/bptest
```

## Getting Started

Here's a simple example testing a detect function that looks for a `pom.xml` file:

```go
package mybuildpack_test

import (
    "testing"

    "github.com/buildpacks/libcnb/v2"
    "github.com/dmikusa/bptest"
)

func TestDetect_WithPomXML(t *testing.T) {
    result := bptest.NewDetectTest().
        WithAppFileString("pom.xml", `<project><modelVersion>4.0.0</modelVersion></project>`).
        ExecuteT(t, myDetector.Detect)

    if !result.Passed() {
        t.Error("expected detect to pass when pom.xml exists")
    }
    if !result.HasPlan("jdk") {
        t.Error("expected to require jdk")
    }
}
```

## Examples

### Detect Tests

#### Basic Detection

```go
func TestDetect_Pass(t *testing.T) {
    result := bptest.NewDetectTest().
        WithAppFileString("package.json", `{"name": "myapp"}`).
        ExecuteT(t, myDetector.Detect)

    if !result.Passed() {
        t.Fatal("expected detect to pass")
    }
}
```

#### Platform Environment Variables

```go
func TestDetect_WithEnvVar(t *testing.T) {
    result := bptest.NewDetectTest().
        WithAppFileString("pom.xml", "<project/>").
        WithPlatformEnv("BP_JAVA_VERSION", "17").
        ExecuteT(t, myDetector.Detect)

    if !result.Passed() {
        t.Fatal("expected detect to pass")
    }
}
```

#### Service Bindings

```go
func TestDetect_WithBinding(t *testing.T) {
    result := bptest.NewDetectTest().
        WithBinding(bptest.Binding{
            Name:   "my-database",
            Type:   "postgresql",
            Secret: map[string]string{
                "host":     "localhost",
                "username": "admin",
                "password": "secret",
            },
        }).
        ExecuteT(t, myDetector.Detect)

    if !result.Passed() {
        t.Fatal("expected detect to pass with database binding")
    }
}
```

#### Inspecting Build Plans

```go
func TestDetect_BuildPlan(t *testing.T) {
    result := bptest.NewDetectTest().
        WithAppFileString("pom.xml", "<project/>").
        ExecuteT(t, myDetector.Detect)

    // Check specific requirements
    if !result.HasPlan("jdk") {
        t.Error("expected jdk in build plan")
    }

    // Get all provides and requires
    provides := result.Provides()  // []string{"jdk", "jre"}
    requires := result.Requires()  // []string{"jdk"}

    // Get metadata for a requirement
    metadata := result.RequireMetadata("jdk")
    if metadata["version"] != "17" {
        t.Errorf("expected version 17, got %v", metadata["version"])
    }
}
```

### Build Tests

#### Basic Build

```go
func TestBuild_Success(t *testing.T) {
    result := bptest.NewBuildTest().
        WithAppFileString("pom.xml", "<project/>").
        WithPlanEntry("jdk", map[string]any{"version": "17"}).
        ExecuteT(t, myBuilder.Build)

    if !result.Succeeded() {
        t.Fatalf("build failed: %v", result.Error())
    }
}
```

#### Inspecting Layers

```go
func TestBuild_LayerCreation(t *testing.T) {
    result := bptest.NewBuildTest().
        WithPlanEntry("jdk", map[string]any{"version": "17"}).
        ExecuteT(t, myBuilder.Build)

    if !result.Succeeded() {
        t.Fatalf("build failed: %v", result.Error())
    }

    // Get a specific layer
    layer := result.Layer("jdk")
    if layer == nil {
        t.Fatal("expected jdk layer")
    }

    // Check layer types
    if !layer.IsBuild() {
        t.Error("expected build layer")
    }
    if !layer.IsCache() {
        t.Error("expected cached layer")
    }

    // Check layer metadata
    if layer.MetadataValue("version") != "17" {
        t.Error("expected version metadata")
    }

    // Check files exist
    if !layer.FileExists("bin/java") {
        t.Error("expected java binary")
    }

    // Read file contents
    content, err := layer.FileContentString("release")
    if err != nil {
        t.Errorf("failed to read release file: %v", err)
    }

    // Check environment variables
    envBuild := layer.EnvBuild()
    if envBuild["JAVA_HOME.override"] == "" {
        t.Error("expected JAVA_HOME to be set")
    }
}
```

#### Inspecting Processes

```go
func TestBuild_Processes(t *testing.T) {
    result := bptest.NewBuildTest().
        WithAppFileString("app.jar", "").
        WithPlanEntry("jdk", nil).
        ExecuteT(t, myBuilder.Build)

    if !result.Succeeded() {
        t.Fatalf("build failed: %v", result.Error())
    }

    // Check for specific process
    if !result.HasProcess("web") {
        t.Error("expected web process")
    }

    // Get process details
    web := result.Process("web")
    if web == nil || len(web.Command) == 0 {
        t.Error("expected web process with command")
    }

    // Check default process
    defaultProc := result.DefaultProcess()
    if defaultProc == nil || defaultProc.Type != "web" {
        t.Error("expected web to be the default process")
    }
}
```

#### Rebuild Scenarios (Existing Layers)

Test how your buildpack handles rebuilds when layers already exist:

```go
func TestBuild_Rebuild(t *testing.T) {
    result := bptest.NewBuildTest().
        WithPlanEntry("jdk", map[string]any{"version": "17"}).
        WithExistingLayer("jdk", &bptest.LayerConfig{
            Build: true,
            Cache: true,
            Metadata: map[string]any{
                "version": "11",  // Previous version
            },
            Files: map[string][]byte{
                "bin/java": []byte("old java"),
            },
        }).
        ExecuteT(t, myBuilder.Build)

    if !result.Succeeded() {
        t.Fatalf("rebuild failed: %v", result.Error())
    }

    // Verify layer was updated
    layer := result.Layer("jdk")
    if layer.MetadataValue("version") != "17" {
        t.Error("expected layer to be updated to version 17")
    }
}
```

#### Persistent Metadata (store.toml)

```go
func TestBuild_PersistentMetadata(t *testing.T) {
    result := bptest.NewBuildTest().
        WithPersistentMetadata(map[string]any{
            "last-build-time": "2024-01-01T00:00:00Z",
        }).
        ExecuteT(t, myBuilder.Build)

    if !result.Succeeded() {
        t.Fatalf("build failed: %v", result.Error())
    }

    // Check updated persistent metadata
    meta := result.PersistentMetadata()
    if meta["last-build-time"] == "" {
        t.Error("expected persistent metadata to be set")
    }
}
```

#### Labels and Slices

```go
func TestBuild_Labels(t *testing.T) {
    result := bptest.NewBuildTest().
        WithPlanEntry("jdk", nil).
        ExecuteT(t, myBuilder.Build)

    if !result.Succeeded() {
        t.Fatalf("build failed: %v", result.Error())
    }

    // Check labels
    if result.Label("org.opencontainers.image.title") == "" {
        t.Error("expected image title label")
    }

    // Check slices
    slices := result.Slices()
    if len(slices) == 0 {
        t.Error("expected at least one slice")
    }
}
```

### Custom Buildpack Configuration

```go
func TestDetect_CustomBuildpack(t *testing.T) {
    result := bptest.NewDetectTest().
        WithBuildpack(bptest.BuildpackConfig{
            API:     "0.9",
            ID:      "example/my-buildpack",
            Version: "1.2.3",
            Name:    "My Buildpack",
        }).
        WithAppFileString("app.txt", "hello").
        ExecuteT(t, myDetector.Detect)

    if !result.Passed() {
        t.Fatal("expected detect to pass")
    }
}
```

### Target Platform Information

```go
func TestBuild_WithTarget(t *testing.T) {
    result := bptest.NewBuildTest().
        WithTarget(bptest.TargetInfo{
            OS:            "linux",
            Arch:          "arm64",
            DistroName:    "ubuntu",
            DistroVersion: "22.04",
        }).
        WithPlanEntry("jdk", nil).
        ExecuteT(t, myBuilder.Build)

    if !result.Succeeded() {
        t.Fatalf("build failed: %v", result.Error())
    }
}
```

### Integration Tests with Real Buildpack Configuration

For integration-style tests that use your actual `buildpack.toml` configuration:

```go
func TestBuild_Integration(t *testing.T) {
    result := bptest.NewBuildTestFromBuildpack("buildpack.toml").
        WithAppFileString("pom.xml", "<project/>").
        WithPlanEntry("jdk", nil).
        ExecuteT(t, myBuilder.Build)

    if !result.Succeeded() {
        t.Fatalf("build failed: %v", result.Error())
    }
}
```

This copies your real `buildpack.toml` to a temporary directory, so you can still use `WithBuildpackFile`, `WithBuildpackSymlink`, and `WithBuildpackDependencyCache` to add test fixtures.

### Dependency Caching with Symlinks

When testing buildpacks that download dependencies, you can use `WithBuildpackDependencyCache` to point the buildpack's dependency cache to a local test fixtures directory. This allows your tests to use mock dependencies instead of downloading from the internet:

```go
func TestBuild_WithDependencyCache(t *testing.T) {
    // Get absolute path to testdata directory containing mock dependencies
    testdataAbs, err := filepath.Abs("testdata")
    if err != nil {
        t.Fatalf("failed to get absolute path: %v", err)
    }

    result := bptest.NewBuildTest().
        WithBuildpackDependencyCache(testdataAbs).
        WithPlanEntry("jdk", nil).
        ExecuteT(t, myBuilder.Build)

    if !result.Succeeded() {
        t.Fatalf("build failed: %v", result.Error())
    }
}
```

You can also add arbitrary files to the buildpack directory:

```go
func TestBuild_WithBuildpackConfig(t *testing.T) {
    result := bptest.NewBuildTest().
        WithBuildpackFileString("config/settings.json", `{"debug": true}`).
        WithPlanEntry("jdk", nil).
        ExecuteT(t, myBuilder.Build)

    if !result.Succeeded() {
        t.Fatalf("build failed: %v", result.Error())
    }
}
```

### Gomega Matchers

The `matchers` subpackage provides custom [Gomega](https://onsi.github.io/gomega/) matchers for more expressive assertions.
Import with dot notation for cleaner test code:

```go
import (
    . "github.com/onsi/gomega"
    "github.com/dmikusa/bptest"
    . "github.com/dmikusa/bptest/matchers"
)

func TestDetect_WithGomega(t *testing.T) {
    g := NewWithT(t)

    result := bptest.NewDetectTest().
        WithAppFileString("pom.xml", "<project/>").
        ExecuteT(t, myDetector.Detect)

    g.Expect(result).To(HavePassed())
    g.Expect(result).To(HavePlan("jdk"))
    g.Expect(result).To(Provide("jdk"))
    g.Expect(result).To(Require("jvm"))
}

func TestBuild_WithGomega(t *testing.T) {
    g := NewWithT(t)

    result := bptest.NewBuildTest().
        WithPlanEntry("jdk", nil).
        ExecuteT(t, myBuilder.Build)

    g.Expect(result).To(HaveSucceeded())
    g.Expect(result).To(HaveLayer("jdk"))
    g.Expect(result).To(HaveProcess("web"))
    g.Expect(result).To(HaveLabel("org.opencontainers.image.title", "My App"))
}
```

#### Available Matchers

**DetectResult matchers:**

| Matcher | Description |
|---------|-------------|
| `HavePassed()` | Check if detection passed |
| `HaveFailed()` | Check if detection failed |
| `HavePlan(name)` | Check if plan provides or requires name |
| `Provide(name)` | Check if provides includes name |
| `Require(name)` | Check if requires includes name |

**BuildResult matchers:**

| Matcher | Description |
|---------|-------------|
| `HaveSucceeded()` | Check if build succeeded |
| `HaveLayer(name)` | Check if layer exists |
| `HaveLayerCount(n)` | Check exact number of layers |
| `HaveLayers(names...)` | Check that specified layers exist (others may exist) |
| `HaveExactlyLayers(names...)` | Check that only specified layers exist |
| `HaveProcess(type)` | Check if process type exists |
| `HaveLabel(key)` | Check if label key exists |
| `HaveLabel(key, value)` | Check if label has specific value |

**LayerResult matchers:**

| Matcher | Description |
|---------|-------------|
| `HaveFile(path)` | Check if file exists in layer |
| `HaveFileWithContents(path, contents)` | Check file exists with exact contents |
| `HaveFileWithPerms(path, perms)` | Check file exists with exact permissions (e.g., `0755`) |
| `BeBuildLayer()` | Check if layer is marked for build |
| `BeLaunchLayer()` | Check if layer is marked for launch |
| `BeCacheLayer()` | Check if layer is marked for caching |
| `HaveMetadata(key)` | Check if layer has metadata key |
| `HaveMetadata(key, value)` | Check if layer metadata has specific value |

#### Layer Matchers Example

```go
func TestBuild_LayerMatchers(t *testing.T) {
    g := NewWithT(t)

    result := bptest.NewBuildTest().
        WithPlanEntry("jdk", nil).
        ExecuteT(t, myBuilder.Build)

    // Check layer counts
    g.Expect(result).To(HaveLayerCount(2))
    g.Expect(result).To(HaveLayers("jdk", "helper"))
    g.Expect(result).To(HaveExactlyLayers("jdk", "helper"))

    // Check layer properties
    layer := result.Layer("jdk")
    g.Expect(layer).To(BeBuildLayer())
    g.Expect(layer).To(BeCacheLayer())
    g.Expect(layer).NotTo(BeLaunchLayer())

    // Check layer files
    g.Expect(layer).To(HaveFile("bin/java"))
    g.Expect(layer).To(HaveFileWithContents("version", "17.0.1"))

    // Check layer metadata
    g.Expect(layer).To(HaveMetadata("version"))
    g.Expect(layer).To(HaveMetadata("version", "17"))
}
```

### Buildpack Metadata

Configure buildpack metadata including configurations and dependencies:

```go
func TestBuild_WithBuildpackMetadata(t *testing.T) {
    result := bptest.NewBuildTest().
        WithConfigurationMetadata(bptest.ConfigurationMetadata{
            Name:        "BP_JVM_VERSION",
            Description: "the Java version to install",
            Default:     "17",
            Build:       true,
        }).
        WithConfigurationMetadata(bptest.ConfigurationMetadata{
            Name:        "BPL_JVM_HEAD_ROOM",
            Description: "the headroom in memory calculation",
            Default:     "0",
            Launch:      true,
        }).
        WithDependencyMetadata(bptest.DependencyMetadata{
            ID:           "jdk",
            Name:         "BellSoft Liberica JDK",
            Version:      "17.0.1",
            URI:          "https://example.com/jdk.tar.gz",
            SHA256:       "abc123...",
            Stacks:       []string{"*"},
            CPEs:         []string{"cpe:2.3:a:oracle:jdk:17.0.1:*:*:*:*:*:*:*"},
            PURL:         "pkg:generic/bellsoft-jdk@17.0.1",
            Source:       "https://example.com/jdk-src.tar.gz",
            SourceSHA256: "def456...",
            Licenses: []bptest.LicenseMetadata{
                {Type: "GPL-2.0 WITH Classpath-exception-2.0", URI: "https://openjdk.java.net/legal/gplv2+ce.html"},
            },
        }).
        WithPlanEntry("jdk", nil).
        ExecuteT(t, myBuilder.Build)

    if !result.Succeeded() {
        t.Fatalf("build failed: %v", result.Error())
    }
}
```

## API Reference

See the full API documentation on [pkg.go.dev](https://pkg.go.dev/github.com/dmikusa/bptest).

## License

This project is released under version 2.0 of the [Apache License][a].

[a]: http://www.apache.org/licenses/LICENSE-2.0
