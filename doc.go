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

// Package bptest provides testing utilities for Cloud Native Buildpacks.
//
// This library allows you to test buildpack detect and build functions directly
// by invoking them with configurable test data, without needing to run the full
// buildpack lifecycle or spawn subprocesses.
//
// # Detect Tests
//
// Use DetectTest to test your buildpack's detect function:
//
//	func TestDetect_WithPomXML(t *testing.T) {
//		result := bptest.NewDetectTest().
//			WithAppFileString("pom.xml", `<project><modelVersion>4.0.0</modelVersion></project>`).
//			ExecuteT(t, myDetector.Detect)
//
//		if !result.Passed() {
//			t.Error("expected detect to pass when pom.xml exists")
//		}
//		if !result.HasPlan("jdk") {
//			t.Error("expected to require jdk")
//		}
//	}
//
// # Build Tests
//
// Use BuildTest to test your buildpack's build function:
//
//	func TestBuild_LayerCreation(t *testing.T) {
//		result := bptest.NewBuildTest().
//			WithAppFileString("pom.xml", "<project/>").
//			WithPlanEntry("jdk", map[string]any{"version": "17"}).
//			ExecuteT(t, myBuilder.Build)
//
//		if !result.Succeeded() {
//			t.Fatalf("build failed: %v", result.Error())
//		}
//
//		layer := result.Layer("jdk")
//		if !layer.IsBuild() || !layer.IsCache() {
//			t.Error("expected build and cache layer")
//		}
//		if !layer.FileExists("bin/java") {
//			t.Error("expected java binary")
//		}
//	}
//
// # Service Bindings
//
// Both DetectTest and BuildTest support service bindings:
//
//	func TestDetect_WithBinding(t *testing.T) {
//		result := bptest.NewDetectTest().
//			WithBinding(bptest.Binding{
//				Name:   "my-database",
//				Type:   "postgresql",
//				Secret: map[string]string{
//					"host":     "localhost",
//					"username": "admin",
//					"password": "secret",
//				},
//			}).
//			ExecuteT(t, myDetector.Detect)
//
//		if !result.Passed() {
//			t.Fatal("expected detect to pass with database binding")
//		}
//	}
//
// # Rebuild Scenarios
//
// Test rebuild scenarios by providing existing layer state:
//
//	func TestBuild_Rebuild(t *testing.T) {
//		result := bptest.NewBuildTest().
//			WithPlanEntry("jdk", map[string]any{"version": "17"}).
//			WithExistingLayer("jdk", &bptest.LayerConfig{
//				Build: true,
//				Cache: true,
//				Metadata: map[string]any{"version": "11"},
//				Files: map[string][]byte{
//					"bin/java": []byte("old java"),
//				},
//			}).
//			ExecuteT(t, myBuilder.Build)
//
//		if !result.Succeeded() {
//			t.Fatalf("rebuild failed: %v", result.Error())
//		}
//
//		layer := result.Layer("jdk")
//		if layer.MetadataValue("version") != "17" {
//			t.Error("expected layer to be updated to version 17")
//		}
//	}
//
// # Platform Environment Variables
//
// Set platform environment variables that your buildpack reads:
//
//	func TestDetect_WithEnvVar(t *testing.T) {
//		result := bptest.NewDetectTest().
//			WithAppFileString("pom.xml", "<project/>").
//			WithPlatformEnv("BP_JAVA_VERSION", "17").
//			ExecuteT(t, myDetector.Detect)
//
//		if !result.Passed() {
//			t.Fatal("expected detect to pass")
//		}
//	}
//
// # Inspecting Build Plans
//
// Examine the build plans returned by detect:
//
//	func TestDetect_BuildPlan(t *testing.T) {
//		result := bptest.NewDetectTest().
//			WithAppFileString("pom.xml", "<project/>").
//			ExecuteT(t, myDetector.Detect)
//
//		provides := result.Provides()  // []string{"jdk", "jre"}
//		requires := result.Requires()  // []string{"jdk"}
//
//		metadata := result.RequireMetadata("jdk")
//		if metadata["version"] != "17" {
//			t.Errorf("expected version 17, got %v", metadata["version"])
//		}
//	}
//
// # Inspecting Layers
//
// Examine layers created during build:
//
//	func TestBuild_LayerInspection(t *testing.T) {
//		result := bptest.NewBuildTest().
//			WithPlanEntry("jdk", nil).
//			ExecuteT(t, myBuilder.Build)
//
//		layer := result.Layer("jdk")
//
//		// Check layer types
//		layer.IsBuild()   // true if build layer
//		layer.IsLaunch()  // true if launch layer
//		layer.IsCache()   // true if cached layer
//
//		// Check metadata
//		layer.MetadataValue("version")
//
//		// Check files
//		layer.FileExists("bin/java")
//		content, _ := layer.FileContentString("release")
//
//		// Check environment variables
//		envBuild := layer.EnvBuild()     // build-time env vars
//		envLaunch := layer.EnvLaunch()   // launch-time env vars
//	}
//
// # Target Platform Information
//
// Set target OS, architecture, and distribution:
//
//	func TestBuild_WithTarget(t *testing.T) {
//		result := bptest.NewBuildTest().
//			WithTarget(bptest.TargetInfo{
//				OS:            "linux",
//				Arch:          "arm64",
//				DistroName:    "ubuntu",
//				DistroVersion: "22.04",
//			}).
//			WithPlanEntry("jdk", nil).
//			ExecuteT(t, myBuilder.Build)
//
//		if !result.Succeeded() {
//			t.Fatalf("build failed: %v", result.Error())
//		}
//	}
//
// # Dependency Caching
//
// When testing buildpacks that download dependencies, use WithBuildpackDependencyCache
// to point the buildpack's dependency cache to a local test fixtures directory:
//
//	func TestBuild_WithDependencyCache(t *testing.T) {
//		testdataAbs, _ := filepath.Abs("testdata")
//
//		result := bptest.NewBuildTest().
//			WithBuildpackDependencyCache(testdataAbs).
//			WithPlanEntry("jdk", nil).
//			ExecuteT(t, myBuilder.Build)
//
//		if !result.Succeeded() {
//			t.Fatalf("build failed: %v", result.Error())
//		}
//	}
//
// You can also add arbitrary files to the buildpack directory with WithBuildpackFile:
//
//	result := bptest.NewBuildTest().
//		WithBuildpackFileString("config/settings.json", `{"debug": true}`).
//		ExecuteT(t, myBuilder.Build)
//
// # Gomega Matchers
//
// The matchers subpackage provides custom Gomega matchers for more expressive assertions.
// Import with dot notation for cleaner test code:
//
//	import (
//		. "github.com/onsi/gomega"
//		"github.com/dmikusa/bptest"
//		. "github.com/dmikusa/bptest/matchers"
//	)
//
//	func TestDetect_WithGomega(t *testing.T) {
//		g := NewWithT(t)
//
//		result := bptest.NewDetectTest().
//			WithAppFileString("pom.xml", "<project/>").
//			ExecuteT(t, myDetector.Detect)
//
//		g.Expect(result).To(HavePassed())
//		g.Expect(result).To(HavePlan("jdk"))
//		g.Expect(result).To(Provide("jdk"))
//		g.Expect(result).To(Require("jvm"))
//	}
//
//	func TestBuild_WithGomega(t *testing.T) {
//		g := NewWithT(t)
//
//		result := bptest.NewBuildTest().
//			WithPlanEntry("jdk", nil).
//			ExecuteT(t, myBuilder.Build)
//
//		g.Expect(result).To(HaveSucceeded())
//		g.Expect(result).To(HaveLayer("jdk"))
//		g.Expect(result).To(HaveProcess("web"))
//		g.Expect(result).To(HaveLabel("org.opencontainers.image.title", "My App"))
//	}
//
// Available matchers for DetectResult:
//   - HavePassed() - check if detection passed
//   - HaveFailed() - check if detection failed
//   - HavePlan(name) - check if plan provides or requires name
//   - Provide(name) - check if provides includes name
//   - Require(name) - check if requires includes name
//
// Available matchers for BuildResult:
//   - HaveSucceeded() - check if build succeeded
//   - HaveLayer(name) - check if layer exists
//   - HaveLayerCount(n) - check exact number of layers
//   - HaveLayers(names...) - check that specified layers exist (others may exist)
//   - HaveExactlyLayers(names...) - check that only specified layers exist
//   - HaveProcess(type) - check if process type exists
//   - HaveLabel(key) or HaveLabel(key, value) - check label exists/matches value
//
// Available matchers for LayerResult:
//   - HaveFile(path) - check if file exists in layer
//   - HaveFileWithContents(path, contents) - check file exists with exact contents
//   - BeBuildLayer() - check if layer is marked for build
//   - BeLaunchLayer() - check if layer is marked for launch
//   - BeCacheLayer() - check if layer is marked for caching
//   - HaveMetadata(key) or HaveMetadata(key, value) - check layer metadata
//
// # Buildpack Metadata
//
// Configure buildpack metadata including configurations and dependencies:
//
//	func TestBuild_WithMetadata(t *testing.T) {
//		result := bptest.NewBuildTest().
//			WithConfigurationMetadata(bptest.ConfigurationMetadata{
//				Name:        "BP_JVM_VERSION",
//				Description: "the Java version",
//				Default:     "17",
//				Build:       true,
//			}).
//			WithDependencyMetadata(bptest.DependencyMetadata{
//				ID:      "jdk",
//				Name:    "BellSoft Liberica JDK",
//				Version: "17.0.1",
//				URI:     "https://example.com/jdk.tar.gz",
//				SHA256:  "abc123...",
//				Stacks:  []string{"*"},
//			}).
//			ExecuteT(t, myBuilder.Build)
//
//		if !result.Succeeded() {
//			t.Fatalf("build failed: %v", result.Error())
//		}
//	}
package bptest
