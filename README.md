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

## API Reference

See the full API documentation on [pkg.go.dev](https://pkg.go.dev/github.com/dmikusa/bptest).

## License

Apache 2.0
