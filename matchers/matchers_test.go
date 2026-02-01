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

package matchers_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb/v2"
	. "github.com/onsi/gomega"

	"github.com/dmikusa/bptest"
	. "github.com/dmikusa/bptest/matchers"
)

func TestHavePassed(t *testing.T) {
	g := NewWithT(t)

	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{Pass: true}, nil
	}

	result := bptest.NewDetectTest().ExecuteT(t, detector)

	g.Expect(result).To(HavePassed())
}

func TestHavePassed_Fails(t *testing.T) {
	g := NewWithT(t)

	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{Pass: false}, nil
	}

	result := bptest.NewDetectTest().ExecuteT(t, detector)

	g.Expect(result).NotTo(HavePassed())
}

func TestHaveFailed(t *testing.T) {
	g := NewWithT(t)

	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{Pass: false}, nil
	}

	result := bptest.NewDetectTest().ExecuteT(t, detector)

	g.Expect(result).To(HaveFailed())
}

func TestHavePlan(t *testing.T) {
	g := NewWithT(t)

	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{
			Pass: true,
			Plans: []libcnb.BuildPlan{
				{
					Provides: []libcnb.BuildPlanProvide{{Name: "jdk"}},
					Requires: []libcnb.BuildPlanRequire{{Name: "jvm"}},
				},
			},
		}, nil
	}

	result := bptest.NewDetectTest().ExecuteT(t, detector)

	g.Expect(result).To(HavePlan("jdk"))
	g.Expect(result).To(HavePlan("jvm"))
	g.Expect(result).NotTo(HavePlan("nodejs"))
}

func TestProvide(t *testing.T) {
	g := NewWithT(t)

	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{
			Pass: true,
			Plans: []libcnb.BuildPlan{
				{
					Provides: []libcnb.BuildPlanProvide{{Name: "jdk"}, {Name: "jre"}},
				},
			},
		}, nil
	}

	result := bptest.NewDetectTest().ExecuteT(t, detector)

	g.Expect(result).To(Provide("jdk"))
	g.Expect(result).To(Provide("jre"))
	g.Expect(result).NotTo(Provide("nodejs"))
}

func TestRequire(t *testing.T) {
	g := NewWithT(t)

	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{
			Pass: true,
			Plans: []libcnb.BuildPlan{
				{
					Requires: []libcnb.BuildPlanRequire{{Name: "jdk"}, {Name: "maven"}},
				},
			},
		}, nil
	}

	result := bptest.NewDetectTest().ExecuteT(t, detector)

	g.Expect(result).To(Require("jdk"))
	g.Expect(result).To(Require("maven"))
	g.Expect(result).NotTo(Require("gradle"))
}

func TestHaveSucceeded(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		return libcnb.NewBuildResult(), nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	g.Expect(result).To(HaveSucceeded())
}

func TestHaveLayer(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, _ := ctx.Layers.Layer("jdk")
		layer.Build = true

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	g.Expect(result).To(HaveLayer("jdk"))
	g.Expect(result).NotTo(HaveLayer("nodejs"))
}

func TestHaveProcess(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		result := libcnb.NewBuildResult()
		result.Processes = []libcnb.Process{
			{Type: "web", Command: []string{"java", "-jar", "app.jar"}},
			{Type: "worker", Command: []string{"java", "Worker"}},
		}
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	g.Expect(result).To(HaveProcess("web"))
	g.Expect(result).To(HaveProcess("worker"))
	g.Expect(result).NotTo(HaveProcess("scheduler"))
}

func TestHaveLabel(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		result := libcnb.NewBuildResult()
		result.Labels = []libcnb.Label{
			{Key: "org.opencontainers.image.title", Value: "My App"},
			{Key: "org.opencontainers.image.version", Value: "1.0.0"},
		}
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	// Check key exists
	g.Expect(result).To(HaveLabel("org.opencontainers.image.title"))
	g.Expect(result).NotTo(HaveLabel("nonexistent"))

	// Check key with value
	g.Expect(result).To(HaveLabel("org.opencontainers.image.title", "My App"))
	g.Expect(result).NotTo(HaveLabel("org.opencontainers.image.title", "Wrong Value"))
	g.Expect(result).To(HaveLabel("org.opencontainers.image.version", "1.0.0"))
}

func TestMatcherTypeErrors(t *testing.T) {
	g := NewWithT(t)

	// Test that matchers return errors for wrong types
	wrongType := "not a result"

	_, err := HavePassed().Match(wrongType)
	g.Expect(err).To(HaveOccurred())

	_, err = HaveFailed().Match(wrongType)
	g.Expect(err).To(HaveOccurred())

	_, err = HavePlan("test").Match(wrongType)
	g.Expect(err).To(HaveOccurred())

	_, err = Provide("test").Match(wrongType)
	g.Expect(err).To(HaveOccurred())

	_, err = Require("test").Match(wrongType)
	g.Expect(err).To(HaveOccurred())

	_, err = HaveSucceeded().Match(wrongType)
	g.Expect(err).To(HaveOccurred())

	_, err = HaveLayer("test").Match(wrongType)
	g.Expect(err).To(HaveOccurred())

	_, err = HaveProcess("test").Match(wrongType)
	g.Expect(err).To(HaveOccurred())

	_, err = HaveLabel("test").Match(wrongType)
	g.Expect(err).To(HaveOccurred())
}

func TestHavePassed_FailureMessages(t *testing.T) {
	g := NewWithT(t)

	// Test failure message when detect fails
	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{Pass: false}, nil
	}
	result := bptest.NewDetectTest().ExecuteT(t, detector)

	matcher := HavePassed()
	msg := matcher.FailureMessage(result)
	g.Expect(msg).To(ContainSubstring("Expected detect to pass"))

	// Test negated failure message when detect passes
	detector = func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{Pass: true}, nil
	}
	result = bptest.NewDetectTest().ExecuteT(t, detector)

	msg = matcher.NegatedFailureMessage(result)
	g.Expect(msg).To(ContainSubstring("Expected detect not to pass"))
}

func TestHavePassed_FailureMessageWithError(t *testing.T) {
	g := NewWithT(t)

	// Test failure message includes error when present
	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{}, fmt.Errorf("something went wrong")
	}
	result := bptest.NewDetectTest().ExecuteT(t, detector)

	matcher := HavePassed()
	msg := matcher.FailureMessage(result)
	g.Expect(msg).To(ContainSubstring("something went wrong"))
}

func TestHaveFailed_FailureMessages(t *testing.T) {
	g := NewWithT(t)

	// Test failure message when detect passes (but we expected failure)
	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{Pass: true}, nil
	}
	result := bptest.NewDetectTest().ExecuteT(t, detector)

	matcher := HaveFailed()
	msg := matcher.FailureMessage(result)
	g.Expect(msg).To(ContainSubstring("Expected detect to fail"))

	// Test negated failure message when detect fails
	detector = func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{Pass: false}, nil
	}
	result = bptest.NewDetectTest().ExecuteT(t, detector)

	msg = matcher.NegatedFailureMessage(result)
	g.Expect(msg).To(ContainSubstring("Expected detect not to fail"))
}

func TestHaveFailed_NegatedFailureMessageWithError(t *testing.T) {
	g := NewWithT(t)

	// Test negated failure message includes error when present
	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{}, fmt.Errorf("detection error")
	}
	result := bptest.NewDetectTest().ExecuteT(t, detector)

	matcher := HaveFailed()
	msg := matcher.NegatedFailureMessage(result)
	g.Expect(msg).To(ContainSubstring("detection error"))
}

func TestHavePlan_FailureMessages(t *testing.T) {
	g := NewWithT(t)

	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{
			Pass: true,
			Plans: []libcnb.BuildPlan{
				{
					Provides: []libcnb.BuildPlanProvide{{Name: "jdk"}},
					Requires: []libcnb.BuildPlanRequire{{Name: "jvm"}},
				},
			},
		}, nil
	}
	result := bptest.NewDetectTest().ExecuteT(t, detector)

	// Test failure message shows provides and requires
	matcher := HavePlan("nodejs")
	msg := matcher.FailureMessage(result)
	g.Expect(msg).To(ContainSubstring("nodejs"))
	g.Expect(msg).To(ContainSubstring("Provides"))
	g.Expect(msg).To(ContainSubstring("Requires"))

	// Test negated failure message
	matcher = HavePlan("jdk")
	msg = matcher.NegatedFailureMessage(result)
	g.Expect(msg).To(ContainSubstring("not to have plan"))
	g.Expect(msg).To(ContainSubstring("jdk"))
}

func TestProvide_FailureMessages(t *testing.T) {
	g := NewWithT(t)

	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{
			Pass: true,
			Plans: []libcnb.BuildPlan{
				{
					Provides: []libcnb.BuildPlanProvide{{Name: "jdk"}, {Name: "jre"}},
				},
			},
		}, nil
	}
	result := bptest.NewDetectTest().ExecuteT(t, detector)

	// Test failure message shows what's provided
	matcher := Provide("nodejs")
	msg := matcher.FailureMessage(result)
	g.Expect(msg).To(ContainSubstring("nodejs"))
	g.Expect(msg).To(ContainSubstring("provides"))

	// Test negated failure message
	matcher = Provide("jdk")
	msg = matcher.NegatedFailureMessage(result)
	g.Expect(msg).To(ContainSubstring("not to provide"))
	g.Expect(msg).To(ContainSubstring("jdk"))
}

func TestRequire_FailureMessages(t *testing.T) {
	g := NewWithT(t)

	detector := func(ctx libcnb.DetectContext) (libcnb.DetectResult, error) {
		return libcnb.DetectResult{
			Pass: true,
			Plans: []libcnb.BuildPlan{
				{
					Requires: []libcnb.BuildPlanRequire{{Name: "jdk"}, {Name: "maven"}},
				},
			},
		}, nil
	}
	result := bptest.NewDetectTest().ExecuteT(t, detector)

	// Test failure message shows what's required
	matcher := Require("gradle")
	msg := matcher.FailureMessage(result)
	g.Expect(msg).To(ContainSubstring("gradle"))
	g.Expect(msg).To(ContainSubstring("requires"))

	// Test negated failure message
	matcher = Require("jdk")
	msg = matcher.NegatedFailureMessage(result)
	g.Expect(msg).To(ContainSubstring("not to require"))
	g.Expect(msg).To(ContainSubstring("jdk"))
}

func TestHaveSucceeded_FailureMessages(t *testing.T) {
	g := NewWithT(t)

	// Test failure message when build fails with error
	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		return libcnb.BuildResult{}, fmt.Errorf("build error")
	}
	result := bptest.NewBuildTest().ExecuteT(t, builder)

	matcher := HaveSucceeded()
	msg := matcher.FailureMessage(result)
	g.Expect(msg).To(ContainSubstring("build error"))

	// Test failure message when build fails without error
	// (This exercises the else branch in FailureMessage)
	builder = func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		return libcnb.BuildResult{}, nil
	}
	result = bptest.NewBuildTest().ExecuteT(t, builder)

	msg = matcher.FailureMessage(result)
	g.Expect(msg).To(ContainSubstring("Expected build to succeed"))

	// Test negated failure message when build succeeds
	builder = func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		return libcnb.NewBuildResult(), nil
	}
	result = bptest.NewBuildTest().ExecuteT(t, builder)

	msg = matcher.NegatedFailureMessage(result)
	g.Expect(msg).To(ContainSubstring("Expected build not to succeed"))
}

func TestHaveLayer_FailureMessages(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, _ := ctx.Layers.Layer("jdk")
		layer.Build = true

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}
	result := bptest.NewBuildTest().ExecuteT(t, builder)

	// Test failure message shows available layers
	matcher := HaveLayer("nodejs")
	msg := matcher.FailureMessage(result)
	g.Expect(msg).To(ContainSubstring("nodejs"))
	g.Expect(msg).To(ContainSubstring("layers are"))

	// Test negated failure message
	matcher = HaveLayer("jdk")
	msg = matcher.NegatedFailureMessage(result)
	g.Expect(msg).To(ContainSubstring("not to have layer"))
	g.Expect(msg).To(ContainSubstring("jdk"))
}

func TestHaveProcess_FailureMessages(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		result := libcnb.NewBuildResult()
		result.Processes = []libcnb.Process{
			{Type: "web", Command: []string{"java", "-jar", "app.jar"}},
		}
		return result, nil
	}
	result := bptest.NewBuildTest().ExecuteT(t, builder)

	// Test failure message shows available processes
	matcher := HaveProcess("worker")
	msg := matcher.FailureMessage(result)
	g.Expect(msg).To(ContainSubstring("worker"))
	g.Expect(msg).To(ContainSubstring("processes are"))

	// Test negated failure message
	matcher = HaveProcess("web")
	msg = matcher.NegatedFailureMessage(result)
	g.Expect(msg).To(ContainSubstring("not to have process"))
	g.Expect(msg).To(ContainSubstring("web"))
}

func TestHaveLabel_FailureMessages(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		result := libcnb.NewBuildResult()
		result.Labels = []libcnb.Label{
			{Key: "org.opencontainers.image.title", Value: "My App"},
		}
		return result, nil
	}
	result := bptest.NewBuildTest().ExecuteT(t, builder)

	// Test failure message when label doesn't exist
	matcher := HaveLabel("nonexistent")
	msg := matcher.FailureMessage(result)
	g.Expect(msg).To(ContainSubstring("nonexistent"))
	g.Expect(msg).To(ContainSubstring("labels are"))

	// Test failure message when value doesn't match
	matcher = HaveLabel("org.opencontainers.image.title", "Wrong Value")
	msg = matcher.FailureMessage(result)
	g.Expect(msg).To(ContainSubstring("Wrong Value"))
	g.Expect(msg).To(ContainSubstring("My App"))

	// Test negated failure message (key only)
	matcher = HaveLabel("org.opencontainers.image.title")
	msg = matcher.NegatedFailureMessage(result)
	g.Expect(msg).To(ContainSubstring("not to have label"))

	// Test negated failure message (key with value)
	matcher = HaveLabel("org.opencontainers.image.title", "My App")
	msg = matcher.NegatedFailureMessage(result)
	g.Expect(msg).To(ContainSubstring("not to have label"))
	g.Expect(msg).To(ContainSubstring("My App"))
}

func TestHaveLayerCount(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer1, _ := ctx.Layers.Layer("jdk")
		layer1.Build = true
		layer2, _ := ctx.Layers.Layer("jre")
		layer2.Launch = true

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer1, layer2)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	g.Expect(result).To(HaveLayerCount(2))
	g.Expect(result).NotTo(HaveLayerCount(1))
	g.Expect(result).NotTo(HaveLayerCount(3))
}

func TestHaveLayerCount_FailureMessages(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, _ := ctx.Layers.Layer("jdk")
		layer.Build = true

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	// Test failure message
	matcher := HaveLayerCount(3)
	msg := matcher.FailureMessage(result)
	g.Expect(msg).To(ContainSubstring("3"))
	g.Expect(msg).To(ContainSubstring("1"))

	// Test negated failure message
	matcher = HaveLayerCount(1)
	msg = matcher.NegatedFailureMessage(result)
	g.Expect(msg).To(ContainSubstring("not to have 1"))
}

func TestHaveLayerCount_TypeErrors(t *testing.T) {
	g := NewWithT(t)

	_, err := HaveLayerCount(1).Match("wrong type")
	g.Expect(err).To(HaveOccurred())
}

func TestHaveLayers(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer1, _ := ctx.Layers.Layer("jdk")
		layer1.Build = true
		layer2, _ := ctx.Layers.Layer("jre")
		layer2.Launch = true
		layer3, _ := ctx.Layers.Layer("maven")
		layer3.Cache = true

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer1, layer2, layer3)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	// All specified layers present
	g.Expect(result).To(HaveLayers("jdk", "jre"))
	g.Expect(result).To(HaveLayers("jdk"))
	g.Expect(result).To(HaveLayers("jdk", "jre", "maven"))

	// Missing layer
	g.Expect(result).NotTo(HaveLayers("nodejs"))
	g.Expect(result).NotTo(HaveLayers("jdk", "nodejs"))
}

func TestHaveLayers_FailureMessages(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, _ := ctx.Layers.Layer("jdk")
		layer.Build = true

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	// Test failure message shows missing layers
	matcher := HaveLayers("jdk", "nodejs", "python")
	matcher.Match(result) // Must call Match first to populate state
	msg := matcher.FailureMessage(result)
	g.Expect(msg).To(ContainSubstring("nodejs"))
	g.Expect(msg).To(ContainSubstring("python"))
	g.Expect(msg).To(ContainSubstring("missing"))

	// Test negated failure message
	matcher = HaveLayers("jdk")
	matcher.Match(result)
	msg = matcher.NegatedFailureMessage(result)
	g.Expect(msg).To(ContainSubstring("not to have all layers"))
}

func TestHaveLayers_TypeErrors(t *testing.T) {
	g := NewWithT(t)

	_, err := HaveLayers("jdk").Match("wrong type")
	g.Expect(err).To(HaveOccurred())
}

func TestHaveExactlyLayers(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer1, _ := ctx.Layers.Layer("jdk")
		layer1.Build = true
		layer2, _ := ctx.Layers.Layer("jre")
		layer2.Launch = true

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer1, layer2)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	// Exactly the right layers (order doesn't matter)
	g.Expect(result).To(HaveExactlyLayers("jdk", "jre"))
	g.Expect(result).To(HaveExactlyLayers("jre", "jdk"))

	// Missing a layer
	g.Expect(result).NotTo(HaveExactlyLayers("jdk"))

	// Extra layer expected
	g.Expect(result).NotTo(HaveExactlyLayers("jdk", "jre", "maven"))
}

func TestHaveExactlyLayers_FailureMessages(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer1, _ := ctx.Layers.Layer("jdk")
		layer1.Build = true
		layer2, _ := ctx.Layers.Layer("jre")
		layer2.Launch = true

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer1, layer2)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	// Test failure message for missing layers
	matcher := HaveExactlyLayers("jdk", "jre", "maven")
	matcher.Match(result) // Must call Match first to populate state
	msg := matcher.FailureMessage(result)
	g.Expect(msg).To(ContainSubstring("maven"))
	g.Expect(msg).To(ContainSubstring("missing"))

	// Test failure message for extra layers
	matcher = HaveExactlyLayers("jdk")
	matcher.Match(result)
	msg = matcher.FailureMessage(result)
	g.Expect(msg).To(ContainSubstring("jre"))
	g.Expect(msg).To(ContainSubstring("extra"))

	// Test failure message for both missing and extra
	matcher = HaveExactlyLayers("jdk", "maven")
	matcher.Match(result)
	msg = matcher.FailureMessage(result)
	g.Expect(msg).To(ContainSubstring("maven"))
	g.Expect(msg).To(ContainSubstring("jre"))

	// Test negated failure message
	matcher = HaveExactlyLayers("jdk", "jre")
	matcher.Match(result)
	msg = matcher.NegatedFailureMessage(result)
	g.Expect(msg).To(ContainSubstring("not to have exactly"))
}

func TestHaveExactlyLayers_TypeErrors(t *testing.T) {
	g := NewWithT(t)

	_, err := HaveExactlyLayers("jdk").Match("wrong type")
	g.Expect(err).To(HaveOccurred())
}

// LayerResult matcher tests

func TestHaveFile(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, _ := ctx.Layers.Layer("jdk")
		layer.Build = true

		// Write a file to the layer
		if err := os.MkdirAll(filepath.Join(layer.Path, "bin"), 0755); err != nil {
			return libcnb.BuildResult{}, err
		}
		if err := os.WriteFile(filepath.Join(layer.Path, "bin/java"), []byte("java binary"), 0755); err != nil {
			return libcnb.BuildResult{}, err
		}

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)
	layer := result.Layer("jdk")

	g.Expect(layer).To(HaveFile("bin/java"))
	g.Expect(layer).NotTo(HaveFile("bin/python"))
}

func TestHaveFile_FailureMessages(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, _ := ctx.Layers.Layer("jdk")
		layer.Build = true

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)
	layer := result.Layer("jdk")

	// Test failure message
	matcher := HaveFile("bin/java")
	matcher.Match(layer)
	msg := matcher.FailureMessage(layer)
	g.Expect(msg).To(ContainSubstring("jdk"))
	g.Expect(msg).To(ContainSubstring("bin/java"))
	g.Expect(msg).To(ContainSubstring("does not exist"))

	// Create file for negated test
	os.MkdirAll(filepath.Join(layer.Path(), "bin"), 0755)
	os.WriteFile(filepath.Join(layer.Path(), "bin/java"), []byte(""), 0755)

	// Test negated failure message
	matcher.Match(layer)
	msg = matcher.NegatedFailureMessage(layer)
	g.Expect(msg).To(ContainSubstring("not to have file"))
}

func TestHaveFile_TypeErrors(t *testing.T) {
	g := NewWithT(t)

	_, err := HaveFile("test").Match("wrong type")
	g.Expect(err).To(HaveOccurred())

	// Test nil layer
	var nilLayer *bptest.LayerResult
	_, err = HaveFile("test").Match(nilLayer)
	g.Expect(err).To(HaveOccurred())
}

func TestHaveFileWithContents(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, _ := ctx.Layers.Layer("jdk")
		layer.Build = true

		// Create the layer directory and write a file with specific contents
		if err := os.MkdirAll(layer.Path, 0755); err != nil {
			return libcnb.BuildResult{}, err
		}
		if err := os.WriteFile(filepath.Join(layer.Path, "version"), []byte("17.0.1"), 0644); err != nil {
			return libcnb.BuildResult{}, err
		}

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)
	layer := result.Layer("jdk")

	g.Expect(layer).To(HaveFileWithContents("version", "17.0.1"))
	g.Expect(layer).NotTo(HaveFileWithContents("version", "11.0.0"))
	g.Expect(layer).NotTo(HaveFileWithContents("nonexistent", "anything"))
}

func TestHaveFileWithContents_FailureMessages(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, _ := ctx.Layers.Layer("jdk")
		layer.Build = true

		// Create the layer directory and write file
		os.MkdirAll(layer.Path, 0755)
		os.WriteFile(filepath.Join(layer.Path, "version"), []byte("17.0.1"), 0644)

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)
	layer := result.Layer("jdk")

	// Test failure message for file not found
	matcher := HaveFileWithContents("nonexistent", "content")
	matcher.Match(layer)
	msg := matcher.FailureMessage(layer)
	g.Expect(msg).To(ContainSubstring("does not exist"))

	// Test failure message for wrong contents
	matcher = HaveFileWithContents("version", "11.0.0")
	matcher.Match(layer)
	msg = matcher.FailureMessage(layer)
	g.Expect(msg).To(ContainSubstring("11.0.0"))
	g.Expect(msg).To(ContainSubstring("17.0.1"))

	// Test negated failure message
	matcher = HaveFileWithContents("version", "17.0.1")
	matcher.Match(layer)
	msg = matcher.NegatedFailureMessage(layer)
	g.Expect(msg).To(ContainSubstring("not to have contents"))
}

func TestHaveFileWithContents_TypeErrors(t *testing.T) {
	g := NewWithT(t)

	_, err := HaveFileWithContents("test", "content").Match("wrong type")
	g.Expect(err).To(HaveOccurred())

	// Test nil layer
	var nilLayer *bptest.LayerResult
	_, err = HaveFileWithContents("test", "content").Match(nilLayer)
	g.Expect(err).To(HaveOccurred())
}

func TestBeBuildLayer(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		buildLayer, _ := ctx.Layers.Layer("build-layer")
		buildLayer.Build = true

		launchLayer, _ := ctx.Layers.Layer("launch-layer")
		launchLayer.Launch = true

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, buildLayer, launchLayer)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	g.Expect(result.Layer("build-layer")).To(BeBuildLayer())
	g.Expect(result.Layer("launch-layer")).NotTo(BeBuildLayer())
}

func TestBeBuildLayer_FailureMessages(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, _ := ctx.Layers.Layer("jdk")
		layer.Launch = true

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)
	layer := result.Layer("jdk")

	matcher := BeBuildLayer()
	matcher.Match(layer)
	msg := matcher.FailureMessage(layer)
	g.Expect(msg).To(ContainSubstring("jdk"))
	g.Expect(msg).To(ContainSubstring("build layer"))

	// For negated, we need a build layer
	builder = func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, _ := ctx.Layers.Layer("jdk")
		layer.Build = true

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}
	result = bptest.NewBuildTest().ExecuteT(t, builder)
	layer = result.Layer("jdk")

	matcher.Match(layer)
	msg = matcher.NegatedFailureMessage(layer)
	g.Expect(msg).To(ContainSubstring("not to be a build layer"))
}

func TestBeBuildLayer_TypeErrors(t *testing.T) {
	g := NewWithT(t)

	_, err := BeBuildLayer().Match("wrong type")
	g.Expect(err).To(HaveOccurred())

	// Test nil layer
	var nilLayer *bptest.LayerResult
	_, err = BeBuildLayer().Match(nilLayer)
	g.Expect(err).To(HaveOccurred())
}

func TestBeLaunchLayer(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		launchLayer, _ := ctx.Layers.Layer("launch-layer")
		launchLayer.Launch = true

		buildLayer, _ := ctx.Layers.Layer("build-layer")
		buildLayer.Build = true

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, launchLayer, buildLayer)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	g.Expect(result.Layer("launch-layer")).To(BeLaunchLayer())
	g.Expect(result.Layer("build-layer")).NotTo(BeLaunchLayer())
}

func TestBeLaunchLayer_FailureMessages(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, _ := ctx.Layers.Layer("jdk")
		layer.Build = true

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)
	layer := result.Layer("jdk")

	matcher := BeLaunchLayer()
	matcher.Match(layer)
	msg := matcher.FailureMessage(layer)
	g.Expect(msg).To(ContainSubstring("launch layer"))

	// For negated, we need a launch layer
	builder = func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, _ := ctx.Layers.Layer("jdk")
		layer.Launch = true

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}
	result = bptest.NewBuildTest().ExecuteT(t, builder)
	layer = result.Layer("jdk")

	matcher.Match(layer)
	msg = matcher.NegatedFailureMessage(layer)
	g.Expect(msg).To(ContainSubstring("not to be a launch layer"))
}

func TestBeLaunchLayer_TypeErrors(t *testing.T) {
	g := NewWithT(t)

	_, err := BeLaunchLayer().Match("wrong type")
	g.Expect(err).To(HaveOccurred())

	// Test nil layer
	var nilLayer *bptest.LayerResult
	_, err = BeLaunchLayer().Match(nilLayer)
	g.Expect(err).To(HaveOccurred())
}

func TestBeCacheLayer(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		cacheLayer, _ := ctx.Layers.Layer("cache-layer")
		cacheLayer.Cache = true

		buildLayer, _ := ctx.Layers.Layer("build-layer")
		buildLayer.Build = true

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, cacheLayer, buildLayer)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)

	g.Expect(result.Layer("cache-layer")).To(BeCacheLayer())
	g.Expect(result.Layer("build-layer")).NotTo(BeCacheLayer())
}

func TestBeCacheLayer_FailureMessages(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, _ := ctx.Layers.Layer("jdk")
		layer.Build = true

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)
	layer := result.Layer("jdk")

	matcher := BeCacheLayer()
	matcher.Match(layer)
	msg := matcher.FailureMessage(layer)
	g.Expect(msg).To(ContainSubstring("cache layer"))

	// For negated, we need a cache layer
	builder = func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, _ := ctx.Layers.Layer("jdk")
		layer.Cache = true

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}
	result = bptest.NewBuildTest().ExecuteT(t, builder)
	layer = result.Layer("jdk")

	matcher.Match(layer)
	msg = matcher.NegatedFailureMessage(layer)
	g.Expect(msg).To(ContainSubstring("not to be a cache layer"))
}

func TestBeCacheLayer_TypeErrors(t *testing.T) {
	g := NewWithT(t)

	_, err := BeCacheLayer().Match("wrong type")
	g.Expect(err).To(HaveOccurred())

	// Test nil layer
	var nilLayer *bptest.LayerResult
	_, err = BeCacheLayer().Match(nilLayer)
	g.Expect(err).To(HaveOccurred())
}

func TestHaveMetadata(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, _ := ctx.Layers.Layer("jdk")
		layer.Build = true
		layer.Metadata = map[string]any{
			"version": "17",
			"vendor":  "Eclipse Adoptium",
		}

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)
	layer := result.Layer("jdk")

	// Check key exists
	g.Expect(layer).To(HaveMetadata("version"))
	g.Expect(layer).To(HaveMetadata("vendor"))
	g.Expect(layer).NotTo(HaveMetadata("nonexistent"))

	// Check key with value
	g.Expect(layer).To(HaveMetadata("version", "17"))
	g.Expect(layer).NotTo(HaveMetadata("version", "11"))
}

func TestHaveMetadata_FailureMessages(t *testing.T) {
	g := NewWithT(t)

	builder := func(ctx libcnb.BuildContext) (libcnb.BuildResult, error) {
		layer, _ := ctx.Layers.Layer("jdk")
		layer.Build = true
		layer.Metadata = map[string]any{
			"version": "17",
		}

		result := libcnb.NewBuildResult()
		result.Layers = append(result.Layers, layer)
		return result, nil
	}

	result := bptest.NewBuildTest().ExecuteT(t, builder)
	layer := result.Layer("jdk")

	// Test failure message for missing key
	matcher := HaveMetadata("nonexistent")
	matcher.Match(layer)
	msg := matcher.FailureMessage(layer)
	g.Expect(msg).To(ContainSubstring("nonexistent"))
	g.Expect(msg).To(ContainSubstring("does not"))

	// Test failure message for wrong value
	matcher = HaveMetadata("version", "11")
	matcher.Match(layer)
	msg = matcher.FailureMessage(layer)
	g.Expect(msg).To(ContainSubstring("11"))
	g.Expect(msg).To(ContainSubstring("17"))

	// Test negated failure message (key only)
	matcher = HaveMetadata("version")
	matcher.Match(layer)
	msg = matcher.NegatedFailureMessage(layer)
	g.Expect(msg).To(ContainSubstring("not to have metadata"))

	// Test negated failure message (key with value)
	matcher = HaveMetadata("version", "17")
	matcher.Match(layer)
	msg = matcher.NegatedFailureMessage(layer)
	g.Expect(msg).To(ContainSubstring("not to have metadata"))
	g.Expect(msg).To(ContainSubstring("17"))
}

func TestHaveMetadata_TypeErrors(t *testing.T) {
	g := NewWithT(t)

	_, err := HaveMetadata("key").Match("wrong type")
	g.Expect(err).To(HaveOccurred())

	// Test nil layer
	var nilLayer *bptest.LayerResult
	_, err = HaveMetadata("key").Match(nilLayer)
	g.Expect(err).To(HaveOccurred())
}
