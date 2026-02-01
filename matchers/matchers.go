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

// Package matchers provides Gomega matchers for bptest result types.
//
// Import this package with dot notation for cleaner test assertions:
//
//	import (
//		. "github.com/onsi/gomega"
//		. "github.com/dmikusa/bptest/matchers"
//	)
//
//	func TestDetect(t *testing.T) {
//		g := NewWithT(t)
//		result := bptest.NewDetectTest().ExecuteT(t, myDetector.Detect)
//		g.Expect(result).To(HavePassed())
//		g.Expect(result).To(HavePlan("jdk"))
//	}
package matchers

import (
	"fmt"

	"github.com/onsi/gomega/types"

	"github.com/dmikusa/bptest"
)

// DetectResult matchers

// HavePassed returns a matcher that checks if detection passed.
func HavePassed() types.GomegaMatcher {
	return &havePassedMatcher{}
}

type havePassedMatcher struct{}

func (m *havePassedMatcher) Match(actual any) (bool, error) {
	result, ok := actual.(*bptest.DetectResult)
	if !ok {
		return false, fmt.Errorf("HavePassed expects a *bptest.DetectResult, got %T", actual)
	}
	return result.Passed(), nil
}

func (m *havePassedMatcher) FailureMessage(actual any) string {
	result := actual.(*bptest.DetectResult)
	if result.HasError() {
		return fmt.Sprintf("Expected detect to pass, but it failed with error: %v", result.Error())
	}
	return "Expected detect to pass, but it failed"
}

func (m *havePassedMatcher) NegatedFailureMessage(actual any) string {
	return "Expected detect not to pass, but it passed"
}

// HaveFailed returns a matcher that checks if detection failed.
func HaveFailed() types.GomegaMatcher {
	return &haveFailedMatcher{}
}

type haveFailedMatcher struct{}

func (m *haveFailedMatcher) Match(actual any) (bool, error) {
	result, ok := actual.(*bptest.DetectResult)
	if !ok {
		return false, fmt.Errorf("HaveFailed expects a *bptest.DetectResult, got %T", actual)
	}
	return result.Failed(), nil
}

func (m *haveFailedMatcher) FailureMessage(actual any) string {
	return "Expected detect to fail, but it passed"
}

func (m *haveFailedMatcher) NegatedFailureMessage(actual any) string {
	result := actual.(*bptest.DetectResult)
	if result.HasError() {
		return fmt.Sprintf("Expected detect not to fail, but it failed with error: %v", result.Error())
	}
	return "Expected detect not to fail, but it failed"
}

// HavePlan returns a matcher that checks if any plan provides or requires the given name.
func HavePlan(name string) types.GomegaMatcher {
	return &havePlanMatcher{name: name}
}

type havePlanMatcher struct {
	name string
}

func (m *havePlanMatcher) Match(actual any) (bool, error) {
	result, ok := actual.(*bptest.DetectResult)
	if !ok {
		return false, fmt.Errorf("HavePlan expects a *bptest.DetectResult, got %T", actual)
	}
	return result.HasPlan(m.name), nil
}

func (m *havePlanMatcher) FailureMessage(actual any) string {
	result := actual.(*bptest.DetectResult)
	return fmt.Sprintf("Expected detect result to have plan %q\n\tProvides: %v\n\tRequires: %v",
		m.name, result.Provides(), result.Requires())
}

func (m *havePlanMatcher) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Expected detect result not to have plan %q, but it did", m.name)
}

// Provide returns a matcher that checks if any plan provides the given name.
func Provide(name string) types.GomegaMatcher {
	return &provideMatcher{name: name}
}

type provideMatcher struct {
	name string
}

func (m *provideMatcher) Match(actual any) (bool, error) {
	result, ok := actual.(*bptest.DetectResult)
	if !ok {
		return false, fmt.Errorf("Provide expects a *bptest.DetectResult, got %T", actual)
	}
	for _, p := range result.Provides() {
		if p == m.name {
			return true, nil
		}
	}
	return false, nil
}

func (m *provideMatcher) FailureMessage(actual any) string {
	result := actual.(*bptest.DetectResult)
	return fmt.Sprintf("Expected detect result to provide %q, but provides: %v", m.name, result.Provides())
}

func (m *provideMatcher) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Expected detect result not to provide %q, but it did", m.name)
}

// Require returns a matcher that checks if any plan requires the given name.
func Require(name string) types.GomegaMatcher {
	return &requireMatcher{name: name}
}

type requireMatcher struct {
	name string
}

func (m *requireMatcher) Match(actual any) (bool, error) {
	result, ok := actual.(*bptest.DetectResult)
	if !ok {
		return false, fmt.Errorf("Require expects a *bptest.DetectResult, got %T", actual)
	}
	for _, r := range result.Requires() {
		if r == m.name {
			return true, nil
		}
	}
	return false, nil
}

func (m *requireMatcher) FailureMessage(actual any) string {
	result := actual.(*bptest.DetectResult)
	return fmt.Sprintf("Expected detect result to require %q, but requires: %v", m.name, result.Requires())
}

func (m *requireMatcher) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Expected detect result not to require %q, but it did", m.name)
}

// BuildResult matchers

// HaveSucceeded returns a matcher that checks if build succeeded.
func HaveSucceeded() types.GomegaMatcher {
	return &haveSucceededMatcher{}
}

type haveSucceededMatcher struct{}

func (m *haveSucceededMatcher) Match(actual any) (bool, error) {
	result, ok := actual.(*bptest.BuildResult)
	if !ok {
		return false, fmt.Errorf("HaveSucceeded expects a *bptest.BuildResult, got %T", actual)
	}
	return result.Succeeded(), nil
}

func (m *haveSucceededMatcher) FailureMessage(actual any) string {
	result := actual.(*bptest.BuildResult)
	if result.HasError() {
		return fmt.Sprintf("Expected build to succeed, but it failed with error: %v", result.Error())
	}
	return "Expected build to succeed, but it failed"
}

func (m *haveSucceededMatcher) NegatedFailureMessage(actual any) string {
	return "Expected build not to succeed, but it succeeded"
}

// HaveLayer returns a matcher that checks if a layer with the given name exists.
func HaveLayer(name string) types.GomegaMatcher {
	return &haveLayerMatcher{name: name}
}

type haveLayerMatcher struct {
	name string
}

func (m *haveLayerMatcher) Match(actual any) (bool, error) {
	result, ok := actual.(*bptest.BuildResult)
	if !ok {
		return false, fmt.Errorf("HaveLayer expects a *bptest.BuildResult, got %T", actual)
	}
	return result.Layer(m.name) != nil, nil
}

func (m *haveLayerMatcher) FailureMessage(actual any) string {
	result := actual.(*bptest.BuildResult)
	return fmt.Sprintf("Expected build result to have layer %q, but layers are: %v", m.name, result.Layers().Names())
}

func (m *haveLayerMatcher) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Expected build result not to have layer %q, but it did", m.name)
}

// HaveProcess returns a matcher that checks if a process with the given type exists.
func HaveProcess(processType string) types.GomegaMatcher {
	return &haveProcessMatcher{processType: processType}
}

type haveProcessMatcher struct {
	processType string
}

func (m *haveProcessMatcher) Match(actual any) (bool, error) {
	result, ok := actual.(*bptest.BuildResult)
	if !ok {
		return false, fmt.Errorf("HaveProcess expects a *bptest.BuildResult, got %T", actual)
	}
	return result.HasProcess(m.processType), nil
}

func (m *haveProcessMatcher) FailureMessage(actual any) string {
	result := actual.(*bptest.BuildResult)
	var types []string
	for _, p := range result.Processes() {
		types = append(types, p.Type)
	}
	return fmt.Sprintf("Expected build result to have process %q, but processes are: %v", m.processType, types)
}

func (m *haveProcessMatcher) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Expected build result not to have process %q, but it did", m.processType)
}

// HaveLabel returns a matcher that checks if a label with the given key exists.
// Optionally checks the value if provided.
func HaveLabel(key string, value ...string) types.GomegaMatcher {
	m := &haveLabelMatcher{key: key}
	if len(value) > 0 {
		m.value = &value[0]
	}
	return m
}

type haveLabelMatcher struct {
	key   string
	value *string
}

func (m *haveLabelMatcher) Match(actual any) (bool, error) {
	result, ok := actual.(*bptest.BuildResult)
	if !ok {
		return false, fmt.Errorf("HaveLabel expects a *bptest.BuildResult, got %T", actual)
	}
	actualValue := result.Label(m.key)
	if actualValue == "" {
		return false, nil
	}
	if m.value != nil {
		return actualValue == *m.value, nil
	}
	return true, nil
}

func (m *haveLabelMatcher) FailureMessage(actual any) string {
	result := actual.(*bptest.BuildResult)
	actualValue := result.Label(m.key)
	if actualValue == "" {
		var keys []string
		for _, l := range result.Labels() {
			keys = append(keys, l.Key)
		}
		return fmt.Sprintf("Expected build result to have label %q, but labels are: %v", m.key, keys)
	}
	// If we reach here, the label exists but value doesn't match
	return fmt.Sprintf("Expected label %q to have value %q, but was %q", m.key, *m.value, actualValue)
}

func (m *haveLabelMatcher) NegatedFailureMessage(actual any) string {
	if m.value != nil {
		return fmt.Sprintf("Expected build result not to have label %q with value %q, but it did", m.key, *m.value)
	}
	return fmt.Sprintf("Expected build result not to have label %q, but it did", m.key)
}
