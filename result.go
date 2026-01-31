package bptest

import (
	"github.com/buildpacks/libcnb/v2"
)

// DetectResult wraps the result of a detect function execution.
type DetectResult struct {
	result libcnb.DetectResult
	err    error
}

// Passed returns true if detection passed.
func (r *DetectResult) Passed() bool {
	return r.err == nil && r.result.Pass
}

// Failed returns true if detection failed (did not pass).
func (r *DetectResult) Failed() bool {
	return !r.Passed()
}

// HasError returns true if the detect function returned an error.
func (r *DetectResult) HasError() bool {
	return r.err != nil
}

// Error returns the error from the detect function, if any.
func (r *DetectResult) Error() error {
	return r.err
}

// Plans returns all build plans from the detect result.
func (r *DetectResult) Plans() []libcnb.BuildPlan {
	return r.result.Plans
}

// HasPlan returns true if any plan provides or requires the given name.
func (r *DetectResult) HasPlan(name string) bool {
	for _, plan := range r.result.Plans {
		for _, p := range plan.Provides {
			if p.Name == name {
				return true
			}
		}
		for _, req := range plan.Requires {
			if req.Name == name {
				return true
			}
		}
	}
	return false
}

// Provides returns all provided dependency names across all plans.
func (r *DetectResult) Provides() []string {
	seen := make(map[string]bool)
	var result []string
	for _, plan := range r.result.Plans {
		for _, p := range plan.Provides {
			if !seen[p.Name] {
				seen[p.Name] = true
				result = append(result, p.Name)
			}
		}
	}
	return result
}

// Requires returns all required dependency names across all plans.
func (r *DetectResult) Requires() []string {
	seen := make(map[string]bool)
	var result []string
	for _, plan := range r.result.Plans {
		for _, req := range plan.Requires {
			if !seen[req.Name] {
				seen[req.Name] = true
				result = append(result, req.Name)
			}
		}
	}
	return result
}

// RequireMetadata returns the metadata for the first requirement with the given name.
func (r *DetectResult) RequireMetadata(name string) map[string]any {
	for _, plan := range r.result.Plans {
		for _, req := range plan.Requires {
			if req.Name == name {
				return req.Metadata
			}
		}
	}
	return nil
}

// BuildResult wraps the result of a build function execution.
type BuildResult struct {
	result    libcnb.BuildResult
	err       error
	layersDir string
}

// Succeeded returns true if the build completed without error.
func (r *BuildResult) Succeeded() bool {
	return r.err == nil
}

// HasError returns true if the build function returned an error.
func (r *BuildResult) HasError() bool {
	return r.err != nil
}

// Error returns the error from the build function, if any.
func (r *BuildResult) Error() error {
	return r.err
}

// Layers returns a LayerInspector for examining the contributed layers.
func (r *BuildResult) Layers() *LayerInspector {
	return &LayerInspector{
		layers:    r.result.Layers,
		layersDir: r.layersDir,
	}
}

// Layer returns a LayerResult for the layer with the given name.
// Returns nil if the layer doesn't exist.
func (r *BuildResult) Layer(name string) *LayerResult {
	for _, layer := range r.result.Layers {
		if layer.Name == name {
			return &LayerResult{layer: layer}
		}
	}
	return nil
}

// Processes returns all processes contributed by the build.
func (r *BuildResult) Processes() []libcnb.Process {
	return r.result.Processes
}

// HasProcess returns true if a process with the given type exists.
func (r *BuildResult) HasProcess(processType string) bool {
	for _, p := range r.result.Processes {
		if p.Type == processType {
			return true
		}
	}
	return false
}

// Process returns the process with the given type, or nil if not found.
func (r *BuildResult) Process(processType string) *libcnb.Process {
	for _, p := range r.result.Processes {
		if p.Type == processType {
			return &p
		}
	}
	return nil
}

// DefaultProcess returns the default process, or nil if none is marked default.
func (r *BuildResult) DefaultProcess() *libcnb.Process {
	for _, p := range r.result.Processes {
		if p.Default {
			return &p
		}
	}
	return nil
}

// Labels returns all labels contributed by the build.
func (r *BuildResult) Labels() []libcnb.Label {
	return r.result.Labels
}

// Label returns the value of the label with the given key, or empty string if not found.
func (r *BuildResult) Label(key string) string {
	for _, l := range r.result.Labels {
		if l.Key == key {
			return l.Value
		}
	}
	return ""
}

// Slices returns all slices contributed by the build.
func (r *BuildResult) Slices() []libcnb.Slice {
	return r.result.Slices
}

// UnmetEntries returns all unmet plan entries.
func (r *BuildResult) UnmetEntries() []libcnb.UnmetPlanEntry {
	return r.result.Unmet
}

// PersistentMetadata returns the persistent metadata from the build result.
func (r *BuildResult) PersistentMetadata() map[string]any {
	return r.result.PersistentMetadata
}
