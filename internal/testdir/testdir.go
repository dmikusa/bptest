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

// Package testdir provides temporary directory management for tests.
package testdir

import (
	"os"
	"path/filepath"
)

// Manager tracks temporary directories and provides helper methods for
// creating files within them.
type Manager struct {
	dirs []string
}

// NewManager creates a new Manager.
func NewManager() *Manager {
	return &Manager{}
}

// CreateTempDir creates a temporary directory with the given prefix and
// tracks it for cleanup.
func (m *Manager) CreateTempDir(prefix string) (string, error) {
	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		return "", err
	}
	m.dirs = append(m.dirs, dir)
	return dir, nil
}

// WriteFile writes content to a file at relativePath within baseDir,
// creating any necessary parent directories.
func (m *Manager) WriteFile(baseDir, relativePath string, content []byte) error {
	fullPath := filepath.Join(baseDir, relativePath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, content, 0644)
}

// WriteFileString is a convenience method that writes a string to a file.
func (m *Manager) WriteFileString(baseDir, relativePath, content string) error {
	return m.WriteFile(baseDir, relativePath, []byte(content))
}

// Cleanup removes all tracked temporary directories.
func (m *Manager) Cleanup() error {
	var firstErr error
	for _, dir := range m.dirs {
		if err := os.RemoveAll(dir); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	m.dirs = nil
	return firstErr
}

// Dirs returns the list of tracked directories.
func (m *Manager) Dirs() []string {
	return m.dirs
}
