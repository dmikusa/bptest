package bptest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb/v2"

	"github.com/dmikusa/bptest"
)

func TestBuildTest_BasicSuccess(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		return libcnb.NewBuildResult(), nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Error("expected build to succeed")
	}
	if result.HasError() {
		t.Errorf("unexpected error: %v", result.Error())
	}
}

func TestBuildTest_WithError(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		return libcnb.BuildResult{}, os.ErrPermission
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	if result.Succeeded() {
		t.Error("expected build to fail")
	}
	if !result.HasError() {
		t.Error("expected HasError() to be true")
	}
	if result.Error() != os.ErrPermission {
		t.Errorf("expected os.ErrPermission, got %v", result.Error())
	}
}

func TestBuildTest_WithPlanEntry(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		if len(ctx.Plan.Entries) != 1 {
			t.Errorf("expected 1 plan entry, got %d", len(ctx.Plan.Entries))
		}
		entry := ctx.Plan.Entries[0]
		if entry.Name != "jdk" {
			t.Errorf("expected jdk entry, got %s", entry.Name)
		}
		if entry.Metadata["version"] != "17" {
			t.Errorf("expected version 17, got %v", entry.Metadata["version"])
		}
		return libcnb.NewBuildResult(), nil
	}

	result := bptest.NewBuildTest().
		WithPlanEntry("jdk", map[string]any{"version": "17"}).
		ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Errorf("build failed: %v", result.Error())
	}
}

func TestBuildTest_WithLayer(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, err := ctx.Layers.Layer("jdk")
		if err != nil {
			return libcnb.BuildResult{}, err
		}

		layer.Build = true
		layer.Cache = true
		layer.Metadata = map[string]any{"version": "17"}

		// Create a file in the layer
		if err := os.MkdirAll(filepath.Join(layer.Path, "bin"), 0755); err != nil {
			return libcnb.BuildResult{}, err
		}
		if err := os.WriteFile(filepath.Join(layer.Path, "bin", "java"), []byte("#!/bin/bash"), 0755); err != nil {
			return libcnb.BuildResult{}, err
		}

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Fatalf("build failed: %v", result.Error())
	}

	layers := result.Layers()
	if layers.Count() != 1 {
		t.Errorf("expected 1 layer, got %d", layers.Count())
	}

	names := layers.Names()
	if len(names) != 1 || names[0] != "jdk" {
		t.Errorf("expected layer names [jdk], got %v", names)
	}

	layer := result.Layer("jdk")
	if layer == nil {
		t.Fatal("expected jdk layer")
	}

	if !layer.IsBuild() {
		t.Error("expected layer to be build")
	}
	if !layer.IsCache() {
		t.Error("expected layer to be cache")
	}
	if layer.IsLaunch() {
		t.Error("expected layer to not be launch")
	}

	if layer.MetadataValue("version") != "17" {
		t.Errorf("expected version 17, got %v", layer.MetadataValue("version"))
	}

	if !layer.FileExists("bin/java") {
		t.Error("expected bin/java to exist")
	}

	binFiles, err := layer.BinFiles()
	if err != nil {
		t.Errorf("BinFiles failed: %v", err)
	}
	if len(binFiles) != 1 || binFiles[0] != "java" {
		t.Errorf("expected bin files [java], got %v", binFiles)
	}
}

func TestBuildTest_WithProcess(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		result := libcnb.NewBuildResult()
		result.Processes = []libcnb.Process{
			{
				Type:    "web",
				Command: []string{"java", "-jar", "app.jar"},
				Default: true,
			},
			{
				Type:    "worker",
				Command: []string{"java", "-cp", "app.jar", "Worker"},
			},
		}
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Fatalf("build failed: %v", result.Error())
	}

	processes := result.Processes()
	if len(processes) != 2 {
		t.Errorf("expected 2 processes, got %d", len(processes))
	}

	if !result.HasProcess("web") {
		t.Error("expected web process")
	}
	if !result.HasProcess("worker") {
		t.Error("expected worker process")
	}
	if result.HasProcess("nonexistent") {
		t.Error("expected nonexistent process to not be found")
	}

	webProcess := result.Process("web")
	if webProcess == nil {
		t.Fatal("expected web process")
	}
	if !webProcess.Default {
		t.Error("expected web process to be default")
	}

	defaultProcess := result.DefaultProcess()
	if defaultProcess == nil || defaultProcess.Type != "web" {
		t.Error("expected default process to be web")
	}
}

func TestBuildTest_WithLabels(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		result := libcnb.NewBuildResult()
		result.Labels = []libcnb.Label{
			{Key: "org.opencontainers.image.title", Value: "My App"},
			{Key: "org.opencontainers.image.version", Value: "1.0.0"},
		}
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Fatalf("build failed: %v", result.Error())
	}

	labels := result.Labels()
	if len(labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(labels))
	}

	if result.Label("org.opencontainers.image.title") != "My App" {
		t.Error("expected title label")
	}
	if result.Label("nonexistent") != "" {
		t.Error("expected empty string for nonexistent label")
	}
}

func TestBuildTest_WithExistingLayer(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, err := ctx.Layers.Layer("jdk")
		if err != nil {
			return libcnb.BuildResult{}, err
		}

		// Check that existing layer metadata is available
		if layer.Metadata["version"] != "11" {
			t.Errorf("expected existing version 11, got %v", layer.Metadata["version"])
		}

		// Check that existing file is present
		existingFile := filepath.Join(layer.Path, "marker.txt")
		if _, err := os.Stat(existingFile); os.IsNotExist(err) {
			t.Error("expected marker.txt to exist from previous build")
		}

		// Update layer for new version
		layer.Build = true
		layer.Cache = true
		layer.Metadata["version"] = "17"

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}

	result := bptest.NewBuildTest().
		WithExistingLayer("jdk", &bptest.LayerConfig{
			Build: true,
			Cache: true,
			Metadata: map[string]any{"version": "11"},
			Files: map[string][]byte{
				"marker.txt": []byte("previous build"),
			},
		}).
		ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Fatalf("build failed: %v", result.Error())
	}

	layer := result.Layer("jdk")
	if layer.MetadataValue("version") != "17" {
		t.Errorf("expected updated version 17, got %v", layer.MetadataValue("version"))
	}
}

func TestBuildTest_WithPersistentMetadata(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		if ctx.PersistentMetadata["last-build"] != "2024-01-01" {
			t.Errorf("expected persistent metadata, got %v", ctx.PersistentMetadata)
		}

		result := libcnb.NewBuildResult()
		result.PersistentMetadata["last-build"] = "2024-01-02"
		return result, nil
	}

	result := bptest.NewBuildTest().
		WithPersistentMetadata(map[string]any{"last-build": "2024-01-01"}).
		ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Fatalf("build failed: %v", result.Error())
	}

	if result.PersistentMetadata()["last-build"] != "2024-01-02" {
		t.Error("expected updated persistent metadata")
	}
}

func TestBuildTest_WithBinding(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		for _, b := range ctx.Platform.Bindings {
			if b.Name == "credentials" && b.Type == "maven" {
				if b.Secret["username"] == "admin" {
					return libcnb.NewBuildResult(), nil
				}
			}
		}
		t.Error("expected binding not found")
		return libcnb.NewBuildResult(), nil
	}

	result := bptest.NewBuildTest().
		WithBinding(bptest.Binding{
			Name:   "credentials",
			Type:   "maven",
			Secret: map[string]string{"username": "admin", "password": "secret"},
		}).
		ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Errorf("build failed: %v", result.Error())
	}
}

func TestBuildTest_WithSlices(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		result := libcnb.NewBuildResult()
		result.Slices = []libcnb.Slice{
			{Paths: []string{"app/**"}},
			{Paths: []string{"lib/**"}},
		}
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Fatalf("build failed: %v", result.Error())
	}

	slices := result.Slices()
	if len(slices) != 2 {
		t.Errorf("expected 2 slices, got %d", len(slices))
	}
}

func TestBuildTest_WithUnmetEntries(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		result := libcnb.NewBuildResult()
		result.Unmet = []libcnb.UnmetPlanEntry{
			{Name: "optional-dep"},
		}
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Fatalf("build failed: %v", result.Error())
	}

	unmet := result.UnmetEntries()
	if len(unmet) != 1 || unmet[0].Name != "optional-dep" {
		t.Errorf("expected unmet entry [optional-dep], got %v", unmet)
	}
}
