// Package bptest provides testing utilities for Cloud Native Buildpacks.
//
// This library allows you to test buildpack detect and build functions directly
// by invoking them with configurable fake data, without needing to run the full
// buildpack lifecycle or spawn subprocesses.
//
// # Detect Tests
//
// Use DetectTest to test your buildpack's detect function:
//
//	result := bptest.NewDetectTest().
//		WithAppFileString("pom.xml", `<project>...</project>`).
//		WithBuildpack(bptest.BuildpackConfig{
//			API:     "0.8",
//			ID:      "example/java-maven",
//			Version: "1.0.0",
//		}).
//		WithPlatformEnv("BP_JAVA_VERSION", "17").
//		ExecuteT(t, myDetector.Detect)
//
//	if !result.Passed() {
//		t.Fatal("detect should have passed")
//	}
//	if !result.HasPlan("jdk") {
//		t.Error("should require jdk")
//	}
//
// # Build Tests
//
// Use BuildTest to test your buildpack's build function:
//
//	result := bptest.NewBuildTest().
//		WithAppFileString("pom.xml", `<project>...</project>`).
//		WithPlanEntry("jdk", map[string]interface{}{"version": "17"}).
//		ExecuteT(t, myBuilder.Build)
//
//	if !result.Succeeded() {
//		t.Fatal(result.Error())
//	}
//	layer := result.Layer("jdk")
//	if !layer.IsBuild() || !layer.IsCache() {
//		t.Error("wrong layer types")
//	}
//
// # Service Bindings
//
// Both DetectTest and BuildTest support service bindings:
//
//	result := bptest.NewDetectTest().
//		WithBinding(bptest.Binding{
//			Name:   "my-db",
//			Type:   "postgresql",
//			Secret: map[string]string{"password": "secret"},
//		}).
//		ExecuteT(t, myDetector.Detect)
//
// # Rebuild Scenarios
//
// Test rebuild scenarios by providing existing layer state:
//
//	result := bptest.NewBuildTest().
//		WithExistingLayer("jdk", &bptest.LayerConfig{
//			Build:    true,
//			Cache:    true,
//			Metadata: map[string]interface{}{"version": "17"},
//		}).
//		ExecuteT(t, myBuilder.Build)
package bptest
