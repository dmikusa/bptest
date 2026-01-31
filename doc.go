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
package bptest
