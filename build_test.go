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

func TestBuildTest_WithAppFile(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		// Verify app file exists
		pomPath := filepath.Join(ctx.ApplicationPath, "pom.xml")
		content, err := os.ReadFile(pomPath)
		if err != nil {
			t.Errorf("failed to read pom.xml: %v", err)
		}
		if string(content) != "<project>test</project>" {
			t.Errorf("unexpected content: %s", string(content))
		}

		// Verify nested file exists
		configPath := filepath.Join(ctx.ApplicationPath, "src", "main", "resources", "config.properties")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("expected nested config file to exist")
		}

		return libcnb.NewBuildResult(), nil
	}

	result := bptest.NewBuildTest().
		WithAppFileString("pom.xml", "<project>test</project>").
		WithAppFile("src/main/resources/config.properties", []byte("key=value")).
		ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Errorf("build failed: %v", result.Error())
	}
}

func TestBuildTest_WithPlatformEnv(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		if ctx.Platform.Environment["BP_JVM_VERSION"] != "17" {
			t.Errorf("expected BP_JVM_VERSION=17, got %s", ctx.Platform.Environment["BP_JVM_VERSION"])
		}
		if ctx.Platform.Environment["BP_MAVEN_BUILT_ARTIFACT"] != "target/*.jar" {
			t.Errorf("expected BP_MAVEN_BUILT_ARTIFACT, got %s", ctx.Platform.Environment["BP_MAVEN_BUILT_ARTIFACT"])
		}
		return libcnb.NewBuildResult(), nil
	}

	result := bptest.NewBuildTest().
		WithPlatformEnv("BP_JVM_VERSION", "17").
		WithPlatformEnv("BP_MAVEN_BUILT_ARTIFACT", "target/*.jar").
		ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Errorf("build failed: %v", result.Error())
	}
}

func TestBuildTest_WithTarget(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		if ctx.TargetInfo.OS != "linux" {
			t.Errorf("expected OS linux, got %s", ctx.TargetInfo.OS)
		}
		if ctx.TargetInfo.Arch != "arm64" {
			t.Errorf("expected Arch arm64, got %s", ctx.TargetInfo.Arch)
		}
		if ctx.TargetDistro.Name != "ubuntu" {
			t.Errorf("expected DistroName ubuntu, got %s", ctx.TargetDistro.Name)
		}
		if ctx.TargetDistro.Version != "22.04" {
			t.Errorf("expected DistroVersion 22.04, got %s", ctx.TargetDistro.Version)
		}
		return libcnb.NewBuildResult(), nil
	}

	result := bptest.NewBuildTest().
		WithTarget(bptest.TargetInfo{
			OS:            "linux",
			Arch:          "arm64",
			DistroName:    "ubuntu",
			DistroVersion: "22.04",
		}).
		ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Errorf("build failed: %v", result.Error())
	}
}

func TestBuildTest_LayerEnvironmentVariables(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, err := ctx.Layers.Layer("jdk")
		if err != nil {
			return libcnb.BuildResult{}, err
		}

		layer.Build = true
		layer.Launch = true

		// Set up environment directories and files
		envBuildDir := filepath.Join(layer.Path, "env.build")
		envLaunchDir := filepath.Join(layer.Path, "env.launch")
		envSharedDir := filepath.Join(layer.Path, "env")

		for _, dir := range []string{envBuildDir, envLaunchDir, envSharedDir} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return libcnb.BuildResult{}, err
			}
		}

		// Write env vars
		os.WriteFile(filepath.Join(envBuildDir, "JAVA_HOME.override"), []byte("/layers/jdk"), 0644)
		os.WriteFile(filepath.Join(envLaunchDir, "JAVA_TOOL_OPTIONS.default"), []byte("-XX:+UseG1GC"), 0644)
		os.WriteFile(filepath.Join(envSharedDir, "PATH.prepend"), []byte("/layers/jdk/bin"), 0644)

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Fatalf("build failed: %v", result.Error())
	}

	layer := result.Layer("jdk")
	if layer == nil {
		t.Fatal("expected jdk layer")
	}

	envBuild := layer.EnvBuild()
	if envBuild["JAVA_HOME.override"] != "/layers/jdk" {
		t.Errorf("expected JAVA_HOME.override, got %v", envBuild)
	}

	envLaunch := layer.EnvLaunch()
	if envLaunch["JAVA_TOOL_OPTIONS.default"] != "-XX:+UseG1GC" {
		t.Errorf("expected JAVA_TOOL_OPTIONS.default, got %v", envLaunch)
	}

	envShared := layer.EnvShared()
	if envShared["PATH.prepend"] != "/layers/jdk/bin" {
		t.Errorf("expected PATH.prepend, got %v", envShared)
	}
}

func TestBuildTest_LayerFileContent(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, err := ctx.Layers.Layer("config")
		if err != nil {
			return libcnb.BuildResult{}, err
		}

		layer.Build = true

		// Create files in the layer
		if err := os.MkdirAll(filepath.Join(layer.Path, "conf"), 0755); err != nil {
			return libcnb.BuildResult{}, err
		}
		os.WriteFile(filepath.Join(layer.Path, "conf", "app.properties"), []byte("key=value\nfoo=bar"), 0644)
		os.WriteFile(filepath.Join(layer.Path, "conf", "binary.dat"), []byte{0x00, 0x01, 0x02}, 0644)

		// Create lib directory with files
		if err := os.MkdirAll(filepath.Join(layer.Path, "lib"), 0755); err != nil {
			return libcnb.BuildResult{}, err
		}
		os.WriteFile(filepath.Join(layer.Path, "lib", "app.jar"), []byte("jar content"), 0644)
		os.WriteFile(filepath.Join(layer.Path, "lib", "utils.jar"), []byte("utils"), 0644)

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Fatalf("build failed: %v", result.Error())
	}

	layer := result.Layer("config")
	if layer == nil {
		t.Fatal("expected config layer")
	}

	// Test FileContentString
	content, err := layer.FileContentString("conf/app.properties")
	if err != nil {
		t.Errorf("FileContentString failed: %v", err)
	}
	if content != "key=value\nfoo=bar" {
		t.Errorf("unexpected content: %s", content)
	}

	// Test FileContent (binary)
	binContent, err := layer.FileContent("conf/binary.dat")
	if err != nil {
		t.Errorf("FileContent failed: %v", err)
	}
	if len(binContent) != 3 || binContent[0] != 0x00 {
		t.Errorf("unexpected binary content: %v", binContent)
	}

	// Test LibFiles
	libFiles, err := layer.LibFiles()
	if err != nil {
		t.Errorf("LibFiles failed: %v", err)
	}
	if len(libFiles) != 2 {
		t.Errorf("expected 2 lib files, got %d", len(libFiles))
	}
}

func TestBuildTest_ReadLayerTOML(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, err := ctx.Layers.Layer("jdk")
		if err != nil {
			return libcnb.BuildResult{}, err
		}

		layer.Build = true
		layer.Launch = true
		layer.Cache = true
		layer.Metadata = map[string]any{
			"version": "17.0.1",
			"vendor":  "adoptium",
		}

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}

	bt := bptest.NewBuildTest()
	result := bt.ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Fatalf("build failed: %v", result.Error())
	}

	// Use ReadLayerTOML to inspect the layer.toml file
	// Note: The layer.toml is written by libcnb after build, but in our test
	// we can read from existing layers set up with WithExistingLayer
	// Let's test with an existing layer instead
}

func TestBuildTest_ReadLayerTOML_WithExistingLayer(t *testing.T) {
	bt := bptest.NewBuildTest().
		WithExistingLayer("cached-jdk", &bptest.LayerConfig{
			Build:  true,
			Launch: true,
			Cache:  true,
			Metadata: map[string]any{
				"version": "17.0.1",
				"vendor":  "adoptium",
			},
		})

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		return libcnb.NewBuildResult(), nil
	}

	result := bt.ExecuteT(t, builder)
	if !result.Succeeded() {
		t.Fatalf("build failed: %v", result.Error())
	}

	// Read the layer.toml that was set up by WithExistingLayer
	paths := bt.Paths()
	layerContent, err := bptest.ReadLayerTOML(paths.LayersDir, "cached-jdk")
	if err != nil {
		t.Fatalf("ReadLayerTOML failed: %v", err)
	}

	if !layerContent.Build {
		t.Error("expected Build to be true")
	}
	if !layerContent.Launch {
		t.Error("expected Launch to be true")
	}
	if !layerContent.Cache {
		t.Error("expected Cache to be true")
	}
	if layerContent.Metadata["version"] != "17.0.1" {
		t.Errorf("expected version 17.0.1, got %v", layerContent.Metadata["version"])
	}
}

func TestBuildTest_WithBuildpack(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		if ctx.Buildpack.Info.ID != "my/buildpack" {
			t.Errorf("expected buildpack ID my/buildpack, got %s", ctx.Buildpack.Info.ID)
		}
		if ctx.Buildpack.Info.Version != "2.0.0" {
			t.Errorf("expected version 2.0.0, got %s", ctx.Buildpack.Info.Version)
		}
		return libcnb.NewBuildResult(), nil
	}

	result := bptest.NewBuildTest().
		WithBuildpack(bptest.BuildpackConfig{
			API:     "0.9",
			ID:      "my/buildpack",
			Version: "2.0.0",
			Name:    "My Buildpack",
		}).
		ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Errorf("build failed: %v", result.Error())
	}
}

func TestBuildTest_LayerInspectorGet(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer1, _ := ctx.Layers.Layer("layer1")
		layer1.Build = true
		layer2, _ := ctx.Layers.Layer("layer2")
		layer2.Launch = true

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer1, layer2)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Fatalf("build failed: %v", result.Error())
	}

	layers := result.Layers()

	// Test Get
	layer1 := layers.Get("layer1")
	if layer1 == nil {
		t.Fatal("expected layer1")
	}
	if layer1.Name() != "layer1" {
		t.Errorf("expected name layer1, got %s", layer1.Name())
	}
	if layer1.Path() == "" {
		t.Error("expected non-empty path")
	}

	// Test Get for non-existent layer
	nonexistent := layers.Get("nonexistent")
	if nonexistent != nil {
		t.Error("expected nil for nonexistent layer")
	}

	// Test Metadata() returning full map
	layer1.Metadata() // Just ensure it doesn't panic
}

func TestBuildTest_LayerExecDFiles(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, _ := ctx.Layers.Layer("helper")
		layer.Build = true

		// Create exec.d directory with scripts
		execDDir := filepath.Join(layer.Path, "exec.d")
		if err := os.MkdirAll(execDDir, 0755); err != nil {
			return libcnb.BuildResult{}, err
		}
		os.WriteFile(filepath.Join(execDDir, "0-setup"), []byte("#!/bin/bash"), 0755)
		os.WriteFile(filepath.Join(execDDir, "1-configure"), []byte("#!/bin/bash"), 0755)

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Fatalf("build failed: %v", result.Error())
	}

	layer := result.Layer("helper")
	execDFiles, err := layer.ExecDFiles()
	if err != nil {
		t.Errorf("ExecDFiles failed: %v", err)
	}
	if len(execDFiles) != 2 {
		t.Errorf("expected 2 exec.d files, got %d", len(execDFiles))
	}
}

func TestDetectResult_Plans(t *testing.T) {
	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{
			Pass: true,
			Plans: []libcnb.BuildPlan{
				{
					Provides: []libcnb.BuildPlanProvide{{Name: "jdk"}},
					Requires: []libcnb.BuildPlanRequire{{Name: "jdk"}},
				},
			},
		}, nil
	}

	result := bptest.NewDetectTest().ExecuteT(t, detector)

	plans := result.Plans()
	if len(plans) != 1 {
		t.Errorf("expected 1 plan, got %d", len(plans))
	}
}

func TestBuildTest_WithPlan(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		if len(ctx.Plan.Entries) != 2 {
			t.Errorf("expected 2 plan entries, got %d", len(ctx.Plan.Entries))
		}
		if ctx.Plan.Entries[0].Name != "jdk" {
			t.Errorf("expected first entry to be jdk, got %s", ctx.Plan.Entries[0].Name)
		}
		if ctx.Plan.Entries[1].Name != "jre" {
			t.Errorf("expected second entry to be jre, got %s", ctx.Plan.Entries[1].Name)
		}
		return libcnb.NewBuildResult(), nil
	}

	plan := libcnb.BuildpackPlan{
		Entries: []libcnb.BuildpackPlanEntry{
			{Name: "jdk", Metadata: map[string]any{"version": "17"}},
			{Name: "jre", Metadata: map[string]any{"version": "17"}},
		},
	}

	result := bptest.NewBuildTest().
		WithPlan(plan).
		ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Error("expected build to succeed")
	}
}

func TestBuildTest_WithStackID(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		if ctx.StackID != "io.buildpacks.stacks.jammy" {
			t.Errorf("expected stack ID io.buildpacks.stacks.jammy, got %s", ctx.StackID)
		}
		return libcnb.NewBuildResult(), nil
	}

	result := bptest.NewBuildTest().
		WithStackID("io.buildpacks.stacks.jammy").
		ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Error("expected build to succeed")
	}
}

func TestBuildTest_WithConfigurationMetadata(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		configs, ok := ctx.Buildpack.Metadata["configurations"].([]map[string]any)
		if !ok {
			t.Error("expected configurations in metadata")
			return libcnb.NewBuildResult(), nil
		}
		if len(configs) != 2 {
			t.Errorf("expected 2 configurations, got %d", len(configs))
		}
		// Check first config
		if configs[0]["name"] != "BPL_JVM_HEAD_ROOM" {
			t.Errorf("expected BPL_JVM_HEAD_ROOM, got %v", configs[0]["name"])
		}
		if configs[0]["launch"] != true {
			t.Error("expected launch to be true")
		}
		// Check second config
		if configs[1]["name"] != "BP_JVM_VERSION" {
			t.Errorf("expected BP_JVM_VERSION, got %v", configs[1]["name"])
		}
		if configs[1]["build"] != true {
			t.Error("expected build to be true")
		}
		return libcnb.NewBuildResult(), nil
	}

	result := bptest.NewBuildTest().
		WithConfigurationMetadata(bptest.ConfigurationMetadata{
			Name:        "BPL_JVM_HEAD_ROOM",
			Description: "the headroom in memory calculation",
			Default:     "0",
			Launch:      true,
		}).
		WithConfigurationMetadata(bptest.ConfigurationMetadata{
			Name:        "BP_JVM_VERSION",
			Description: "the Java version",
			Default:     "17",
			Build:       true,
		}).
		ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Error("expected build to succeed")
	}
}

func TestBuildTest_WithDependencyMetadata(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		deps, ok := ctx.Buildpack.Metadata["dependencies"].([]map[string]any)
		if !ok {
			t.Error("expected dependencies in metadata")
			return libcnb.NewBuildResult(), nil
		}
		if len(deps) != 1 {
			t.Errorf("expected 1 dependency, got %d", len(deps))
		}
		if deps[0]["id"] != "jdk" {
			t.Errorf("expected jdk, got %v", deps[0]["id"])
		}
		if deps[0]["version"] != "8.0.482" {
			t.Errorf("expected 8.0.482, got %v", deps[0]["version"])
		}
		// Check licenses
		licenses, ok := deps[0]["licenses"].([]map[string]any)
		if !ok {
			t.Error("expected licenses in dependency")
			return libcnb.NewBuildResult(), nil
		}
		if len(licenses) != 1 {
			t.Errorf("expected 1 license, got %d", len(licenses))
		}
		return libcnb.NewBuildResult(), nil
	}

	result := bptest.NewBuildTest().
		WithDependencyMetadata(bptest.DependencyMetadata{
			ID:           "jdk",
			Name:         "BellSoft Liberica JDK",
			Version:      "8.0.482",
			URI:          "https://example.com/jdk.tar.gz",
			SHA256:       "abc123",
			Stacks:       []string{"*"},
			CPEs:         []string{"cpe:2.3:a:oracle:jdk:1.8.0:*"},
			PURL:         "pkg:generic/bellsoft-jdk@8.0.482",
			Source:       "https://example.com/jdk-src.tar.gz",
			SourceSHA256: "def456",
			Licenses: []bptest.LicenseMetadata{
				{
					Type: "GPL-2.0",
					URI:  "https://example.com/license",
				},
			},
		}).
		ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Error("expected build to succeed")
	}
}

func TestDetectTest_WithConfigurationMetadata(t *testing.T) {
	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		configs, ok := ctx.Buildpack.Metadata["configurations"].([]map[string]any)
		if !ok {
			t.Error("expected configurations in metadata")
			return libcnb.DetectResult{Pass: false}, nil
		}
		if len(configs) != 1 {
			t.Errorf("expected 1 configuration, got %d", len(configs))
		}
		if configs[0]["name"] != "BP_JAVA_VERSION" {
			t.Errorf("expected BP_JAVA_VERSION, got %v", configs[0]["name"])
		}
		return libcnb.DetectResult{Pass: true}, nil
	}

	result := bptest.NewDetectTest().
		WithConfigurationMetadata(bptest.ConfigurationMetadata{
			Name:        "BP_JAVA_VERSION",
			Description: "the Java version",
			Default:     "17",
		}).
		ExecuteT(t, detector)

	if !result.Passed() {
		t.Error("expected detect to pass")
	}
}

func TestDetectTest_WithDependencyMetadata(t *testing.T) {
	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		deps, ok := ctx.Buildpack.Metadata["dependencies"].([]map[string]any)
		if !ok {
			t.Error("expected dependencies in metadata")
			return libcnb.DetectResult{Pass: false}, nil
		}
		if len(deps) != 1 {
			t.Errorf("expected 1 dependency, got %d", len(deps))
		}
		if deps[0]["id"] != "jdk" {
			t.Errorf("expected jdk, got %v", deps[0]["id"])
		}
		return libcnb.DetectResult{Pass: true}, nil
	}

	result := bptest.NewDetectTest().
		WithDependencyMetadata(bptest.DependencyMetadata{
			ID:      "jdk",
			Name:    "OpenJDK",
			Version: "17.0.1",
			URI:     "https://example.com/jdk.tar.gz",
			SHA256:  "abc123",
			Stacks:  []string{"*"},
		}).
		ExecuteT(t, detector)

	if !result.Passed() {
		t.Error("expected detect to pass")
	}
}

func TestBuildTest_WithEnv(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		// Check that environment variable is set
		if os.Getenv("JAVA_HOME") != "/layers/jdk" {
			t.Errorf("expected JAVA_HOME=/layers/jdk, got %s", os.Getenv("JAVA_HOME"))
		}
		if os.Getenv("BPTEST_VAR") != "test_value" {
			t.Errorf("expected BPTEST_VAR=test_value, got %s", os.Getenv("BPTEST_VAR"))
		}
		return libcnb.NewBuildResult(), nil
	}

	result := bptest.NewBuildTest().
		WithEnv("JAVA_HOME", "/layers/jdk").
		WithEnv("BPTEST_VAR", "test_value").
		ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Error("expected build to succeed")
	}

	// Verify env vars are cleaned up after execution
	if os.Getenv("BPTEST_VAR") != "" {
		t.Error("expected BPTEST_VAR to be unset after test")
	}
}

func TestDetectTest_WithEnv(t *testing.T) {
	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		// Check that environment variable is set
		if os.Getenv("JAVA_HOME") != "/layers/jdk" {
			t.Errorf("expected JAVA_HOME=/layers/jdk, got %s", os.Getenv("JAVA_HOME"))
		}
		return libcnb.DetectResult{Pass: true}, nil
	}

	result := bptest.NewDetectTest().
		WithEnv("JAVA_HOME", "/layers/jdk").
		ExecuteT(t, detector)

	if !result.Passed() {
		t.Error("expected detect to pass")
	}
}

func TestBuildTest_WithBuildpackFile(t *testing.T) {
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		// Check that the buildpack file exists
		content, err := os.ReadFile(filepath.Join(ctx.Buildpack.Path, "config", "settings.json"))
		if err != nil {
			t.Errorf("failed to read buildpack file: %v", err)
		}
		if string(content) != `{"key": "value"}` {
			t.Errorf("unexpected content: %s", string(content))
		}
		return libcnb.NewBuildResult(), nil
	}

	result := bptest.NewBuildTest().
		WithBuildpackFileString("config/settings.json", `{"key": "value"}`).
		ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Error("expected build to succeed")
	}
}

func TestBuildTest_WithBuildpackSymlink(t *testing.T) {
	// Create a temp directory to serve as the symlink target
	targetDir, err := os.MkdirTemp("", "bptest-target-")
	if err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}
	defer os.RemoveAll(targetDir)

	// Create a file in the target directory
	if err := os.WriteFile(filepath.Join(targetDir, "dependency.tgz"), []byte("mock dependency"), 0644); err != nil {
		t.Fatalf("failed to write target file: %v", err)
	}

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		// Check that the symlink exists and points to the correct target
		symlinkPath := filepath.Join(ctx.Buildpack.Path, "dependencies")

		// Read the symlink
		target, err := os.Readlink(symlinkPath)
		if err != nil {
			t.Errorf("failed to read symlink: %v", err)
		}
		if target != targetDir {
			t.Errorf("symlink target = %s, want %s", target, targetDir)
		}

		// Check that we can read files through the symlink
		content, err := os.ReadFile(filepath.Join(symlinkPath, "dependency.tgz"))
		if err != nil {
			t.Errorf("failed to read through symlink: %v", err)
		}
		if string(content) != "mock dependency" {
			t.Errorf("unexpected content through symlink: %s", string(content))
		}

		return libcnb.NewBuildResult(), nil
	}

	result := bptest.NewBuildTest().
		WithBuildpackSymlink("dependencies", targetDir).
		ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Error("expected build to succeed")
	}
}

func TestDetectTest_WithBuildpackFile(t *testing.T) {
	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		// Check that the buildpack file exists
		content, err := os.ReadFile(filepath.Join(ctx.Buildpack.Path, "config.yaml"))
		if err != nil {
			t.Errorf("failed to read buildpack file: %v", err)
		}
		if string(content) != "enabled: true" {
			t.Errorf("unexpected content: %s", string(content))
		}
		return libcnb.DetectResult{Pass: true}, nil
	}

	result := bptest.NewDetectTest().
		WithBuildpackFileString("config.yaml", "enabled: true").
		ExecuteT(t, detector)

	if !result.Passed() {
		t.Error("expected detect to pass")
	}
}

func TestBuildTest_WithBuildpackDependencyCache(t *testing.T) {
	// Create a temp directory to serve as the dependency cache
	targetDir, err := os.MkdirTemp("", "bptest-deps-")
	if err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}
	defer os.RemoveAll(targetDir)

	// Create a mock dependency file
	if err := os.WriteFile(filepath.Join(targetDir, "jdk-17.tar.gz"), []byte("mock jdk"), 0644); err != nil {
		t.Fatalf("failed to write target file: %v", err)
	}

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		// Verify the dependencies symlink exists
		content, err := os.ReadFile(filepath.Join(ctx.Buildpack.Path, "dependencies", "jdk-17.tar.gz"))
		if err != nil {
			t.Errorf("failed to read through dependencies symlink: %v", err)
		}
		if string(content) != "mock jdk" {
			t.Errorf("unexpected content: %s", string(content))
		}
		return libcnb.NewBuildResult(), nil
	}

	result := bptest.NewBuildTest().
		WithBuildpackDependencyCache(targetDir).
		ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Error("expected build to succeed")
	}
}

func TestDetectTest_WithBuildpackSymlink(t *testing.T) {
	// Create a temp directory to serve as the symlink target
	targetDir, err := os.MkdirTemp("", "bptest-target-")
	if err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}
	defer os.RemoveAll(targetDir)

	// Create a file in the target directory
	if err := os.WriteFile(filepath.Join(targetDir, "cached-file.txt"), []byte("cached content"), 0644); err != nil {
		t.Fatalf("failed to write target file: %v", err)
	}

	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		// Check that we can read files through the symlink
		content, err := os.ReadFile(filepath.Join(ctx.Buildpack.Path, "cache", "cached-file.txt"))
		if err != nil {
			t.Errorf("failed to read through symlink: %v", err)
		}
		if string(content) != "cached content" {
			t.Errorf("unexpected content through symlink: %s", string(content))
		}
		return libcnb.DetectResult{Pass: true}, nil
	}

	result := bptest.NewDetectTest().
		WithBuildpackSymlink("cache", targetDir).
		ExecuteT(t, detector)

	if !result.Passed() {
		t.Error("expected detect to pass")
	}
}

func TestBuildTest_FromBuildpack(t *testing.T) {
	// Create a temp buildpack.toml file
	tmpDir, err := os.MkdirTemp("", "bptest-toml-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	buildpackTOML := `api = "0.8"

[buildpack]
id = "integration/test-buildpack"
name = "Integration Test Buildpack"
version = "1.2.3"

[[buildpack.licenses]]
type = "Apache-2.0"

[metadata]
include-files = ["buildpack.toml"]

[[metadata.configurations]]
name = "BP_TEST_CONFIG"
description = "A test configuration"
default = "default-value"
build = true
`
	tomlPath := filepath.Join(tmpDir, "buildpack.toml")
	if err := os.WriteFile(tomlPath, []byte(buildpackTOML), 0644); err != nil {
		t.Fatalf("failed to write buildpack.toml: %v", err)
	}

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		// Verify the buildpack info was loaded from the real file
		if ctx.Buildpack.Info.ID != "integration/test-buildpack" {
			t.Errorf("expected buildpack ID integration/test-buildpack, got %s", ctx.Buildpack.Info.ID)
		}
		if ctx.Buildpack.Info.Version != "1.2.3" {
			t.Errorf("expected version 1.2.3, got %s", ctx.Buildpack.Info.Version)
		}
		return libcnb.NewBuildResult(), nil
	}

	result := bptest.NewBuildTestFromBuildpack(tomlPath).
		ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Error("expected build to succeed")
	}
}

func TestBuildTest_FromBuildpack_WithDependencyCache(t *testing.T) {
	// Create a temp buildpack.toml file
	tmpDir, err := os.MkdirTemp("", "bptest-toml-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	buildpackTOML := `api = "0.8"

[buildpack]
id = "integration/test-buildpack"
version = "1.0.0"
`
	tomlPath := filepath.Join(tmpDir, "buildpack.toml")
	if err := os.WriteFile(tomlPath, []byte(buildpackTOML), 0644); err != nil {
		t.Fatalf("failed to write buildpack.toml: %v", err)
	}

	// Create a dependency cache directory
	cacheDir, err := os.MkdirTemp("", "bptest-cache-")
	if err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}
	defer os.RemoveAll(cacheDir)

	if err := os.WriteFile(filepath.Join(cacheDir, "dep.tar.gz"), []byte("mock dep"), 0644); err != nil {
		t.Fatalf("failed to write mock dep: %v", err)
	}

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		// Verify the dependency cache symlink works
		content, err := os.ReadFile(filepath.Join(ctx.Buildpack.Path, "dependencies", "dep.tar.gz"))
		if err != nil {
			t.Errorf("failed to read through dependencies symlink: %v", err)
		}
		if string(content) != "mock dep" {
			t.Errorf("unexpected content: %s", string(content))
		}
		return libcnb.NewBuildResult(), nil
	}

	result := bptest.NewBuildTestFromBuildpack(tomlPath).
		WithBuildpackDependencyCache(cacheDir).
		ExecuteT(t, builder)

	if !result.Succeeded() {
		t.Error("expected build to succeed")
	}
}

func TestDetectTest_FromBuildpack(t *testing.T) {
	// Create a temp buildpack.toml file
	tmpDir, err := os.MkdirTemp("", "bptest-toml-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	buildpackTOML := `api = "0.8"

[buildpack]
id = "integration/detect-test"
name = "Integration Detect Test"
version = "2.0.0"
`
	tomlPath := filepath.Join(tmpDir, "buildpack.toml")
	if err := os.WriteFile(tomlPath, []byte(buildpackTOML), 0644); err != nil {
		t.Fatalf("failed to write buildpack.toml: %v", err)
	}

	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		// Verify the buildpack info was loaded from the real file
		if ctx.Buildpack.Info.ID != "integration/detect-test" {
			t.Errorf("expected buildpack ID integration/detect-test, got %s", ctx.Buildpack.Info.ID)
		}
		if ctx.Buildpack.Info.Version != "2.0.0" {
			t.Errorf("expected version 2.0.0, got %s", ctx.Buildpack.Info.Version)
		}
		return libcnb.DetectResult{Pass: true}, nil
	}

	result := bptest.NewDetectTestFromBuildpack(tomlPath).
		ExecuteT(t, detector)

	if !result.Passed() {
		t.Error("expected detect to pass")
	}
}
