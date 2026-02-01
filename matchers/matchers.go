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
	"sort"

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

// HaveLayerCount returns a matcher that checks if the build result has exactly n layers.
func HaveLayerCount(n int) types.GomegaMatcher {
	return &haveLayerCountMatcher{count: n}
}

type haveLayerCountMatcher struct {
	count int
}

func (m *haveLayerCountMatcher) Match(actual any) (bool, error) {
	result, ok := actual.(*bptest.BuildResult)
	if !ok {
		return false, fmt.Errorf("HaveLayerCount expects a *bptest.BuildResult, got %T", actual)
	}
	return result.Layers().Count() == m.count, nil
}

func (m *haveLayerCountMatcher) FailureMessage(actual any) string {
	result := actual.(*bptest.BuildResult)
	return fmt.Sprintf("Expected build result to have %d layer(s), but has %d: %v",
		m.count, result.Layers().Count(), result.Layers().Names())
}

func (m *haveLayerCountMatcher) NegatedFailureMessage(actual any) string {
	result := actual.(*bptest.BuildResult)
	return fmt.Sprintf("Expected build result not to have %d layer(s), but it does: %v",
		m.count, result.Layers().Names())
}

// HaveLayers returns a matcher that checks if all specified layers are present.
// Additional layers may also be present.
func HaveLayers(names ...string) types.GomegaMatcher {
	return &haveLayersMatcher{names: names}
}

type haveLayersMatcher struct {
	names   []string
	missing []string
}

func (m *haveLayersMatcher) Match(actual any) (bool, error) {
	result, ok := actual.(*bptest.BuildResult)
	if !ok {
		return false, fmt.Errorf("HaveLayers expects a *bptest.BuildResult, got %T", actual)
	}
	m.missing = nil
	for _, name := range m.names {
		if result.Layer(name) == nil {
			m.missing = append(m.missing, name)
		}
	}
	return len(m.missing) == 0, nil
}

func (m *haveLayersMatcher) FailureMessage(actual any) string {
	result := actual.(*bptest.BuildResult)
	return fmt.Sprintf("Expected build result to have layers %v, but missing: %v (actual layers: %v)",
		m.names, m.missing, result.Layers().Names())
}

func (m *haveLayersMatcher) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Expected build result not to have all layers %v, but it did", m.names)
}

// HaveExactlyLayers returns a matcher that checks if the build result has exactly
// the specified layers and no others.
func HaveExactlyLayers(names ...string) types.GomegaMatcher {
	return &haveExactlyLayersMatcher{names: names}
}

type haveExactlyLayersMatcher struct {
	names   []string
	missing []string
	extra   []string
}

func (m *haveExactlyLayersMatcher) Match(actual any) (bool, error) {
	result, ok := actual.(*bptest.BuildResult)
	if !ok {
		return false, fmt.Errorf("HaveExactlyLayers expects a *bptest.BuildResult, got %T", actual)
	}

	m.missing = nil
	m.extra = nil

	// Check for missing layers
	actualNames := result.Layers().Names()
	actualSet := make(map[string]bool)
	for _, name := range actualNames {
		actualSet[name] = true
	}
	for _, name := range m.names {
		if !actualSet[name] {
			m.missing = append(m.missing, name)
		}
	}

	// Check for extra layers
	expectedSet := make(map[string]bool)
	for _, name := range m.names {
		expectedSet[name] = true
	}
	for _, name := range actualNames {
		if !expectedSet[name] {
			m.extra = append(m.extra, name)
		}
	}

	return len(m.missing) == 0 && len(m.extra) == 0, nil
}

func (m *haveExactlyLayersMatcher) FailureMessage(actual any) string {
	result := actual.(*bptest.BuildResult)
	expected := make([]string, len(m.names))
	copy(expected, m.names)
	sort.Strings(expected)

	if len(m.missing) > 0 && len(m.extra) > 0 {
		return fmt.Sprintf("Expected build result to have exactly layers %v, but missing: %v, extra: %v",
			expected, m.missing, m.extra)
	}
	if len(m.missing) > 0 {
		return fmt.Sprintf("Expected build result to have exactly layers %v, but missing: %v (actual: %v)",
			expected, m.missing, result.Layers().Names())
	}
	return fmt.Sprintf("Expected build result to have exactly layers %v, but has extra: %v",
		expected, m.extra)
}

func (m *haveExactlyLayersMatcher) NegatedFailureMessage(actual any) string {
	expected := make([]string, len(m.names))
	copy(expected, m.names)
	sort.Strings(expected)
	return fmt.Sprintf("Expected build result not to have exactly layers %v, but it did", expected)
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

// LayerResult matchers

// HaveFile returns a matcher that checks if a file exists in the layer.
func HaveFile(relativePath string) types.GomegaMatcher {
	return &haveFileMatcher{path: relativePath}
}

type haveFileMatcher struct {
	path string
}

func (m *haveFileMatcher) Match(actual any) (bool, error) {
	layer, ok := actual.(*bptest.LayerResult)
	if !ok {
		return false, fmt.Errorf("HaveFile expects a *bptest.LayerResult, got %T", actual)
	}
	if layer == nil {
		return false, fmt.Errorf("HaveFile received nil layer")
	}
	return layer.FileExists(m.path), nil
}

func (m *haveFileMatcher) FailureMessage(actual any) string {
	layer := actual.(*bptest.LayerResult)
	return fmt.Sprintf("Expected layer %q to have file %q, but it does not exist", layer.Name(), m.path)
}

func (m *haveFileMatcher) NegatedFailureMessage(actual any) string {
	layer := actual.(*bptest.LayerResult)
	return fmt.Sprintf("Expected layer %q not to have file %q, but it exists", layer.Name(), m.path)
}

// HaveFileWithContents returns a matcher that checks if a file exists with specific contents.
func HaveFileWithContents(relativePath string, contents string) types.GomegaMatcher {
	return &haveFileWithContentsMatcher{path: relativePath, contents: contents}
}

type haveFileWithContentsMatcher struct {
	path        string
	contents    string
	actualValue string
	fileExists  bool
}

func (m *haveFileWithContentsMatcher) Match(actual any) (bool, error) {
	layer, ok := actual.(*bptest.LayerResult)
	if !ok {
		return false, fmt.Errorf("HaveFileWithContents expects a *bptest.LayerResult, got %T", actual)
	}
	if layer == nil {
		return false, fmt.Errorf("HaveFileWithContents received nil layer")
	}
	m.fileExists = layer.FileExists(m.path)
	if !m.fileExists {
		return false, nil
	}
	content, err := layer.FileContentString(m.path)
	if err != nil {
		return false, err
	}
	m.actualValue = content
	return content == m.contents, nil
}

func (m *haveFileWithContentsMatcher) FailureMessage(actual any) string {
	layer := actual.(*bptest.LayerResult)
	if !m.fileExists {
		return fmt.Sprintf("Expected layer %q to have file %q with contents %q, but file does not exist",
			layer.Name(), m.path, m.contents)
	}
	return fmt.Sprintf("Expected layer %q file %q to have contents:\n\t%q\nbut was:\n\t%q",
		layer.Name(), m.path, m.contents, m.actualValue)
}

func (m *haveFileWithContentsMatcher) NegatedFailureMessage(actual any) string {
	layer := actual.(*bptest.LayerResult)
	return fmt.Sprintf("Expected layer %q file %q not to have contents %q, but it did",
		layer.Name(), m.path, m.contents)
}

// BeBuildLayer returns a matcher that checks if a layer is marked for build.
func BeBuildLayer() types.GomegaMatcher {
	return &beBuildLayerMatcher{}
}

type beBuildLayerMatcher struct{}

func (m *beBuildLayerMatcher) Match(actual any) (bool, error) {
	layer, ok := actual.(*bptest.LayerResult)
	if !ok {
		return false, fmt.Errorf("BeBuildLayer expects a *bptest.LayerResult, got %T", actual)
	}
	if layer == nil {
		return false, fmt.Errorf("BeBuildLayer received nil layer")
	}
	return layer.IsBuild(), nil
}

func (m *beBuildLayerMatcher) FailureMessage(actual any) string {
	layer := actual.(*bptest.LayerResult)
	return fmt.Sprintf("Expected layer %q to be a build layer, but it is not", layer.Name())
}

func (m *beBuildLayerMatcher) NegatedFailureMessage(actual any) string {
	layer := actual.(*bptest.LayerResult)
	return fmt.Sprintf("Expected layer %q not to be a build layer, but it is", layer.Name())
}

// BeLaunchLayer returns a matcher that checks if a layer is marked for launch.
func BeLaunchLayer() types.GomegaMatcher {
	return &beLaunchLayerMatcher{}
}

type beLaunchLayerMatcher struct{}

func (m *beLaunchLayerMatcher) Match(actual any) (bool, error) {
	layer, ok := actual.(*bptest.LayerResult)
	if !ok {
		return false, fmt.Errorf("BeLaunchLayer expects a *bptest.LayerResult, got %T", actual)
	}
	if layer == nil {
		return false, fmt.Errorf("BeLaunchLayer received nil layer")
	}
	return layer.IsLaunch(), nil
}

func (m *beLaunchLayerMatcher) FailureMessage(actual any) string {
	layer := actual.(*bptest.LayerResult)
	return fmt.Sprintf("Expected layer %q to be a launch layer, but it is not", layer.Name())
}

func (m *beLaunchLayerMatcher) NegatedFailureMessage(actual any) string {
	layer := actual.(*bptest.LayerResult)
	return fmt.Sprintf("Expected layer %q not to be a launch layer, but it is", layer.Name())
}

// BeCacheLayer returns a matcher that checks if a layer is marked for caching.
func BeCacheLayer() types.GomegaMatcher {
	return &beCacheLayerMatcher{}
}

type beCacheLayerMatcher struct{}

func (m *beCacheLayerMatcher) Match(actual any) (bool, error) {
	layer, ok := actual.(*bptest.LayerResult)
	if !ok {
		return false, fmt.Errorf("BeCacheLayer expects a *bptest.LayerResult, got %T", actual)
	}
	if layer == nil {
		return false, fmt.Errorf("BeCacheLayer received nil layer")
	}
	return layer.IsCache(), nil
}

func (m *beCacheLayerMatcher) FailureMessage(actual any) string {
	layer := actual.(*bptest.LayerResult)
	return fmt.Sprintf("Expected layer %q to be a cache layer, but it is not", layer.Name())
}

func (m *beCacheLayerMatcher) NegatedFailureMessage(actual any) string {
	layer := actual.(*bptest.LayerResult)
	return fmt.Sprintf("Expected layer %q not to be a cache layer, but it is", layer.Name())
}

// HaveMetadata returns a matcher that checks if a layer has metadata with the given key.
// Optionally checks the value if provided.
func HaveMetadata(key string, value ...any) types.GomegaMatcher {
	m := &haveMetadataMatcher{key: key}
	if len(value) > 0 {
		m.value = &value[0]
	}
	return m
}

type haveMetadataMatcher struct {
	key         string
	value       *any
	actualValue any
}

func (m *haveMetadataMatcher) Match(actual any) (bool, error) {
	layer, ok := actual.(*bptest.LayerResult)
	if !ok {
		return false, fmt.Errorf("HaveMetadata expects a *bptest.LayerResult, got %T", actual)
	}
	if layer == nil {
		return false, fmt.Errorf("HaveMetadata received nil layer")
	}
	m.actualValue = layer.MetadataValue(m.key)
	if m.actualValue == nil {
		return false, nil
	}
	if m.value != nil {
		return fmt.Sprintf("%v", m.actualValue) == fmt.Sprintf("%v", *m.value), nil
	}
	return true, nil
}

func (m *haveMetadataMatcher) FailureMessage(actual any) string {
	layer := actual.(*bptest.LayerResult)
	if m.actualValue == nil {
		return fmt.Sprintf("Expected layer %q to have metadata key %q, but it does not", layer.Name(), m.key)
	}
	return fmt.Sprintf("Expected layer %q metadata %q to be %v, but was %v",
		layer.Name(), m.key, *m.value, m.actualValue)
}

func (m *haveMetadataMatcher) NegatedFailureMessage(actual any) string {
	layer := actual.(*bptest.LayerResult)
	if m.value != nil {
		return fmt.Sprintf("Expected layer %q not to have metadata %q with value %v, but it did",
			layer.Name(), m.key, *m.value)
	}
	return fmt.Sprintf("Expected layer %q not to have metadata %q, but it did", layer.Name(), m.key)
}
