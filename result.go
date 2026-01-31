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
