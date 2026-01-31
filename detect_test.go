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

func TestDetectTest_BasicPass(t *testing.T) {
	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{Pass: true}, nil
	}

	result := bptest.NewDetectTest().ExecuteT(t, detector)

	if !result.Passed() {
		t.Error("expected detect to pass")
	}
	if result.Failed() {
		t.Error("expected Failed() to be false")
	}
	if result.HasError() {
		t.Errorf("unexpected error: %v", result.Error())
	}
}

func TestDetectTest_BasicFail(t *testing.T) {
	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{Pass: false}, nil
	}

	result := bptest.NewDetectTest().ExecuteT(t, detector)

	if result.Passed() {
		t.Error("expected detect to fail")
	}
	if !result.Failed() {
		t.Error("expected Failed() to be true")
	}
}

func TestDetectTest_WithError(t *testing.T) {
	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{}, os.ErrNotExist
	}

	result := bptest.NewDetectTest().ExecuteT(t, detector)

	if !result.HasError() {
		t.Error("expected HasError() to be true")
	}
	if result.Error() != os.ErrNotExist {
		t.Errorf("expected os.ErrNotExist, got %v", result.Error())
	}
	if result.Passed() {
		t.Error("expected Passed() to be false when there's an error")
	}
}

func TestDetectTest_WithAppFile(t *testing.T) {
	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		pomPath := filepath.Join(ctx.ApplicationPath, "pom.xml")
		if _, err := os.Stat(pomPath); os.IsNotExist(err) {
			return libcnb.DetectResult{Pass: false}, nil
		}
		return libcnb.DetectResult{Pass: true}, nil
	}

	result := bptest.NewDetectTest().
		WithAppFileString("pom.xml", "<project></project>").
		ExecuteT(t, detector)

	if !result.Passed() {
		t.Error("expected detect to pass when pom.xml exists")
	}
}

func TestDetectTest_WithPlatformEnv(t *testing.T) {
	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		if ctx.Platform.Environment["BP_JAVA_VERSION"] != "17" {
			return libcnb.DetectResult{Pass: false}, nil
		}
		return libcnb.DetectResult{Pass: true}, nil
	}

	result := bptest.NewDetectTest().
		WithPlatformEnv("BP_JAVA_VERSION", "17").
		ExecuteT(t, detector)

	if !result.Passed() {
		t.Error("expected detect to pass with correct platform env")
	}
}

func TestDetectTest_WithBinding(t *testing.T) {
	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		for _, b := range ctx.Platform.Bindings {
			if b.Name == "my-db" && b.Type == "postgresql" {
				if b.Secret["password"] == "secret123" {
					return libcnb.DetectResult{Pass: true}, nil
				}
			}
		}
		return libcnb.DetectResult{Pass: false}, nil
	}

	result := bptest.NewDetectTest().
		WithBinding(bptest.Binding{
			Name:   "my-db",
			Type:   "postgresql",
			Secret: map[string]string{"password": "secret123"},
		}).
		ExecuteT(t, detector)

	if !result.Passed() {
		t.Error("expected detect to pass with binding")
	}
}

func TestDetectTest_WithBuildPlan(t *testing.T) {
	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{
			Pass: true,
			Plans: []libcnb.BuildPlan{
				{
					Provides: []libcnb.BuildPlanProvide{{Name: "jdk"}},
					Requires: []libcnb.BuildPlanRequire{
						{Name: "jdk", Metadata: map[string]any{"version": "17"}},
						{Name: "jre"},
					},
				},
			},
		}, nil
	}

	result := bptest.NewDetectTest().ExecuteT(t, detector)

	if !result.Passed() {
		t.Error("expected detect to pass")
	}
	if !result.HasPlan("jdk") {
		t.Error("expected to have jdk plan")
	}
	if !result.HasPlan("jre") {
		t.Error("expected to have jre plan")
	}
	if result.HasPlan("nonexistent") {
		t.Error("expected nonexistent plan to not be found")
	}

	provides := result.Provides()
	if len(provides) != 1 || provides[0] != "jdk" {
		t.Errorf("expected provides [jdk], got %v", provides)
	}

	requires := result.Requires()
	if len(requires) != 2 {
		t.Errorf("expected 2 requires, got %d", len(requires))
	}

	metadata := result.RequireMetadata("jdk")
	if metadata["version"] != "17" {
		t.Errorf("expected jdk version 17, got %v", metadata["version"])
	}

	if result.RequireMetadata("nonexistent") != nil {
		t.Error("expected nil metadata for nonexistent requirement")
	}
}

func TestDetectTest_WithBuildpackConfig(t *testing.T) {
	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		if ctx.Buildpack.Info.ID != "example/java" {
			return libcnb.DetectResult{Pass: false}, nil
		}
		if ctx.Buildpack.Info.Version != "1.2.3" {
			return libcnb.DetectResult{Pass: false}, nil
		}
		return libcnb.DetectResult{Pass: true}, nil
	}

	result := bptest.NewDetectTest().
		WithBuildpack(bptest.BuildpackConfig{
			API:     "0.8",
			ID:      "example/java",
			Version: "1.2.3",
		}).
		ExecuteT(t, detector)

	if !result.Passed() {
		t.Error("expected detect to pass with custom buildpack config")
	}
}

func TestDetectTest_Cleanup(t *testing.T) {
	dt := bptest.NewDetectTest().
		WithAppFileString("test.txt", "hello")

	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{Pass: true}, nil
	}

	result, err := dt.Execute(detector)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if !result.Passed() {
		t.Error("expected pass")
	}

	paths := dt.Paths()
	if _, err := os.Stat(paths.AppDir); os.IsNotExist(err) {
		t.Error("app dir should exist before cleanup")
	}

	if err := dt.Cleanup(); err != nil {
		t.Errorf("cleanup failed: %v", err)
	}

	if _, err := os.Stat(paths.AppDir); !os.IsNotExist(err) {
		t.Error("app dir should not exist after cleanup")
	}
}

func TestDetectTest_WithTarget(t *testing.T) {
	// Note: DetectContext doesn't have TargetInfo in the same way BuildContext does
	// but we test that WithTarget doesn't break anything and the test completes
	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{Pass: true}, nil
	}

	result := bptest.NewDetectTest().
		WithTarget(bptest.TargetInfo{
			OS:            "linux",
			Arch:          "amd64",
			DistroName:    "ubuntu",
			DistroVersion: "22.04",
		}).
		ExecuteT(t, detector)

	if !result.Passed() {
		t.Error("expected detect to pass with target info")
	}
}

func TestDetectTest_WithStackID(t *testing.T) {
	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		if ctx.StackID != "io.buildpacks.stacks.jammy" {
			t.Errorf("expected stack ID io.buildpacks.stacks.jammy, got %s", ctx.StackID)
		}
		return libcnb.DetectResult{Pass: true}, nil
	}

	result := bptest.NewDetectTest().
		WithStackID("io.buildpacks.stacks.jammy").
		ExecuteT(t, detector)

	if !result.Passed() {
		t.Error("expected detect to pass with stack ID")
	}
}
